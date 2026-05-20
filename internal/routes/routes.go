package routes

import "github.com/gofiber/fiber/v2"

// SetupRoutes acts as the master air traffic controller
func SetupRoutes(app *fiber.App) {
	// Base API Version Group
	api := app.Group("/api")

	// Hand off specific features to their dedicated routing files!
	RegisterAuthRoutes(api)

	// When you build loan management later, it cleanly maps like this:
	// RegisterLoanRoutes(api) 
}