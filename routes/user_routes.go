package routes

import (
	"donnes-backend/controllers/user"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(r fiber.Router) {
	r.Post("/get_user", user.GetUser)
	r.Delete("/delete_user", user.DeleteUser)
	r.Put("/update_user", user.UpdateUser)
	r.Put("/update_welcome_prompt", user.UpdateWelcomePrompt)
	r.Post("/send_email_verification", user.SendEmailVerification)
	r.Post("/verify_user_email", user.VerifyUserEmail)
	r.Post("/send_reset_password_email", user.SendResetPasswordEmail)
	r.Post("/verify_reset_password", user.VerifyResetPassword)
}
