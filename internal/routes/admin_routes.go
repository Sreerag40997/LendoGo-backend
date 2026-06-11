package routes

import (
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/internal/controllers/admin_controller"
	"lendogo-backend/internal/middlewares"
)

func SetupAdminRoutes(api fiber.Router, adminCtrl *admin_controller.AdminController) {
	adminGroup := api.Group("/admin")
	
	// 1. Base Security: Must be logged in AND be an Admin type
	adminGroup.Use(middlewares.Protected(), middlewares.AdminOnly())

	// ==========================================
	// USER MANAGEMENT (Locked to "user_management")
	// ==========================================
	adminGroup.Get("/all-users", middlewares.RequirePermission("user_management"), adminCtrl.GetAllUsers)
	adminGroup.Post("/users", middlewares.RequirePermission("user_management"), adminCtrl.CreateUser)
	adminGroup.Put("/users/:id", middlewares.RequirePermission("user_management"), adminCtrl.UpdateUser)
	adminGroup.Delete("/users/:id", middlewares.RequirePermission("user_management"), adminCtrl.DeleteUser)
	adminGroup.Patch("/users/:id/status", middlewares.RequirePermission("user_management"), adminCtrl.UpdateUserStatus)
	
	// ==========================================
	// LOAN APPLICATIONS (Locked to "loan_applications")
	// ==========================================
	adminGroup.Get("/applications", middlewares.RequirePermission("loan_applications"), adminCtrl.GetAllApplications)
	adminGroup.Patch("/applications/:id/status", middlewares.RequirePermission("loan_applications"), adminCtrl.UpdateApplicationStatus)

	// ==========================================
	// SYSTEM DASHBOARD (Locked to "dashboard")
	// ==========================================
	adminGroup.Get("/system-stats", middlewares.RequirePermission("dashboard"), adminCtrl.GetSystemStats)

	// ==========================================
	// CUSTOMER CARE (Locked to "customer_care")
	// ==========================================
	adminGroup.Get("/consultations", middlewares.RequirePermission("customer_care"), adminCtrl.GetAllConsultations)

	// ==========================================
	// STAFF MANAGEMENT 
	// ==========================================
	adminGroup.Post("/staff", middlewares.RequirePermission("user_management"), adminCtrl.CreateStaff)
	adminGroup.Get("/staff", middlewares.RequirePermission("user_management"), adminCtrl.GetAllStaff)
}