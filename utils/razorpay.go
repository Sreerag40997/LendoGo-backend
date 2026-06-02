package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	// "fmt"
	"os"

	"github.com/razorpay/razorpay-go"
)

// CreateRazorpayOrder creates a new order in Razorpay (converts Rupees to Paise)
func CreateRazorpayOrder(amountInRupees float64, receiptID string) (map[string]interface{}, error) {
	keyID := os.Getenv("RAZORPAY_KEY_ID")
	secret := os.Getenv("RAZORPAY_SECRET")
	client := razorpay.NewClient(keyID, secret)

	data := map[string]interface{}{
		"amount":   int(amountInRupees * 100),
		"currency": "INR",
		"receipt":  receiptID,
	}

	return client.Order.Create(data, nil)
}

// VerifyRazorpaySignature performs the cryptographic check
func VerifyRazorpaySignature(orderID, paymentID, signature string) bool {
	secret := os.Getenv("RAZORPAY_SECRET")
	data := orderID + "|" + paymentID

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(data))
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return subtle.ConstantTimeCompare([]byte(expectedSignature), []byte(signature)) == 1
}