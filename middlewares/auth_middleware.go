package middlewares

import (
	"fmt"
	"numerra-backend/config"
	"numerra-backend/models"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func BuildKeyFunc() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid missing in token")
		}
		jwk, err := getKeyFromKid(kid)
		if err != nil {
			return nil, err
		}
		return jsonKeyToPublicKey(jwk)
	}
}

func KeycloakAuthMiddleware(c *fiber.Ctx) error {
	path := c.Path()

	if strings.HasPrefix(path, "/v1/integrations/callback/") {
		return c.Next()
	}

	if path == "/v1/organization/invite/accept_invite" ||
		path == "/v1/organization/invite/complete_onboarding" ||
		path == "/v1/user/send_reset_password_email" ||
		path == "/v1/user/verify_reset_password" {
		return c.Next()
	}

	if path == "/v1/user/get_user" ||
		path == "/v1/user/send_email_verification" ||
		path == "/v1/user/verify_user_email" {
		return validateJWT(c)
	}

	if path == "/v1/user/update_user" ||
		path == "/v1/user/delete_user" ||
		path == "/v1/user/update_welcome_prompt" ||
		path == "/v1/organization/create_organization" ||
		path == "/v1/auth/set_session_org" ||
		path == "/v1/auth/logout" {
		return validateJWTWithEmail(c)
	}

	return validateJWTSelectedOrgEmail(c)
}

func validateJWT(c *fiber.Ctx) error {
	claims, err := parseAndValidateToken(c)
	if err != nil {
		return err
	}
	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No email in token"})
	}
	c.Locals("email", email)
	if givenName, hasGivenName := claims["given_name"].(string); hasGivenName {
		c.Locals("given_name", givenName)
	}
	if familyName, hasFamilyName := claims["family_name"].(string); hasFamilyName {
		c.Locals("family_name", familyName)
	}
	return c.Next()
}

func validateJWTWithEmail(c *fiber.Ctx) error {
	claims, err := parseAndValidateToken(c)
	if err != nil {
		return err
	}
	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No email in token"})
	}
	c.Locals("email", email)
	if givenName, hasGivenName := claims["given_name"].(string); hasGivenName {
		c.Locals("given_name", givenName)
	}
	if familyName, hasFamilyName := claims["family_name"].(string); hasFamilyName {
		c.Locals("family_name", familyName)
	}
	return checkEmailVerified(c)
}

func validateJWTSelectedOrgEmail(c *fiber.Ctx) error {
	claims, err := parseAndValidateToken(c)
	if err != nil {
		return err
	}

	email, ok := claims["email"].(string)
	if !ok || email == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No email in token"})
	}
	c.Locals("email", email)
	if givenName, hasGivenName := claims["given_name"].(string); hasGivenName {
		c.Locals("given_name", givenName)
	}
	if familyName, hasFamilyName := claims["family_name"].(string); hasFamilyName {
		c.Locals("family_name", familyName)
	}

	var session models.SessionMetadata
	if err := config.DB.Where("user_email = ?", email).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "No selected org found. Please set session org first."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB error reading session"})
	}
	if session.SelectedOrgID == nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "No selected org found. Please set session org first."})
	}

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found in DB"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error retrieving user"})
	}

	var membership models.OrganizationMember
	if err := config.DB.Where("user_id = ? AND organization_id = ?", user.ID, *session.SelectedOrgID).
		First(&membership).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "User not a member of selected organization"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error checking membership"})
	}

	c.Locals("orgRole", membership.Role)
	c.Locals("selectedOrgID", *session.SelectedOrgID)
	return checkEmailVerified(c)
}

func parseAndValidateToken(c *fiber.Ctx) (jwt.MapClaims, error) {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing Authorization header"})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Authorization format"})
	}
	tokenStr := parts[1]

	issuer := os.Getenv("KEYCLOAK_ISSUER_URL")

	token, err := jwt.Parse(tokenStr,
		BuildKeyFunc(),
		jwt.WithIssuer(issuer),
	)

	if err != nil || !token.Valid {
		c.Status(fiber.StatusUnauthorized)
		return nil, c.JSON(fiber.Map{"error": "Invalid token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.Status(fiber.StatusUnauthorized)
		return nil, c.JSON(fiber.Map{"error": "Invalid token claims"})
	}

	return claims, nil
}

func checkEmailVerified(c *fiber.Ctx) error {
	emailVal := c.Locals("email")
	if emailVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "No email in context"},
		)
	}
	email := emailVal.(string)

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "User not found in DB"},
		)
	}

	if !user.IsEmailVerified {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{"error": "User email not verified"},
		)
	}

	return c.Next()
}
