package organization

import (
	"donnes-backend/config"
	"donnes-backend/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// POST /v1/organization/create_organization
func CreateOrganization(c *fiber.Ctx) error {
	type OrgRequest struct {
		Name               string  `json:"organizationName"`
		IncorporationState string  `json:"incorporationState"`
		IncorporationYear  int     `json:"incorporationYear"`
		Industry           string  `json:"industry"`
		Revenue            float64 `json:"revenue"`
	}
	var req OrgRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).
			JSON(fiber.Map{"error": "Invalid request"})
	}

	emailVal := c.Locals("email")
	if emailVal == nil {
		return c.Status(fiber.StatusUnauthorized).
			JSON(fiber.Map{"error": "No email in context"})
	}
	email := emailVal.(string)

	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).
				JSON(fiber.Map{"error": "User not found"})
		}
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "DB error finding user"})
	}

	newOrg := models.Organization{
		OrganizationID:     uuid.New().String(),
		OrganizationName:   req.Name,
		IncorporationState: req.IncorporationState,
		IncorporationYear:  req.IncorporationYear,
		Industry:           req.Industry,
		Revenue:            req.Revenue,
	}
	if err := config.DB.Create(&newOrg).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Failed to create organization"})
	}

	newMember := models.OrganizationMember{
		MemberID:       uuid.New().String(),
		UserID:         user.ID,
		OrganizationID: newOrg.OrganizationID,
		Role:           models.ROLE_OWNER,
		JoinedAt:       time.Now(),
	}
	if err := config.DB.Create(&newMember).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(fiber.Map{"error": "Failed to create organization member"})
	}

	return c.JSON(fiber.Map{
		"message":          "Organization created",
		"organizationId":   newOrg.OrganizationID,
		"organizationName": newOrg.OrganizationName,
	})
}
