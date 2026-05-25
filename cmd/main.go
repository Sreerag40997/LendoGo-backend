package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"lendogo-backend/database"
	"lendogo-backend/internal/controllers/auth_controller"
	controllers "lendogo-backend/internal/controllers/consultation_controller"
	"lendogo-backend/structures/models"

	// === NEW: Import your consultation controllers and models ===
	// ==========================================================

	"lendogo-backend/internal/repositories"
	"lendogo-backend/internal/routes"
	"lendogo-backend/internal/services"
)

func main() {
    // 1. Load the .env file before doing ANYTHING else
    if err := godotenv.Load(); err != nil {
        log.Println("Warning: No .env file found or failed to load")
    }

    app := fiber.New()

    // 2. Enable CORS
    app.Use(cors.New(cors.Config{
        AllowOrigins: "http://localhost:5173",
        AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    }))

    // ==========================================
    // 3. DATABASE CONNECTIONS
    // ==========================================
    
    // Connect to PostgreSQL
    err := database.Connect()
    if err != nil {
        log.Fatal("Failed to connect to PostgreSQL:", err)
    }

    // === NEW: Run the migration to create the table in Postgres ===
    log.Println("Running Database Migrations...")
    database.DB.AutoMigrate(&models.Consultation{})
    // ==============================================================

    // Connect to Redis
    err = database.ConnectRedis()
    if err != nil {
        log.Fatal("Failed to connect to Redis:", err)
    }

    // ==========================================
    // 4. DEPENDENCY INJECTION WIRING
    // ==========================================
    
    // --- Auth Feature Wiring ---
    userRepo := repositories.NewUserRepository(database.DB)
    authService := services.NewAuthService(userRepo)
    authController := auth_controller.NewAuthController(authService)

    // === NEW: Consultation Feature Wiring ===
    // Step A: Give the Global Database to the Consultation Repository
    consultationRepo := repositories.NewConsultationRepository(database.DB)
    
    // Step B: Give the Repository to the Consultation Service
    consultationService := services.NewConsultationService(consultationRepo)
    
    // Step C: Give the Service to the Consultation Controller
    consultationController := controllers.NewConsultationController(consultationService)
    // ========================================

    // ==========================================
    // 5. Setup Routes
    // ==========================================
    api := app.Group("/api")
    
    // Auth Routes
    routes.SetupAuthRoutes(api, authController)

    // === NEW: Consultation Routes ===
    routes.SetupConsultationRoutes(api, consultationController)
    // ================================

    // 6. Start the Server
    log.Println("🚀 Fiber Server running on port 8080...")
    log.Fatal(app.Listen(":8080"))
}