package app

import (
	"github.com/gofiber/fiber/v2"

	"lendogo-backend/database"
	"lendogo-backend/internal/controllers/admin_controller"
	"lendogo-backend/internal/controllers/auth_controller"
	controllers "lendogo-backend/internal/controllers/consultation_controller"
	"lendogo-backend/internal/controllers/loan_controller" 
	"lendogo-backend/internal/repositories"
	"lendogo-backend/internal/routes"
	"lendogo-backend/internal/services"
)

// SetupApp initializes all dependencies and registers routes
func SetupApp(app *fiber.App) {

	// ==========================================
	// 1. REPOSITORIES (Data Layer)
	// ==========================================
	userRepo := repositories.NewUserRepository(database.DB)
	consultationRepo := repositories.NewConsultationRepository(database.DB)
	loanRepo := repositories.NewLoanRepository(database.DB) 
	// ==========================================
	// 2. SERVICES (Business Logic Layer)
	// ==========================================
	authService := services.NewAuthService(userRepo)
	consultationService := services.NewConsultationService(consultationRepo)
	loanService := services.NewLoanService(loanRepo) // 👈 Added Loan Service

	// ==========================================
	// 3. CONTROLLERS (HTTP Layer)
	// ==========================================
	authController := auth_controller.NewAuthController(authService)
	consultationController := controllers.NewConsultationController(consultationService)
	adminController := admin_controller.NewAdminController() 
	loanController := loan_controller.NewLoanController(loanService) 

	// ==========================================
	// 4. ROUTER SETUP
	// ==========================================
	api := app.Group("/api")

	routes.SetupAuthRoutes(api, authController)
	routes.SetupConsultationRoutes(api, consultationController)
	routes.SetupAdminRoutes(api, adminController) 
	routes.SetupLoanRoutes(api, loanController) 
}