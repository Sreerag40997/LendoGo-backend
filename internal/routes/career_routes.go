package routes

import (
	"github.com/gofiber/fiber/v2"

	"lendogo-backend/internal/controllers/career_controller"
	"lendogo-backend/internal/middlewares"
	"lendogo-backend/internal/services" // 👈 1. ADDED: Import the services package
)

// 👇 2. ADDED: configService services.ConfigService to the parameters
func SetupCareerRoutes(api fiber.Router, careerCtrl *career_controller.CareerController, configService services.ConfigService) {
	careerGroup := api.Group("/careers")

	// ==========================================
	// 🟢 PUBLIC ROUTES (For your React website visitors)
	// ==========================================
	careerGroup.Get("/openings", careerCtrl.GetOpenings)
	careerGroup.Get("/openings/:id", careerCtrl.GetOpeningByID)

	// 👇 3. ADDED: The middleware to check if 'apply_job' is enabled in the database!
	careerGroup.Post(
		"/openings/:id/apply", 
		middlewares.RequireFeature(configService, "apply_job"), // 👈 The Bouncer!
		careerCtrl.SubmitApplication,
	)

	// ==========================================
	// 🔴 PROTECTED ADMIN ROUTES (For HR Staff)
	// ==========================================
	adminCareerGroup := careerGroup.Group("/admin")
	adminCareerGroup.Use(middlewares.Protected(), middlewares.AdminOnly())

	// Lock creating jobs behind a specific HR permission
	adminCareerGroup.Post("/openings", middlewares.RequirePermission("careers.manage"), careerCtrl.CreateOpening)
	adminCareerGroup.Put("/openings/:id", middlewares.RequirePermission("careers.manage"), careerCtrl.UpdateOpening)
	adminCareerGroup.Patch("/openings/:id/status", middlewares.RequirePermission("careers.manage"), careerCtrl.UpdateOpeningStatus)

	// 👇 Routes for managing candidate applications
	adminCareerGroup.Get("/applications", middlewares.RequirePermission("careers.manage"), careerCtrl.GetAllApplications)
	adminCareerGroup.Patch("/applications/:id/status", middlewares.RequirePermission("careers.manage"), careerCtrl.UpdateApplicationStatus)
}