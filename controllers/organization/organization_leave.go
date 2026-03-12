package organization

import (
	"donnes-backend/config"
	"donnes-backend/models"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// POST /v1/organization/leave
func LeaveOrganization(c *fiber.Ctx) error {
	orgVal := c.Locals("selectedOrgID")
	if orgVal == nil || orgVal == "" {
		return c.Status(http.StatusForbidden).JSON(
			fiber.Map{"error": "No selected organization in session"},
		)
	}
	selectedOrgID := orgVal.(string)

	emailVal := c.Locals("email")
	if emailVal == nil {
		return c.Status(http.StatusUnauthorized).JSON(
			fiber.Map{"error": "No user email in context"},
		)
	}
	email := emailVal.(string)

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusForbidden).JSON(
				fiber.Map{"error": "User not found in DB"},
			)
		}
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "DB error loading user"},
		)
	}

	var membership models.OrganizationMember
	if err := config.DB.Where("user_id = ? AND organization_id = ?", user.ID, selectedOrgID).
		First(&membership).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusForbidden).JSON(
				fiber.Map{"error": "User is not a member of the selected organization"},
			)
		}
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "DB error checking membership"},
		)
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "failed to begin DB transaction"},
		)
	}

	var totalMembers int64
	if errCount := tx.Model(&models.OrganizationMember{}).
		Where("organization_id = ?", selectedOrgID).
		Count(&totalMembers).Error; errCount != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "DB error counting members"},
		)
	}

	if errDel := tx.Delete(&membership).Error; errDel != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to remove membership"},
		)
	}

	if totalMembers == 1 {
		var org models.Organization
		if errO := tx.Where("organization_id = ?", selectedOrgID).First(&org).Error; errO != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(
				fiber.Map{"error": "DB error loading organization"},
			)
		}
		if errDelO := tx.Delete(&org).Error; errDelO != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(
				fiber.Map{"error": "Failed to delete organization"},
			)
		}
	} else {
		if membership.Role == models.ROLE_OWNER {
			var earliestAdmin models.OrganizationMember
			errEA := tx.Where("organization_id = ? AND role = ?", selectedOrgID, models.ROLE_ADMIN).
				Order("joined_at ASC").
				First(&earliestAdmin).Error
			if errEA != nil {
				if errEA == gorm.ErrRecordNotFound {
					var countOthers int64
					if errC2 := tx.Model(&models.OrganizationMember{}).
						Where("organization_id = ? AND role IN (?, ?, ?)",
							selectedOrgID,
							models.ROLE_OWNER,
							models.ROLE_ADMIN,
							models.ROLE_EDITOR,
						).Count(&countOthers).Error; errC2 == nil {
						log.Println("countOthers:", countOthers)
					}
					if errDO := deleteOrganizationCompletely(tx, selectedOrgID); errDO != nil {
						tx.Rollback()
						return c.Status(http.StatusInternalServerError).JSON(
							fiber.Map{"error": "Failed to delete org when no admin found"},
						)
					}
				} else {
					tx.Rollback()
					return c.Status(http.StatusInternalServerError).JSON(
						fiber.Map{"error": "DB error searching earliest admin"},
					)
				}
			} else {
				if errUp := tx.Model(&earliestAdmin).Update("role", models.ROLE_OWNER).Error; errUp != nil {
					tx.Rollback()
					return c.Status(http.StatusInternalServerError).JSON(
						fiber.Map{"error": "Failed to transfer ownership to earliest admin"},
					)
				}
			}
		}
	}

	if errC := tx.Commit().Error; errC != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to commit transaction"},
		)
	}

	return c.JSON(fiber.Map{"message": "Organization left successfully"})
}

func deleteOrganizationCompletely(tx *gorm.DB, orgID string) error {
	var org models.Organization
	if err := tx.Where("organization_id = ?", orgID).First(&org).Error; err != nil {
		return err
	}
	if errDel := tx.Delete(&org).Error; errDel != nil {
		return errDel
	}
	return nil
}
