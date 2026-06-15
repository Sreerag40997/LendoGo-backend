package routes

import (
	controllers "lendogo-backend/internal/controllers/user_profile_controller"
	"lendogo-backend/internal/middlewares"
	"lendogo-backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

func SetupUserProfileRoutes(api fiber.Router, controller *controllers.UserProfileController, configService services.ConfigService) {
	profileGroup := api.Group("/user/profile", middlewares.Protected())
	
	profileGroup.Get("/", controller.GetProfile)
	profileGroup.Put("/", middlewares.RequireFeature(configService, "profile_update"), controller.UpdateProfile) // Use PUT for updates
}