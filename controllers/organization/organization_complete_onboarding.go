package organization

import (
	"donnes-backend/config"
	"donnes-backend/models"
	"donnes-backend/utils"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// POST /v1/organization/invite/complete_onboarding
func CompleteOnboarding(c *fiber.Ctx) error {
	type OnboardingRequest struct {
		OrgID     string `json:"orgId"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Password  string `json:"password"`
	}
	var req OnboardingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid request body"},
		)
	}
	if req.OrgID == "" || req.Email == "" || req.Password == "" {
		return c.Status(http.StatusBadRequest).JSON(
			fiber.Map{"error": "Missing orgId, email, or password"},
		)
	}

	var inv models.Invitation
	if err := config.DB.Where("user_email = ? AND organization_id = ?", req.Email, req.OrgID).
		First(&inv).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				fiber.Map{"error": "No invitation found for this email/org"},
			)
		}
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "DB error reading invitation"},
		)
	}
	if time.Now().After(inv.ExpiresAt) {
		return c.Status(http.StatusBadRequest).JSON(
			fiber.Map{"error": "Invitation expired"},
		)
	}
	if inv.Status != models.StatusIncomplete {
		return c.Status(http.StatusBadRequest).JSON(
			fiber.Map{"error": fmt.Sprintf("Invitation not in 'INCOMPLETE' state. Current status=%s", inv.Status)},
		)
	}

	var user models.User
	if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(
				fiber.Map{"error": "User not found"},
			)
		}
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "DB error retrieving user"},
		)
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	if err := config.DB.Save(&user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed updating local user names"},
		)
	}

	encryptionKey := os.Getenv("INVITE_ENCRYPTION_KEY")
	decryptedPwd, errDec := utils.Decrypt(req.Password, encryptionKey)
	if errDec != nil {
		return c.Status(http.StatusBadRequest).JSON(
			fiber.Map{"error": fmt.Sprintf("Cannot decrypt password: %v", errDec)},
		)
	}

	kcUserID, errF := utils.FindKeycloakUserIDByEmail(req.Email)
	if errF != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": fmt.Sprintf("Cannot find Keycloak user by email: %v", errF)},
		)
	}
	if errUp := utils.UpdateKeycloakUserName(kcUserID, req.FirstName, req.LastName); errUp != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": fmt.Sprintf("Failed updating Keycloak user name: %v", errUp)},
		)
	}
	if errPwd := utils.SetKeycloakUserPassword(kcUserID, decryptedPwd); errPwd != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": fmt.Sprintf("Failed setting Keycloak password: %v", errPwd)},
		)
	}

	inv.Status = models.StatusConfirmed
	inv.ExpiresAt = time.Now()
	if err := config.DB.Save(&inv).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed updating invitation to CONFIRMED"},
		)
	}

	return c.JSON(fiber.Map{
		"message": "Onboarding completed. You may now log in with your new credentials.",
	})
}
