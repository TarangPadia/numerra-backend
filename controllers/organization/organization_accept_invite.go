package organization

import (
	"donnes-backend/config"
	"donnes-backend/models"
	"donnes-backend/utils"
	"fmt"
	"net/http"
	"net/mail"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// POST /v1/organization/invite/accept_invite
func AcceptInvite(c *fiber.Ctx) error {
	type AcceptRequest struct {
		EncryptedCode string `json:"code"`
	}

	var req AcceptRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid request"},
		)
	}

	decrypted, err := utils.Decrypt(req.EncryptedCode, os.Getenv("INVITE_ENCRYPTION_KEY"), "INVITE")
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(
			fiber.Map{"error": "Invalid invitation code"},
		)
	}

	parts := utils.ParseDecryptedInvite(decrypted)
	if len(parts) != 4 {
		return c.Status(http.StatusUnauthorized).JSON(
			fiber.Map{"error": "Malformed invitation code"},
		)
	}
	inviteeEmail := parts[0]
	orgID := parts[1]
	//parts[2] = inviterEmail
	//parts[3] = random

	// 1) load invitation
	var inv models.Invitation
	err2 := config.DB.
		Where("user_email = ? AND organization_id = ? AND expires_at > ?", inviteeEmail, orgID, time.Now()).
		First(&inv).Error
	if err2 != nil {
		if err2 == gorm.ErrRecordNotFound {
			return c.Status(http.StatusUnauthorized).JSON(
				fiber.Map{"error": "Invalid or expired invitation"},
			)
		}
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "DB error reading invitation"},
		)
	}

	// check status
	switch inv.Status {
	case models.StatusConfirmed:
		return c.Status(http.StatusBadRequest).JSON(
			fiber.Map{"error": "Invitation already confirmed"},
		)
	case models.StatusIncomplete:
		// user already partially accepted
		return c.JSON(fiber.Map{
			"message":          "Onboarding pending",
			"userEmail":        inv.UserEmail,
			"orgId":            inv.OrganizationID,
			"alreadyExisted":   false,
			"invitationStatus": inv.Status,
		})
	case models.StatusPending:
		// proceed
	default:
		return c.Status(http.StatusBadRequest).JSON(
			fiber.Map{"error": fmt.Sprintf("Invalid invitation status: %s", inv.Status)},
		)
	}

	if _, parseErr := mail.ParseAddress(inviteeEmail); parseErr != nil {
		return c.Status(http.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid invitee email format"},
		)
	}

	// DB transaction
	tx := config.DB.Begin()
	if tx.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to start DB transaction"},
		)
	}

	// check if user is in DB
	var existingUser models.User
	errUser := tx.Where("email = ?", inviteeEmail).First(&existingUser).Error
	if errUser == nil {
		// user found => check membership
		var mem models.OrganizationMember
		errMem := tx.Where("organization_id = ? AND user_id = ?", orgID, existingUser.ID).
			First(&mem).Error
		if errMem == nil {
			// membership found => user already in org
			tx.Rollback()
			return c.Status(http.StatusBadRequest).JSON(
				fiber.Map{"error": "User already in organization"},
			)
		} else if errMem != gorm.ErrRecordNotFound {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(
				fiber.Map{"error": "DB error checking membership"},
			)
		}
		// user found, not in org => create membership, set invite status = CONFIRMED
		newMember := models.OrganizationMember{
			MemberID:       uuid.New().String(),
			UserID:         existingUser.ID,
			OrganizationID: orgID,
			Role:           inv.Role,
			JoinedAt:       time.Now(),
		}
		if errC := tx.Create(&newMember).Error; errC != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(
				fiber.Map{"error": "Failed to create membership for existing user"},
			)
		}

		if errUp := tx.Model(&inv).Updates(map[string]interface{}{
			"status": models.StatusConfirmed,
		}).Error; errUp != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(
				fiber.Map{"error": "Failed updating invitation status to CONFIRMED"},
			)
		}
		tx.Commit()

		return c.JSON(fiber.Map{
			"message":          "Invitation accepted; existing user added to org",
			"userEmail":        existingUser.Email,
			"userId":           existingUser.ID,
			"orgId":            orgID,
			"alreadyExisted":   true,
			"invitationStatus": models.StatusConfirmed,
		})
	}

	// if errUser != nil && errUser != gorm.ErrRecordNotFound {
	// 	tx.Rollback()
	// 	return c.Status(http.StatusInternalServerError).JSON(
	// 		fiber.Map{"error": "DB error retrieving user"},
	// 	)
	// }

	// user not in DB => partial approach
	// create user in Keycloak
	if errKC := utils.CreateKeycloakUser(utils.KeycloakUserRequest{
		Email:     inviteeEmail,
		Username:  inviteeEmail,
		FirstName: "Temp",
		LastName:  "User",
		Enabled:   true,
	}); errKC != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": fmt.Sprintf("Failed to create user in Keycloak: %v", errKC)},
		)
	}

	if errVfy := utils.VerifyKeycloakUserEmailByEmail(inviteeEmail); errVfy != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": fmt.Sprintf("Failed verifying email in Keycloak: %v", errVfy)},
		)
	}

	newUser := models.User{
		ID:                uuid.New().String(),
		Email:             inviteeEmail,
		FirstName:         "Temp",
		LastName:          "User",
		ShowWelcomePrompt: true,
		IsEmailVerified:   true,
	}
	if errCU := tx.Create(&newUser).Error; errCU != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to create user in local DB"},
		)
	}

	// create membership
	newMember := models.OrganizationMember{
		MemberID:       uuid.New().String(),
		UserID:         newUser.ID,
		OrganizationID: orgID,
		Role:           inv.Role,
		JoinedAt:       time.Now(),
	}
	if errNM := tx.Create(&newMember).Error; errNM != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to create membership"},
		)
	}

	// mark invitation => "INCOMPLETE"
	if errUp := tx.Model(&inv).Updates(map[string]interface{}{
		"status": models.StatusIncomplete,
	}).Error; errUp != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed updating invitation to INCOMPLETE"},
		)
	}
	tx.Commit()

	return c.JSON(fiber.Map{
		"message":          "Invitation accepted; new user created in partial mode",
		"userEmail":        inviteeEmail,
		"userId":           newUser.ID,
		"orgId":            orgID,
		"alreadyExisted":   false,
		"invitationStatus": models.StatusIncomplete,
	})
}
