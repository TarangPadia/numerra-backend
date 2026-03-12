package routes

import (
	"donnes-backend/controllers/auth"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(r fiber.Router) {
	r.Post("/set_session_org", auth.SetSessionOrg)
	r.Post("/logout", auth.Logout)
	r.Get("/get_user_session", auth.GetUserSession)
}
