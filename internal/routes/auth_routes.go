package routes

import (
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/internal/controllers/auth_controller"
)

func SetupAuthRoutes(api fiber.Router, authController *auth_controller.AuthController) {
	auth := api.Group("/auth")
	
	// The 3-Step Registration Flow
	auth.Post("/send-otp", authController.SendOTP)
	auth.Post("/verify-otp", authController.VerifyOTP)
	auth.Post("/set-password", authController.SetPassword)
	
	// Keep login working
	auth.Post("/login", authController.Login)

	auth.Post("/forgot-password/send-otp", authController.ForgotPasswordSendOTP)
    auth.Post("/forgot-password/reset", authController.ResetPassword)

}