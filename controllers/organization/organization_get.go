package organization

import (
	"donnes-backend/config"
	"donnes-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// GET /v1/organization/get_organization
func GetOrganization(c *fiber.Ctx) error {
	orgIDVal := c.Locals("selectedOrgID")
	if orgIDVal == nil {
		return c.Status(fiber.StatusForbidden).JSON(
			fiber.Map{"error": "No selected org"},
		)
	}
	orgID := orgIDVal.(string)

	var org models.Organization
	if err := config.DB.Where("organization_id = ?", orgID).First(&org).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(
				fiber.Map{"error": "Org not found"},
			)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "DB error retrieving org"},
		)
	}

	var memberships []models.OrganizationMember
	if err := config.DB.Where("organization_id = ?", orgID).
		Find(&memberships).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			fiber.Map{"error": "Error loading memberships"},
		)
	}

	type MemberResponse struct {
		UserID     string `json:"userId"`
		Email      string `json:"email"`
		FirstName  string `json:"firstName"`
		LastName   string `json:"lastName"`
		Role       string `json:"role"`
		IsAccepted bool   `json:"isAccepted"`
	}

	var members []MemberResponse

	hasMembershipEmail := make(map[string]bool)

	for _, m := range memberships {
		var u models.User
		if err := config.DB.Where("id = ?", m.UserID).First(&u).Error; err != nil {
			continue
		}
		members = append(members, MemberResponse{
			UserID:     u.ID,
			Email:      u.Email,
			FirstName:  u.FirstName,
			LastName:   u.LastName,
			Role:       string(m.Role),
			IsAccepted: true,
		})
		hasMembershipEmail[u.Email] = true
	}

	orgData := fiber.Map{
		"organizationID":     org.OrganizationID,
		"organizationName":   org.OrganizationName,
		"incorporationState": org.IncorporationState,
		"incorporationYear":  org.IncorporationYear,
		"industry":           org.Industry,
		"revenue":            org.Revenue,
	}

	return c.JSON(fiber.Map{
		"organization": orgData,
		"members":      members,
	})
}
