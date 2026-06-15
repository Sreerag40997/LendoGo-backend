package routes

import (
	"github.com/gofiber/fiber/v2"
	controllers "lendogo-backend/internal/controllers/consultation_controller"
	"lendogo-backend/internal/middlewares"
	"lendogo-backend/internal/services"
)

func SetupConsultationRoutes(api fiber.Router, controller *controllers.ConsultationController, configService services.ConfigService) {
    consultation := api.Group("/consultation", middlewares.RequireFeature(configService, "free_consultation"))
    consultation.Post("/request", controller.SubmitForm)
}