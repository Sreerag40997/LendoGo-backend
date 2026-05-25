package controllers

import (
	"lendogo-backend/internal/services"
	"lendogo-backend/structures/dto"

	"github.com/gofiber/fiber/v2"
)

type ConsultationController struct {
    service services.ConsultationService
}

func NewConsultationController(service services.ConsultationService) *ConsultationController {
    return &ConsultationController{service: service}
}

func (cc *ConsultationController) SubmitForm(c *fiber.Ctx) error {
    var req dto.ConsultationReq

    // Parse JSON body
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "error": "Invalid request payload",
        })
    }

    // Call the service
    err := cc.service.RequestConsultation(req)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "error": "Failed to submit consultation request",
        })
    }

    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "message": "Consultation request submitted successfully",
    })
}