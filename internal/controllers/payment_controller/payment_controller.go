package payment_controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"lendogo-backend/internal/repositories"
	"lendogo-backend/internal/services"
)

type PaymentController struct {
	paymentService services.PaymentService
	paymentRepo    repositories.PaymentRepository
}

func NewPaymentController(ps services.PaymentService, pr repositories.PaymentRepository) *PaymentController {
	return &PaymentController{paymentService: ps, paymentRepo: pr}
}

// CreateOrder is called when the user clicks "Confirm & Pay"
func (c *PaymentController) CreateOrder(ctx *fiber.Ctx) error {
	var req struct {
		Amount float64 `json:"amount"`
		LoanID string  `json:"loan_id"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Generate a short, unique receipt ID
	receiptID := "rcpt_" + uuid.New().String()[:8]

	// Ask Razorpay for the Order ID
	orderID, err := c.paymentService.CreateOrder(req.Amount, receiptID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create Razorpay order: " + err.Error()})
	}

	// Send it back to React so it can open the payment window!
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"order_id": orderID,
		"amount":   req.Amount,
		"currency": "INR",
	})
}

// VerifyPayment is called AFTER the user types their UPI PIN and succeeds
func (c *PaymentController) VerifyPayment(ctx *fiber.Ctx) error {
	// STRICT AUTH: Identify the user securely
	localUID, ok := ctx.Locals("user_id").(string)
	if !ok || localUID == "" {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}
	userID := uuid.MustParse(localUID)

	var req struct {
		RazorpayPaymentID string  `json:"razorpay_payment_id"`
		RazorpayOrderID   string  `json:"razorpay_order_id"`
		RazorpaySignature string  `json:"razorpay_signature"`
		LoanID            string  `json:"loan_id"`
		AmountPaid        float64 `json:"amount_paid"`
	}

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// 1. Vault Security: Verify the mathematical signature so hackers can't fake success
	err := c.paymentService.VerifyPaymentSignature(req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Payment fraud detected: " + err.Error()})
	}

	// 2. Execute the Waterfall Database Transaction
	loanUUID, err := uuid.Parse(req.LoanID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Loan ID"})
	}

	// ✨ SANDBOX BYPASS: If this is the frontend's default mock loan, return success without DB hit
	if loanUUID.String() == "1092a1a1-1234-4321-abcd-1234567890ab" {
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Sandbox Payment verified and simulated successfully!",
			"status":  "success",
		})
	}

	err = c.paymentRepo.ExecuteRepaymentTx(loanUUID, userID, req.AmountPaid, req.RazorpayPaymentID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database update failed: " + err.Error()})
	}

	// 3. Success!
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Payment verified and loan updated successfully!",
		"status":  "success",
	})
}