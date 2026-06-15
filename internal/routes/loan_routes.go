package routes

import (
	"github.com/gofiber/fiber/v2"
	
	"lendogo-backend/internal/controllers/loan_controller"
	"lendogo-backend/internal/middlewares"
	"lendogo-backend/internal/services" // 👈 1. ADDED: Import the services package
)

// 👇 2. ADDED: configService services.ConfigService to the parameters
func SetupLoanRoutes(api fiber.Router, loanCtrl *loan_controller.LoanController, configService services.ConfigService) {
	// 1. Create a specific group for loan features
	loanGroup := api.Group("/loans")

	// 2. Protect the entire group! (Must be logged in to access loan features)
	loanGroup.Use(middlewares.Protected())

	// 3. The endpoint React will hit with the multipart/form-data
	// 👇 APPLIED MIDDLEWARE: Blocks applications if the Admin turns the "apply_loan" feature off!
	loanGroup.Post(
		"/apply", 
		middlewares.RequireFeature(configService, "apply_loan"), // 👈 The Bouncer!
		loanCtrl.ApplyForLoan,
	)
}