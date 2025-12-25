package organization

import (
	"numerra-backend/config"
	"numerra-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// DELETE /v1/organization/delete_user
func DeleteOrganizationUser(c *fiber.Ctx) error {
	roleVal := c.Locals("orgRole")
	if roleVal == nil {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{"error": "orgRole missing"},
		)
	}
	requestorRole := roleVal.(models.Role)

	orgIDVal := c.Locals("selectedOrgID")
	if orgIDVal == nil {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{"error": "No selected org in session"},
		)
	}
	orgID := orgIDVal.(string)

	if requestorRole != models.ROLE_OWNER && requestorRole != models.ROLE_ADMIN {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{"error": "Must be ROLE_OWNER or ROLE_ADMIN to delete users"},
		)
	}

	type Body struct {
		UserIDs []string `json:"userIds"`
	}
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid request body"},
		)
	}
	if len(body.UserIDs) == 0 {
		return c.JSON(fiber.Map{"message": "No users to delete"})
	}

	deletableUserIDs := make([]string, 0, len(body.UserIDs))

	for _, uid := range body.UserIDs {
		var membership models.OrganizationMember
		if err := config.DB.Where("user_id = ? AND organization_id = ?", uid, orgID).
			First(&membership).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				continue
			}
			return c.Status(fiber.StatusInternalServerError).JSON(
				fiber.Map{"error": "Failed to check membership"},
			)
		}

		if requestorRole == models.ROLE_ADMIN && membership.Role == models.ROLE_OWNER {
			continue
		}

		deletableUserIDs = append(deletableUserIDs, uid)
	}

	if len(deletableUserIDs) == 0 {
		return c.JSON(fiber.Map{"message": "No users were deletable"})
	}

	if err := config.DB.Where(
		"organization_id = ? AND user_id IN ?",
		orgID,
		deletableUserIDs,
	).Delete(&models.OrganizationMember{}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to delete members"},
		)
	}

	return c.JSON(fiber.Map{
		"message":      "Users removed from organization",
		"deletedUsers": deletableUserIDs,
	})
}
