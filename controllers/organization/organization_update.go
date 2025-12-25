package organization

import (
	"numerra-backend/config"
	"numerra-backend/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// PUT /v1/organization/update_organization
func UpdateOrganization(c *fiber.Ctx) error {
	roleVal := c.Locals("orgRole")
	if roleVal == nil || roleVal.(models.Role) != models.ROLE_OWNER {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "ROLE_OWNER required"})
	}

	orgIDVal := c.Locals("selectedOrgID")
	if orgIDVal == nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "No selected organization"})
	}
	orgID := orgIDVal.(string)

	type OrgUpdateRequest struct {
		OrganizationName   *string  `json:"organizationName"`
		IncorporationState *string  `json:"incorporationState"`
		IncorporationYear  *int     `json:"incorporationYear"`
		Industry           *string  `json:"industry"`
		Revenue            *float64 `json:"revenue"`
	}
	var req OrgUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}

	var org models.Organization
	if err := config.DB.Where("organization_id = ?", orgID).First(&org).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Org not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "DB error retrieving org"})
	}

	if req.OrganizationName != nil {
		org.OrganizationName = *req.OrganizationName
	}
	if req.IncorporationState != nil {
		org.IncorporationState = *req.IncorporationState
	}
	if req.IncorporationYear != nil {
		org.IncorporationYear = *req.IncorporationYear
	}
	if req.Industry != nil {
		org.Industry = *req.Industry
	}
	if req.Revenue != nil {
		org.Revenue = *req.Revenue
	}

	if err := config.DB.Save(&org).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update org"})
	}
	return c.JSON(fiber.Map{"message": "Organization updated", "organization": org})
}
