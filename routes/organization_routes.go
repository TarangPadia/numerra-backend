package routes

import (
	org "numerra-backend/controllers/organization"

	"github.com/gofiber/fiber/v2"
)

func OrganizationRoutes(r fiber.Router) {
	r.Post("/create_organization", org.CreateOrganization)
	r.Put("/update_organization", org.UpdateOrganization)
	r.Delete("/delete_organization", org.DeleteOrganization)
	r.Get("/get_organization", org.GetOrganization)
	r.Post("/leave_organization", org.LeaveOrganization)
	r.Post("/invite/complete_onboarding", org.CompleteOnboarding)
	r.Post("/invite/send_invite", org.SendInvite)
	r.Post("/invite/accept_invite", org.AcceptInvite)
	r.Delete("/delete_user", org.DeleteOrganizationUser)
	r.Put("/change_user_role", org.ChangeUserRole)
}
