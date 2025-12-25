package integrations

import (
	"net/url"
	"os"
	"time"

	"numerra-backend/config"
	"numerra-backend/models"
	"numerra-backend/providers"
	"numerra-backend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// GET /v1/integrations/status
func GetIntegrationStatus(c *fiber.Ctx) error {
	orgID, err := selectedOrgID(c)
	if err != nil {
		return err
	}

	var rows []models.OrganizationIntegration
	if err := config.DB.Where("organization_id = ?", orgID).Find(&rows).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "DB error"})
	}

	type Resp struct {
		Provider          string     `json:"provider"`
		Status            string     `json:"status"`
		ExternalAccountID *string    `json:"externalAccountId"`
		AccessExpiresAt   *time.Time `json:"accessExpiresAt"`
		RefreshExpiresAt  *time.Time `json:"refreshExpiresAt"`
	}

	out := make([]Resp, 0, len(rows))
	for _, r := range rows {
		out = append(out, Resp{
			Provider:          r.Provider,
			Status:            string(r.Status),
			ExternalAccountID: r.ExternalAccountID,
			AccessExpiresAt:   r.AccessExpiresAt,
			RefreshExpiresAt:  r.RefreshExpiresAt,
		})
	}

	return c.JSON(fiber.Map{
		"organizationId": orgID,
		"integrations":   out,
	})
}

// POST /v1/integrations/connect
func ConnectIntegration(c *fiber.Ctx) error {
	if err := requireOwnerAdmin(c); err != nil {
		return err
	}
	orgID, err := selectedOrgID(c)
	if err != nil {
		return err
	}
	user, err2 := currentUser(c)
	if err2 != nil {
		return err2
	}

	type Body struct {
		Provider string `json:"provider"`
		Shop     string `json:"shop,omitempty"`
	}
	var body Body
	if err := c.BodyParser(&body); err != nil || body.Provider == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	driver, derr := providers.GetDriver(body.Provider, body.Shop)
	if derr != nil {
		return c.Status(400).JSON(fiber.Map{"error": derr.Error()})
	}

	state := uuid.New().String()

	st := models.OAuthState{
		OrganizationID:  orgID,
		Provider:        body.Provider,
		State:           state,
		CreatedByUserID: user.ID,
		ExpiresAt:       time.Now().Add(10 * time.Minute),
	}
	if err := config.DB.Create(&st).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed saving oauth state"})
	}

	authURL, err := driver.BuildAuthURL(state)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed building auth url"})
	}

	return c.JSON(fiber.Map{"authUrl": authURL})
}

