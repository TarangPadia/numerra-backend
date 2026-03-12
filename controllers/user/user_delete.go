package user

import (
	"donnes-backend/config"
	"donnes-backend/models"
	"donnes-backend/utils"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// DELETE /v1/user/delete_user
func DeleteUser(c *fiber.Ctx) error {
	emailVal := c.Locals("email")
	if emailVal == nil {
		return c.Status(http.StatusUnauthorized).JSON(
			fiber.Map{"error": "No email in context"},
		)
	}
	email := emailVal.(string)

	tx := config.DB.Begin()
	if tx.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to start DB transaction"},
		)
	}

	var user models.User
	if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			tx.Rollback()
			return c.Status(http.StatusNotFound).JSON(
				fiber.Map{"error": "User not found"},
			)
		}
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "DB error retrieving user"},
		)
	}

	var memberships []models.OrganizationMember
	if err := tx.Where("user_id = ?", user.ID).Find(&memberships).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to load memberships"},
		)
	}

	for _, m := range memberships {
		var memberCount int64
		if err := tx.Model(&models.OrganizationMember{}).
			Where("organization_id = ?", m.OrganizationID).
			Count(&memberCount).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(
				fiber.Map{"error": "Failed to count membership"},
			)
		}

		if memberCount == 1 {
			if err := deleteOrganization(tx, m.OrganizationID); err != nil {
				tx.Rollback()
				return c.Status(http.StatusInternalServerError).JSON(
					fiber.Map{"error": err.Error()},
				)
			}
			continue
		}

		if m.Role == models.ROLE_OWNER {
			var earliestAdmin models.OrganizationMember
			err := tx.Where("organization_id = ? AND role = ?", m.OrganizationID, models.ROLE_ADMIN).
				Order("joined_at ASC").
				First(&earliestAdmin).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					if err2 := handleNoAdminCase(tx, m.OrganizationID, user.ID); err2 != nil {
						tx.Rollback()
						return c.Status(http.StatusInternalServerError).JSON(
							fiber.Map{"error": err2.Error()},
						)
					}
				} else {
					tx.Rollback()
					return c.Status(http.StatusInternalServerError).JSON(
						fiber.Map{"error": "Error searching earliest admin"},
					)
				}
			} else {
				if err2 := tx.Model(&earliestAdmin).Update("role", models.ROLE_OWNER).Error; err2 != nil {
					tx.Rollback()
					return c.Status(http.StatusInternalServerError).JSON(
						fiber.Map{"error": "Failed to transfer ownership"},
					)
				}
			}
		}
	}

	if err := tx.Where("user_id = ?", user.ID).
		Delete(&models.OrganizationMember{}).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to remove user from memberships"},
		)
	}

	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to delete user locally"},
		)
	}

	if err := utils.DeleteKeycloakUserByEmail(email); err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": fmt.Sprintf("Keycloak user deletion failed: %v", err)},
		)
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to commit transaction"},
		)
	}

	return c.JSON(fiber.Map{"message": "User (and possibly organizations) deleted successfully"})
}

func deleteOrganization(tx *gorm.DB, orgID string) error {
	var org models.Organization
	if err := tx.Where("organization_id = ?", orgID).First(&org).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return fmt.Errorf("DB error retrieving org for deletion: %v", err)
	}
	if err := tx.Delete(&org).Error; err != nil {
		return fmt.Errorf("failed to delete org: %v", err)
	}
	return nil
}

func handleNoAdminCase(tx *gorm.DB, orgID, deletingUserID string) error {
	var countOtherOwners int64
	err := tx.Model(&models.OrganizationMember{}).
		Where("organization_id = ? AND role = ? AND user_id != ?", orgID, models.ROLE_OWNER, deletingUserID).
		Count(&countOtherOwners).Error
	if err != nil {
		return fmt.Errorf("error checking other owners: %v", err)
	}

	if countOtherOwners > 0 {
		return nil
	}

	return deleteOrganization(tx, orgID)
}
