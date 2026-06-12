package wallet_controller

import (
	"lendogo-backend/internal/services"
	"lendogo-backend/structures/dto"
	"lendogo-backend/structures/responses"

	"github.com/gofiber/fiber/v2"
)

type WalletController struct {
	service services.WalletService
}

func NewWalletController(service services.WalletService) *WalletController {
	return &WalletController{service: service}
}

// ==========================================
// 1. INBOUND CAPITAL (System Top-ups)
// ==========================================

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

// ==========================================
// 2. OUTBOUND CAPITAL (Loan Disbursal)
// ==========================================

func (c *WalletController) DisburseFunds(ctx *fiber.Ctx) error {
	var req dto.DisburseLoanRequest

	// 1. Parse incoming JSON from the React Admin Panel
	if err := ctx.BodyParser(&req); err != nil {
		return responses.Error(ctx, 400, "Invalid JSON payload")
	}

	// 2. Pass it to the Service layer (which handles math verification & UUID parsing)
	if err := c.service.ProcessDisbursal(req); err != nil {
		return responses.Error(ctx, 400, err.Error()) 
	}

	// 3. Return the Standardized Success Response
	return responses.Success(ctx, 200, "Capital successfully disbursed to user wallet", nil)
}

// ==========================================
// 3. USER WALLET LOGIC
// ==========================================

func (c *WalletController) GetMyBalance(ctx *fiber.Ctx) error {
	// 1. Extract the User ID from the JWT Middleware context
	// NOTE: Make sure your AuthMiddleware uses "user_id" as the key!
	userID, ok := ctx.Locals("user_id").(string)
	if !ok || userID == "" {
		return responses.Error(ctx, 401, "Unauthorized: User ID missing from token")
	}

	// 2. Fetch the real balance from the database
	balance, err := c.service.GetUserBalance(userID)
	if err != nil {
		return responses.Error(ctx, 500, "Failed to fetch wallet balance")
	}

	// 3. Return it to React!
	return responses.Success(ctx, 200, "Balance fetched successfully", fiber.Map{
		"balance": balance,
	})
}