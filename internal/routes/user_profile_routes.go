package routes

import (
	controllers "lendogo-backend/internal/controllers/user_profile_controller"
	"lendogo-backend/internal/middlewares"

	"github.com/gofiber/fiber/v2"
)

func SetupUserProfileRoutes(api fiber.Router, controller *controllers.UserProfileController) {
	profileGroup := api.Group("/user/profile", middlewares.Protected())
	
	profileGroup.Get("/", controller.GetProfile)
	profileGroup.Put("/", controller.UpdateProfile) // Use PUT for updates
}