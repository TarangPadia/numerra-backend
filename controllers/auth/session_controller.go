package auth

import (
	"donnes-backend/config"
	"donnes-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// POST /v1/auth/set_session_org
func SetSessionOrg(c *fiber.Ctx) error {
	type ReqBody struct {
		OrgID *string `json:"organizationId"`
	}
	var body ReqBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	emailVal := c.Locals("email")
	if emailVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing email in context"})
	}
	email := emailVal.(string)

	var orgToSet *string
	if body.OrgID != nil && *body.OrgID != "" {
		orgID := *body.OrgID

		var user models.User
		if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "User not found"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error retrieving user"})
		}
		var membership models.OrganizationMember
		if err := config.DB.Where("user_id = ? AND organization_id = ?", user.ID, orgID).
			First(&membership).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "User not a member of given org"})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB error checking membership"})
		}
		orgToSet = &orgID
	} else {
		orgToSet = nil
	}

	var session models.SessionMetadata
	err := config.DB.Where("user_email = ?", email).First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			session.UserEmail = email
			session.SelectedOrgID = orgToSet
			if ce := config.DB.Create(&session).Error; ce != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create session"})
			}
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to query session"})
		}
	} else {
		session.SelectedOrgID = orgToSet
		if ue := config.DB.Save(&session).Error; ue != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update session"})
		}
	}

	return c.JSON(fiber.Map{"message": "Selected org set successfully"})
}

// POST /v1/auth/get_user_session
func GetUserSession(c *fiber.Ctx) error {
	emailVal := c.Locals("email")
	if emailVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing email in context"})
	}
	email := emailVal.(string)

	var session models.SessionMetadata
	if err := config.DB.Where("user_email = ?", email).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(fiber.Map{
				"selectedOrgID": nil,
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB error retrieving session"})
	}

	return c.JSON(fiber.Map{
		"selectedOrgID": session.SelectedOrgID,
	})
}

// POST /v1/auth/logout
func Logout(c *fiber.Ctx) error {
	emailVal := c.Locals("email")
	if emailVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "No email in context"},
		)
	}
	email := emailVal.(string)

	if err := config.DB.Where("user_email = ?", email).
		Delete(&models.SessionMetadata{}).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(
				fiber.Map{"message": "No active session found"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to delete session"},
		)
	}

	return c.JSON(fiber.Map{"message": "Logout successful"})
}
