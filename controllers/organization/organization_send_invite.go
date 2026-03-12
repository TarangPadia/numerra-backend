package organization

import (
	"donnes-backend/config"
	"donnes-backend/models"
	"donnes-backend/templates"
	"donnes-backend/utils"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// POST /v1/organization/invite/send_invite
func SendInvite(c *fiber.Ctx) error {
	roleVal := c.Locals("orgRole")
	if roleVal == nil {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{"error": "orgRole missing"},
		)
	}
	senderRole := roleVal.(models.Role)
	if senderRole != models.ROLE_OWNER && senderRole != models.ROLE_ADMIN {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{"error": "ROLE_OWNER or ROLE_ADMIN required"},
		)
	}

	orgIDVal := c.Locals("selectedOrgID")
	if orgIDVal == nil {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{"error": "No selected org"},
		)
	}
	orgID := orgIDVal.(string)

	emailVal := c.Locals("email")
	if emailVal == nil {
		return c.Status(http.StatusUnauthorized).JSON(
			fiber.Map{"error": "Inviter email missing in context"},
		)
	}
	inviterEmail := emailVal.(string)

	type SingleInvite struct {
		InviteeEmail string      `json:"inviteeEmail"`
		Role         models.Role `json:"role"`
	}
	type InviteRequest struct {
		Invitees []SingleInvite `json:"invitees"`
	}

	var req InviteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid request body"},
		)
	}
	if len(req.Invitees) == 0 {
		return c.JSON(fiber.Map{"message": "No invitees"})
	}

	var org models.Organization
	if err := config.DB.Where("organization_id = ?", orgID).First(&org).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to find org"},
		)
	}
	var inviter models.User
	if err := config.DB.Where("email = ?", inviterEmail).First(&inviter).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to find inviter user"},
		)
	}

	for _, iv := range req.Invitees {
		inviteeEmail := iv.InviteeEmail
		if _, parseErr := mail.ParseAddress(inviteeEmail); parseErr != nil {
			continue
		}

		var membership models.OrganizationMember
		errMem := config.DB.Where(`
			organization_id = ?
			AND user_id IN (SELECT id FROM users WHERE email = ?)`,
			orgID, inviteeEmail).First(&membership).Error
		if errMem == nil {
			return c.Status(fiber.StatusBadRequest).JSON(
				fiber.Map{
					"error": fmt.Sprintf("User %s is already part of this organization", inviteeEmail),
				},
			)
		} else if errMem != gorm.ErrRecordNotFound {
			return c.Status(http.StatusInternalServerError).JSON(
				fiber.Map{"error": "DB error checking membership"},
			)
		}

		config.DB.Model(&models.Invitation{}).
			Where("user_email = ? AND organization_id = ? AND status <> ? AND expires_at > ?",
				inviteeEmail, orgID, models.StatusConfirmed, time.Now()).
			Update("expires_at", time.Now())

		codePlain := fmt.Sprintf("%s|%s|%s|%s", inviteeEmail, orgID, inviterEmail, uuid.New().String())
		encrypted, encErr := utils.Encrypt(codePlain, os.Getenv("INVITE_ENCRYPTION_KEY"))
		if encErr != nil {
			return c.Status(http.StatusInternalServerError).JSON(
				fiber.Map{"error": "Error encrypting invitation code"},
			)
		}

		invRec := models.Invitation{
			ID:             uuid.New().String(),
			UserEmail:      inviteeEmail,
			OrganizationID: orgID,
			InviterUserID:  inviter.ID,
			Role:           iv.Role,
			Status:         models.StatusPending,
			ExpiresAt:      time.Now().Add(5 * 24 * time.Hour),
		}
		config.DB.Create(&invRec)

		log.Println(time.Now().Add(5 * 24 * time.Hour))

		frontURL := os.Getenv("FRONTEND_BASE_URL")
		link := fmt.Sprintf("%s/invitation?code=%s", frontURL, encrypted)

		subject := "Organization Invitation"
		body := templates.EmailInvitationTemplate(org.OrganizationName, inviter.FirstName, link)
		if err := utils.SendEmailGomail(inviteeEmail, subject, body); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(
				fiber.Map{"error": "Failed to send email"},
			)
		}
	}

	return c.JSON(fiber.Map{"message": "Invitations sent"})
}
