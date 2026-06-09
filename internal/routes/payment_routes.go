package routes

import (
	"github.com/gofiber/fiber/v2"
	
	"lendogo-backend/internal/controllers/payment_controller"
	"lendogo-backend/internal/middlewares" // Import your auth middleware!
)

func SetupPaymentRoutes(router fiber.Router, pc *payment_controller.PaymentController) {
	// Group them under /api/payments and protect them with JWT
	payments := router.Group("/payments", middlewares.Protected())
	
	payments.Post("/order", pc.CreateOrder)
	payments.Post("/verify", pc.VerifyPayment)
}