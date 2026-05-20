package main

import (
	"log"
	"lendogo-backen/internal/config"
	"lendogo-backen/internal/database"
	"lendogo-backen/internal/routes" // 1. IMPORT YOUR ROUTES PACKAGES!

	"github.com/gofiber/fiber/v2"
)

func main() {
	config.LoadConfig()

	if err := database.Connect(); err != nil {
		log.Fatalf("Server shutting down: %v", err)
	}
	if err := database.ConnectRedis(); err != nil {
		log.Fatalf("Server shutting down (Redis): %v", err)
	}

	app := fiber.New()
	
	// 2. CONNECT THE HUB: Replace the comment with this active line!
	routes.SetupRoutes(app)

	log.Fatal(app.Listen(":8080"))
}