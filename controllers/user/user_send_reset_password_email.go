package user

import (
	"donnes-backend/config"
	"donnes-backend/models"
	"donnes-backend/templates"
	"donnes-backend/utils"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// POST /v1/user/send_reset_password_email
func SendResetPasswordEmail(c *fiber.Ctx) error {
	type ReqBody struct {
		Email string `json:"email"`
	}
	var body ReqBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}
	if body.Email == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Email is required"})
	}

	var user models.User
	if err := config.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(fiber.Map{"message": "If that email exists, a reset link was sent"})
		}
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "DB error checking user"},
		)
	}

	resetID := uuid.New().String()
	resetCode := uuid.New().String()
	expires := time.Now().Add(30 * time.Minute)

	pr := models.PasswordResetCode{
		ID:        resetID,
		Email:     user.Email,
		Code:      resetCode,
		ExpiresAt: expires,
	}
	if err2 := config.DB.Create(&pr).Error; err2 != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to create password reset code"},
		)
	}

	plain := fmt.Sprintf("%s|%s", user.Email, resetCode)
	encKey := os.Getenv("INVITE_ENCRYPTION_KEY")
	encrypted, errEnc := utils.Encrypt(plain, encKey)
	if errEnc != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to encrypt reset link code"},
		)
	}

	frontURL := os.Getenv("FRONTEND_BASE_URL")
	link := fmt.Sprintf("%s/reset-password?code=%s", frontURL, encrypted)

	emailBody := templates.EmailResetPasswordTemplate(link)
	subject := "Reset Password Request"

	if errSend := utils.SendEmailGomail(user.Email, subject, emailBody); errSend != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to send reset email"},
		)
	}

	return c.JSON(fiber.Map{"message": "Reset password email sent"})
}
