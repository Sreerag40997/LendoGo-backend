package wallet_controller

import (
	"github.com/gofiber/fiber/v2"
	"lendogo-backend/internal/services"
)

type WalletController struct {
	service services.WalletService
}

func NewWalletController(service services.WalletService) *WalletController {
	return &WalletController{service: service}
}

func (c *WalletController) GetSystemBalance(ctx *fiber.Ctx) error {
	balance, err := c.service.GetBalance()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch balance"})
	}
	return ctx.JSON(fiber.Map{"balance": balance})
}

func (c *WalletController) CreateRazorpayOrder(ctx *fiber.Ctx) error {
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	orderData, err := c.service.GenerateRechargeOrder(req.Amount)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(fiber.Map{
		"order_id": orderData["id"],
		"amount":   orderData["amount"],
	})
}

func (c *WalletController) VerifyRazorpayPayment(ctx *fiber.Ctx) error {
	var req struct {
		PaymentID string  `json:"razorpay_payment_id"`
		OrderID   string  `json:"razorpay_order_id"`
		Signature string  `json:"razorpay_signature"`
		Amount    float64 `json:"amount"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	err := c.service.ProcessPaymentVerification(req.OrderID, req.PaymentID, req.Signature, req.Amount)
	if err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.JSON(fiber.Map{"message": "Wallet recharged successfully!"})
}
func (c *WalletController) DeveloperCheatFund(ctx *fiber.Ctx) error {
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	// Use our new direct funding service
	err := c.service.DirectFund(req.Amount)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to inject funds"})
	}

	return ctx.JSON(fiber.Map{"message": "💰 God Mode Activated: Funds injected successfully!"})
}	