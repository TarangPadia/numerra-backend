package routes

import (
	"donnes-backend/controllers/integrations"

	"github.com/gofiber/fiber/v2"
)

func IntegrationRoutes(r fiber.Router) {
	r.Get("/status", integrations.GetIntegrationStatus)
	r.Post("/connect", integrations.ConnectIntegration)
	r.Post("/disconnect", integrations.DisconnectIntegration)
	r.Get("/callback/:provider", integrations.OAuthCallback)
}
