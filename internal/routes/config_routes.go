package routes

import (
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/internal/controllers/config_controller"
	"lendogo-backend/internal/middlewares"
)

func SetupConfigRoutes(api fiber.Router, configCtrl *config_controller.ConfigController) {
	configGroup := api.Group("/config")

	// 🟢 PUBLIC: React fetches this once when the user opens the website
	configGroup.Get("/", configCtrl.GetPublicConfig)

	// 🔴 PROTECTED: Admin panel toggles
	adminConfigGroup := configGroup.Group("/admin")
	adminConfigGroup.Use(middlewares.Protected(), middlewares.AdminOnly())
	
	// Assuming you want high-level admins only to change this
	adminConfigGroup.Put("/", middlewares.RequirePermission("system.manage"), configCtrl.UpdateConfig)
}