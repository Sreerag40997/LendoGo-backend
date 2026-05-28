package routes

import (
	controllers "lendogo-backend/internal/controllers/consultation_controller"
	"github.com/gofiber/fiber/v2"
)

func SetupConsultationRoutes(api fiber.Router, controller *controllers.ConsultationController) {
    consultation := api.Group("/consultation")
    consultation.Post("/request", controller.SubmitForm)
}