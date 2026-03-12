package user

import (
	"donnes-backend/config"
	"donnes-backend/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrganizationResponse struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Role string `json:"role,omitempty"`
}

type UserResponse struct {
	Email             string                 `json:"email"`
	ID                string                 `json:"id"`
	FirstName         string                 `json:"firstName"`
	LastName          string                 `json:"lastName"`
	ShowWelcomePrompt bool                   `json:"showWelcomePrompt"`
	IsEmailVerified   bool                   `json:"isEmailVerified"`
	OrganizationList  []OrganizationResponse `json:"organizationList"`
}

// POST /v1/auth/get_user
func GetUser(c *fiber.Ctx) error {
	emailVal := c.Locals("email")
	if emailVal == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(
			fiber.Map{"error": "No email in token/middleware context"},
		)
	}
	email := emailVal.(string)

	var firstName, lastName string

	if givenVal := c.Locals("given_name"); givenVal != nil {
		firstName = givenVal.(string)
	} else {
		firstName = "Unknown"
	}

	if famVal := c.Locals("family_name"); famVal != nil {
		lastName = famVal.(string)
	} else {
		lastName = "User"
	}

	var user models.User
	result := config.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			newUser := models.User{
				ID:                uuid.New().String(),
				Email:             email,
				FirstName:         firstName,
				LastName:          lastName,
				ShowWelcomePrompt: true,
				IsEmailVerified:   false,
			}
			if cerr := config.DB.Create(&newUser).Error; cerr != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(
					fiber.Map{"error": "Failed to create user"},
				)
			}
			return c.JSON(toUserResponse(newUser, []OrganizationResponse{}))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Database error retrieving user"},
		)
	}

	var memberships []models.OrganizationMember
	config.DB.Where("user_id = ?", user.ID).Find(&memberships)

	var orgList []OrganizationResponse
	for _, membership := range memberships {
		var org models.Organization
		config.DB.Where("organization_id = ?", membership.OrganizationID).First(&org)

		orgList = append(orgList, OrganizationResponse{
			Name: org.OrganizationName,
			ID:   org.OrganizationID,
			Role: string(membership.Role),
		})
	}

	return c.JSON(toUserResponse(user, orgList))
}

func toUserResponse(u models.User, orgs []OrganizationResponse) UserResponse {
	return UserResponse{
		Email:             u.Email,
		ID:                u.ID,
		FirstName:         u.FirstName,
		LastName:          u.LastName,
		ShowWelcomePrompt: u.ShowWelcomePrompt,
		IsEmailVerified:   u.IsEmailVerified,
		OrganizationList:  orgs,
	}
}
