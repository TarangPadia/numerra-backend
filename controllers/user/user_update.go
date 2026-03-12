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

// PUT /v1/user/update_user
func UpdateUser(c *fiber.Ctx) error {
	emailVal := c.Locals("email")
	if emailVal == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "No email in context"})
	}
	email := emailVal.(string)

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "DB error retrieving user"})
	}

	type UpdateUserRequest struct {
		Email     *string `json:"email"`
		FirstName *string `json:"firstName"`
		LastName  *string `json:"lastName"`
	}
	var req UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	tx := config.DB.Begin()
	if tx.Error != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "Could not start DB transaction"})
	}

	oldEmail := user.Email

	if req.Email != nil {
		user.Email = *req.Email
		user.IsEmailVerified = false
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}

	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to update local user"},
		)
	}

	if err := utils.UpdateKeycloakUserByEmail(oldEmail, user.Email, user.FirstName, user.LastName); err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": fmt.Sprintf("Keycloak update failed: %v", err)},
		)
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(
			fiber.Map{"error": "Failed to commit transaction"},
		)
	}

	return c.JSON(user)
}
