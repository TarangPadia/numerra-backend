package organization

import (
	"numerra-backend/config"
	"numerra-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// DELETE /v1/organization/delete_organization
func DeleteOrganization(c *fiber.Ctx) error {
	roleVal := c.Locals("orgRole")
	if roleVal == nil || roleVal.(models.Role) != models.ROLE_OWNER {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{"error": "ROLE_OWNER required"},
		)
	}

	orgIDVal := c.Locals("selectedOrgID")
	if orgIDVal == nil {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{"error": "No selected organization in session"},
		)
	}
	orgID := orgIDVal.(string)

	var org models.Organization
	if err := config.DB.Where("organization_id = ?", orgID).First(&org).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Org not found"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "DB error retrieving org"},
		)
	}

	if err := config.DB.Delete(&org).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to delete org"},
		)
	}

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
			return c.Status(fiber.StatusUnauthorized).JSON(
				fiber.Map{"error": "User not found in DB"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Error retrieving user"},
		)
	}

	if err := config.DB.Model(&models.SessionMetadata{}).
		Where("user_email = ?", user.Email).
		Update("selected_org_id", nil).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to clear selected org in session metadata"},
		)
	}

	return c.JSON(fiber.Map{
		"message": "Organization deleted and session updated",
	})
}
