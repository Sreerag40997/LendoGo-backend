package routes

import (
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/internal/controllers/auth_controller"
	"lendogo-backend/internal/middlewares"
	"lendogo-backend/internal/services"
)

func SetupAuthRoutes(api fiber.Router, authController *auth_controller.AuthController, configService services.ConfigService) {
	auth := api.Group("/auth")
	
	// The 3-Step Registration Flow
	registerMiddleware := middlewares.RequireFeature(configService, "register")
	auth.Post("/send-otp", registerMiddleware, authController.SendOTP)
	auth.Post("/verify-otp", registerMiddleware, authController.VerifyOTP)
	auth.Post("/set-password", registerMiddleware, authController.SetPassword)
	
	// Keep login working 
	auth.Post("/login", middlewares.RequireFeature(configService, "login"), authController.Login)
	auth.Post("/forgot-password/send-otp", authController.ForgotPasswordSendOTP)
    auth.Post("/forgot-password/reset", authController.ResetPassword)
}