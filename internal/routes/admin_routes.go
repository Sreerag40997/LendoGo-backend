package routes

import (
	"github.com/gofiber/fiber/v2"

	// Using the unified controllers package we just fixed!
	controllers "lendogo-backend/internal/controllers/admin_controller"
	"lendogo-backend/internal/middlewares"
)

// SetupAdminRoutes locks down and routes all admin traffic
func SetupAdminRoutes(api fiber.Router, adminCtrl *controllers.AdminController) {
	// 1. Create a specific group for admin features
	adminGroup := api.Group("/admin")

	// 2. Apply BOTH middlewares to everything inside this group!
	// (This guarantees nobody gets in without an Admin JWT)
	adminGroup.Use(middlewares.Protected(), middlewares.AdminOnly())

	// 3. Your existing admin routes
	adminGroup.Get("/all-users", adminCtrl.GetAllUsers)
	adminGroup.Get("/system-stats", adminCtrl.GetSystemStats)

	// ==========================================
	// 4. NEW: Loan Application Management
	// ==========================================
	// Fetches all loans + KYC + Financials in one big payload
	adminGroup.Get("/applications", adminCtrl.GetAllApplications)
	
	// Mutates the status (UNDER_REVIEW -> APPROVED/REJECTED)
	adminGroup.Patch("/applications/:id/status", adminCtrl.UpdateApplicationStatus)

	// Fetches all free consultation requests from database
	adminGroup.Get("/consultations", adminCtrl.GetAllConsultations)
}