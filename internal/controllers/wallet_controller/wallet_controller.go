package walletcontroller

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/razorpay/razorpay-go"
	"gorm.io/gorm"

	"lendogo-backend/database"
	"lendogo-backend/structures/models"
)

// 1. Get Current Admin Wallet Balance
func GetSystemBalance(c *fiber.Ctx) error {
	var wallet models.SystemWallet
	if err := database.DB.Where("wallet_name = ?", "capital_disbursement").First(&wallet).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Admin wallet not found"})
	}
	return c.JSON(fiber.Map{"balance": wallet.Balance})
}

// 2. Create Razorpay Order (Admin enters amount)
func CreateRazorpayOrder(c *fiber.Ctx) error {
	var req struct {
		Amount float64 `json:"amount"` // The amount the admin typed in!
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid amount"})
	}

	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_SECRET"))
	
	data := map[string]interface{}{
		"amount":   int(req.Amount * 100), // Convert INR to Paise for Razorpay
		"currency": "INR",
		"receipt":  "admin_recharge_tx", 
	}

	body, err := client.Order.Create(data, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create order"})
	}

	return c.JSON(fiber.Map{
		"order_id": body["id"],
		"amount":   body["amount"],
	})
}

// 3. Verify Payment & Add Money to Admin Wallet
func VerifyRazorpayPayment(c *fiber.Ctx) error {
	var req struct {
		PaymentID string  `json:"razorpay_payment_id"`
		OrderID   string  `json:"razorpay_order_id"`
		Signature string  `json:"razorpay_signature"`
		Amount    float64 `json:"amount"` // The amount to add to the DB
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid payload"})
	}

	// Cryptographic Security Check
	secret := os.Getenv("RAZORPAY_SECRET")
	data := req.OrderID + "|" + req.PaymentID
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if subtle.ConstantTimeCompare([]byte(expectedSignature), []byte(req.Signature)) != 1 {
		return c.Status(401).JSON(fiber.Map{"error": "Fake payment detected!"})
	}

	// ADD MONEY TO THE ADMIN WALLET IN POSTGRES
	result := database.DB.Model(&models.SystemWallet{}).
		Where("wallet_name = ?", "capital_disbursement").
		UpdateColumn("balance", gorm.Expr("balance + ?", req.Amount))

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Database update failed"})
	}

	return c.JSON(fiber.Map{"message": "Admin Wallet recharged successfully!"})
}