package user

import (
	"donnes-backend/config"
	"donnes-backend/models"
	"donnes-backend/utils"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// POST /v1/user/verify_reset_password
func VerifyResetPassword(c *fiber.Ctx) error {
	type ReqBody struct {
		Code              string `json:"code"`
		EncryptedPassword string `json:"encryptedPassword"`
	}
	var body ReqBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if body.Code == "" || body.EncryptedPassword == "" {
		return c.Status(http.StatusBadRequest).JSON(
			fiber.Map{"error": "Missing code or encryptedPassword"},
		)
	}

	encKey := os.Getenv("INVITE_ENCRYPTION_KEY")

	decrypted, errDec := utils.Decrypt(body.Code, encKey, "INVITE")
	if errDec != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"isVerified": false,
			"error":      fmt.Sprintf("Cannot decrypt link code: %v", errDec),
		})
	}
	parts := utils.ParseDecryptedInvite(decrypted)
	if len(parts) != 2 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"isVerified": false,
			"error":      "Malformed code param",
		})
	}
	emailFromLink := parts[0]
	codeFromLink := parts[1]

	var pr models.PasswordResetCode
	if err := config.DB.Where("code = ?", codeFromLink).First(&pr).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{"isVerified": false, "error": "Invalid code"})
		}
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "DB error retrieving password reset code"},
		)
	}
	if pr.Email != emailFromLink {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"isVerified": false,
			"error":      "Email mismatch in DB vs link",
		})
	}
	if time.Now().After(pr.ExpiresAt) {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"isVerified": false,
			"error":      "Code expired",
		})
	}

	newPlainPwd, errDecPwd := utils.Decrypt(body.EncryptedPassword, encKey, "PASSWORD")
	if errDecPwd != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"isVerified": false,
			"error":      fmt.Sprintf("Cannot decrypt new password: %v", errDecPwd),
		})
	}
	kcUserID, errKC := utils.FindKeycloakUserIDByEmail(pr.Email)
	if errKC != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"isVerified": false,
			"error":      fmt.Sprintf("Cannot find user in Keycloak: %v", errKC),
		})
	}
	if errPwd := utils.SetKeycloakUserPassword(kcUserID, newPlainPwd); errPwd != nil {
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"isVerified": false,
			"error":      fmt.Sprintf("Failed to set password in Keycloak: %v", errPwd),
		})
	}

	config.DB.Delete(&pr)

	return c.JSON(fiber.Map{
		"email":      pr.Email,
		"isVerified": true,
	})
}