// GET /v1/integrations/callback/:provider?code=...&state=... (&realmId=... for QBO, &shop=...&hmac=... for Shopify)
func OAuthCallback(c *fiber.Ctx) error {
	providerName := c.Params("provider")
	code := c.Query("code")
	state := c.Query("state")

	if providerName == "" || code == "" || state == "" {
		return c.Status(400).SendString("Missing params")
	}

	if providerName == "shopify" {
		secret := os.Getenv("SHOPIFY_API_SECRET")
		if secret == "" {
			return c.Status(500).SendString("Server not configured")
		}
		if !utils.VerifyShopifyHMACFromFiberQuery(c, secret) {
			return c.Status(400).SendString("Invalid Shopify HMAC")
		}
	}

	var st models.OAuthState
	if err := config.DB.Where("state = ? AND provider = ?", state, providerName).First(&st).Error; err != nil {
		return c.Status(400).SendString("Invalid state")
	}
	if time.Now().After(st.ExpiresAt) {
		return c.Status(400).SendString("State expired")
	}

	extra := map[string]string{
		"realmId": c.Query("realmId"),
		"shop":    c.Query("shop"),
	}
	driver, derr := providers.GetDriver(providerName, extra["shop"])
	if derr != nil {
		return c.Status(400).SendString("Unknown provider")
	}

	tokenRes, err := driver.ExchangeCode(code, state, extra)
	if err != nil {
		return c.Status(500).SendString("Token exchange failed")
	}

	encKey := os.Getenv("CONNECTOR_ENCRYPTION_KEY")
	if encKey == "" {
		return c.Status(500).SendString("Missing CONNECTOR_ENCRYPTION_KEY")
	}

	accessEnc, err := utils.EncryptV2(tokenRes.AccessToken, encKey)
	if err != nil {
		return c.Status(500).SendString("Encryption failed")
	}

	var refreshEnc *string
	if tokenRes.RefreshToken != nil && *tokenRes.RefreshToken != "" {
		rEnc, err := utils.EncryptV2(*tokenRes.RefreshToken, encKey)
		if err != nil {
			return c.Status(500).SendString("Encryption failed")
		}
		refreshEnc = &rEnc
	}

	now := time.Now()
	upd := map[string]interface{}{
		"status":               "CONNECTED",
		"external_account_id":  tokenRes.ExternalAccountID,
		"scopes":               tokenRes.Scopes,
		"access_token_enc":     &accessEnc,
		"refresh_token_enc":    refreshEnc,
		"access_expires_at":    tokenRes.AccessExpiresAt,
		"refresh_expires_at":   tokenRes.RefreshExpiresAt,
		"connected_by_user_id": &st.CreatedByUserID,
		"connected_at":         &now,
		"last_refreshed_at":    &now,
	}

	var row models.OrganizationIntegration
	errFind := config.DB.Where("organization_id = ? AND provider = ?", st.OrganizationID, providerName).First(&row).Error
	if errFind == nil {
		if err := config.DB.Model(&row).Updates(upd).Error; err != nil {
			return c.Status(500).SendString("DB update failed")
		}
	} else if errFind == gorm.ErrRecordNotFound {
		row = models.OrganizationIntegration{
			OrganizationID: st.OrganizationID,
			Provider:       providerName,
			Status:         models.IntegrationConnected,
		}
		if err := config.DB.Create(&row).Error; err != nil {
			return c.Status(500).SendString("DB insert failed")
		}
		if err := config.DB.Model(&row).Updates(upd).Error; err != nil {
			return c.Status(500).SendString("DB update failed")
		}
	} else {
		return c.Status(500).SendString("DB error")
	}

	config.DB.Delete(&st)

	front := os.Getenv("FRONTEND_BASE_URL")
	redirect := front + "/dashboard?connected=" + url.QueryEscape(providerName)
	return c.Redirect(redirect, 302)
}

// POST /v1/integrations/disconnect
func DisconnectIntegration(c *fiber.Ctx) error {
	if err := requireOwnerAdmin(c); err != nil {
		return err
	}
	orgID, err := selectedOrgID(c)
	if err != nil {
		return err
	}

	type Body struct {
		Provider string `json:"provider"`
	}
	var body Body
	if err := c.BodyParser(&body); err != nil || body.Provider == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}

	var row models.OrganizationIntegration
	if err := config.DB.Where("organization_id = ? AND provider = ?", orgID, body.Provider).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(fiber.Map{"message": "Already disconnected"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "DB error"})
	}

	if row.Provider == "quickbooks" && row.RefreshTokenEnc != nil {
		encKey := os.Getenv("CONNECTOR_ENCRYPTION_KEY")
		if encKey != "" {
			if rt, err := utils.DecryptV2(*row.RefreshTokenEnc, encKey); err == nil {
				if drv, err := providers.GetDriver("quickbooks", ""); err == nil {
					_ = drv.Revoke(rt, "refresh_token")
				}
			}
		}
	}

	if err := config.DB.Model(&row).Updates(map[string]interface{}{
		"status":              "DISCONNECTED",
		"access_token_enc":    nil,
		"refresh_token_enc":   nil,
		"access_expires_at":   nil,
		"refresh_expires_at":  nil,
		"external_account_id": nil,
		"scopes":              nil,
	}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to disconnect"})
	}

	return c.JSON(fiber.Map{"message": "Disconnected"})
}
