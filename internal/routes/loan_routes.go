package routes

import (
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/internal/controllers/loan_controller"
	"lendogo-backend/internal/middlewares"
)

func SetupLoanRoutes(api fiber.Router, loanCtrl *loan_controller.LoanController) {
	// 1. Create a specific group for loan features
	loanGroup := api.Group("/loans")

	// 2. Protect the entire group! (Must be logged in to apply)
	loanGroup.Use(middlewares.Protected())

	// 3. The endpoint React will hit with the multipart/form-data
	// This becomes: POST http://localhost:8080/api/loans/apply
	loanGroup.Post("/apply", loanCtrl.ApplyForLoan)
}