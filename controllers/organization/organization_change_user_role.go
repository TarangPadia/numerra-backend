package organization

import (
	"donnes-backend/config"
	"donnes-backend/models"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// PUT /v1/organization/change_user_role
func ChangeUserRole(c *fiber.Ctx) error {
	roleVal := c.Locals("orgRole")
	if roleVal == nil {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{"error": "No orgRole found in context"},
		)
	}
	requestorRole := roleVal.(models.Role)

	if requestorRole != models.ROLE_OWNER && requestorRole != models.ROLE_ADMIN {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{"error": "Must be ROLE_OWNER or ROLE_ADMIN to change user roles"},
		)
	}

	orgIDVal := c.Locals("selectedOrgID")
	if orgIDVal == nil {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{"error": "No selected org in session"},
		)
	}
	orgID := orgIDVal.(string)

	type RoleChange struct {
		UserID string      `json:"userId"`
		Role   models.Role `json:"role"`
	}
	var changes []RoleChange
	if err := c.BodyParser(&changes); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid JSON body"},
		)
	}
	if len(changes) == 0 {
		return c.JSON(fiber.Map{"message": "No role changes requested"})
	}

	type ChangeResult struct {
		UserID  string `json:"userId"`
		Role    string `json:"role"`
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	results := make([]ChangeResult, 0, len(changes))

	for _, ch := range changes {
		switch ch.Role {
		case models.ROLE_ADMIN, models.ROLE_EDITOR, models.ROLE_SPECTATOR, models.ROLE_OWNER:

		default:
			results = append(results, ChangeResult{
				UserID:  ch.UserID,
				Role:    string(ch.Role),
				Success: false,
				Message: fmt.Sprintf("Invalid role: %s", ch.Role),
			})
			continue
		}

		if requestorRole == models.ROLE_ADMIN && ch.Role == models.ROLE_OWNER {
			results = append(results, ChangeResult{
				UserID:  ch.UserID,
				Role:    string(ch.Role),
				Success: false,
				Message: "ADMIN cannot assign ROLE_OWNER",
			})
			continue
		}

		var membership models.OrganizationMember
		errMem := config.DB.
			Where("user_id = ? AND organization_id = ?", ch.UserID, orgID).
			First(&membership).Error
		if errMem != nil {
			if errMem == gorm.ErrRecordNotFound {
				results = append(results, ChangeResult{
					UserID:  ch.UserID,
					Role:    string(ch.Role),
					Success: false,
					Message: "User not found in organization",
				})
			} else {
				results = append(results, ChangeResult{
					UserID:  ch.UserID,
					Role:    string(ch.Role),
					Success: false,
					Message: "DB error retrieving membership",
				})
			}
			continue
		}

		if requestorRole == models.ROLE_ADMIN && membership.Role == models.ROLE_OWNER {
			results = append(results, ChangeResult{
				UserID:  ch.UserID,
				Role:    string(ch.Role),
				Success: false,
				Message: "ADMIN cannot change the role of an OWNER user",
			})
			continue
		}

		if err := config.DB.Model(&membership).Update("role", ch.Role).Error; err != nil {
			results = append(results, ChangeResult{
				UserID:  ch.UserID,
				Role:    string(ch.Role),
				Success: false,
				Message: "Failed to update role in DB",
			})
			continue
		}
		results = append(results, ChangeResult{
			UserID:  ch.UserID,
			Role:    string(ch.Role),
			Success: true,
			Message: "Role updated successfully",
		})
	}

	return c.JSON(fiber.Map{
		"results": results,
	})
}
