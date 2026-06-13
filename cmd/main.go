package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"lendogo-backend/config"
	"lendogo-backend/database"
	"lendogo-backend/internal/app"
)

func main() {
	config.LoadConfig()
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to PostgreSQL: ", err)
	}
	if err := database.ConnectRedis(); err != nil {
		log.Fatal("Failed to connect to Redis: ", err)
	}
	log.Println("Running Seeders...")
	database.RunSeeders()
	fiberApp := fiber.New()
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))
	fiberApp.Static("/uploads", "./uploads")
	app.SetupApp(fiberApp)
	port := config.GetEnv("PORT", "8080")
	log.Printf("Fiber Server running on port %s...", port)
	log.Fatal(fiberApp.Listen(":" + port))
}