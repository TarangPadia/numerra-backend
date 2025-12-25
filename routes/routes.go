package routes

import (
	"numerra-backend/middlewares"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	app.Use(middlewares.KeycloakAuthMiddleware)

	authGroup := app.Group("/v1/auth")
	AuthRoutes(authGroup)

	userGroup := app.Group("/v1/user")
	UserRoutes(userGroup)

	orgGroup := app.Group("/v1/organization")
	OrganizationRoutes(orgGroup)

	intGroup := app.Group("/v1/integrations")
	IntegrationRoutes(intGroup)
}
