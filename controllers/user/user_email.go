package user

import (
	"crypto/rand"
	"donnes-backend/config"
	"donnes-backend/models"
	"donnes-backend/templates"
	"donnes-backend/utils"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// POST /v1/user/send_email_verification
func SendEmailVerification(c *fiber.Ctx) error {
	emailVal := c.Locals("email")
	if emailVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "No email in context"},
		)
	}
	email := emailVal.(string)

	otpVal := generateOTP()
	expMin, _ := strconv.Atoi(os.Getenv("OTP_EXPIRATION_MINUTES"))
	expires := time.Now().Add(time.Duration(expMin) * time.Minute)

	otpRecord := models.UserOTP{
		Email:     email,
		OTP:       otpVal,
		ExpiresAt: expires,
	}
	if err := config.DB.Create(&otpRecord).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to store OTP"},
		)
	}

	subject := "Email Verification"
	body := templates.EmailVerificationTemplate(otpVal)
	if err := utils.SendEmailGomail(email, subject, body); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to send email"},
		)
	}

	return c.JSON(
		fiber.Map{"message": "OTP sent successfully"},
	)
}

// POST /v1/user/verify_user_email
func VerifyUserEmail(c *fiber.Ctx) error {
	type VerifyRequest struct {
		OTP string `json:"otp"`
	}
	var req VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(
			fiber.Map{"error": "Invalid request"},
		)
	}

	emailVal := c.Locals("email")
	if emailVal == nil {
		return c.Status(http.StatusUnauthorized).JSON(
			fiber.Map{"error": "No email in context"},
		)
	}
	email := emailVal.(string)

	var otpRecord models.UserOTP
	if err := config.DB.Where("email = ? AND otp = ?", email, req.OTP).
		First(&otpRecord).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusUnauthorized).JSON(
				fiber.Map{"error": "Invalid OTP"},
			)
		}
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to retrieve OTP record"},
		)
	}

	if time.Now().After(otpRecord.ExpiresAt) {
		return c.Status(http.StatusUnauthorized).JSON(
			fiber.Map{"error": "OTP expired"},
		)
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to start transaction"},
		)
	}

	if err := tx.Model(&models.User{}).
		Where("email = ?", email).
		Update("is_email_verified", true).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to verify user in DB"},
		)
	}

	if err := utils.VerifyKeycloakUserEmailByEmail(email); err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": fmt.Sprintf("Failed to verify email in Keycloak: %v", err)},
		)
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to commit transaction"},
		)
	}

	return c.JSON(
		fiber.Map{"message": "Email verified successfully in local DB and Keycloak"},
	)
}

func generateOTP() string {
	b := make([]byte, 3)
	rand.Read(b)
	return fmt.Sprintf("%06d",
		int(b[0])%10*100000+
			int(b[1])%1000+
			int(b[2])%10,
	)
}
