package routes

import (
	"lendogo-backen/internal/controllers/auth"
	"github.com/gofiber/fiber/v2"
)

func RegisterAuthRoutes(api fiber.Router) {
	authGroup := api.Group("/auth")

	authGroup.Post("/signup", auth.Signup)
	
	// 🚀 ADD THIS LINE RIGHT HERE:
	authGroup.Post("/verify-otp", auth.VerifyOTP)
}