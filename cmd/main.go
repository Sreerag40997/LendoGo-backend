package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"lendogo-backend/database"
	"lendogo-backend/internal/app"
	"lendogo-backend/structures/models"
)

func main() {
	// 1. Load the .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or failed to load")
	}

	fiberApp := fiber.New()

	// 2. Enable CORS (Updated for Cookies!)
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true, // 👈 CRITICAL: Must be true for httpOnly cookies to work!
	}))

	// ==========================================
	// 3. DATABASE CONNECTIONS
	// ==========================================
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	if err := database.ConnectRedis(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// ==========================================
	// 4. MIGRATIONS & SEEDING
	// ==========================================
	log.Println("Running Database Migrations...")
	// Tip: You can pass multiple models into AutoMigrate at once!
	// database.DB.AutoMigrate(&models.User{}, &models.Consultation{})
	database.DB.AutoMigrate(&models.User{},
		 &models.Consultation{},
		 &models.LoanApplication{},
		 &models.KYCDocuments{},     
        &models.FinancialDetails{},
		&models.SystemWallet{},
		&models.ChatMessage{},
		&models.UserWallet{},
	    &models.LedgerEntry{},) 
	
	log.Println("Running Seeders...")
	database.SeedAdmin()
	database.RunSeeders()

	// ==========================================
	// 5. WIRING & STARTUP
	// ==========================================
	
	// Call your new app.go hub to wire everything together
	app.SetupApp(fiberApp)

	log.Println("Fiber Server running on port 8080...")
	log.Fatal(fiberApp.Listen(":8080"))
}