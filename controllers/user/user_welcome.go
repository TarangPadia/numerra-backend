package user

import (
	"numerra-backend/config"
	"numerra-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// PUT /v1/user/update_welcome_prompt
func UpdateWelcomePrompt(c *fiber.Ctx) error {
	emailVal := c.Locals("email")
	if emailVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "No email in context"},
		)
	}
	email := emailVal.(string)

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "User not found"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Database error retrieving user"},
		)
	}

	if err := config.DB.Model(&user).
		Update("show_welcome_prompt", false).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to update welcome prompt"},
		)
	}

	return c.JSON(
		fiber.Map{"message": "Welcome prompt updated successfully"},
	)
}
