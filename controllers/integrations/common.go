package integrations

import (
	"numerra-backend/config"
	"numerra-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func requireOwnerAdmin(c *fiber.Ctx) error {
	roleVal := c.Locals("orgRole")
	if roleVal == nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "orgRole missing"})
	}
	role := roleVal.(models.Role)
	if role != models.ROLE_OWNER && role != models.ROLE_ADMIN {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "ROLE_OWNER or ROLE_ADMIN required"})
	}
	return nil
}

func selectedOrgID(c *fiber.Ctx) (string, error) {
	orgIDVal := c.Locals("selectedOrgID")
	if orgIDVal == nil {
		return "", c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "No selected org"})
	}
	return orgIDVal.(string), nil
}

func currentUser(c *fiber.Ctx) (*models.User, error) {
	emailVal := c.Locals("email")
	if emailVal == nil {
		return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "No email in context"})
	}
	email := emailVal.(string)

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
		}
		return nil, c.Status(500).JSON(fiber.Map{"error": "DB error"})
	}
	return &user, nil
}
