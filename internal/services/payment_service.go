package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"

	razorpay "github.com/razorpay/razorpay-go"
)

type PaymentService interface {
	CreateOrder(amount float64, receiptID string) (string, error)
	VerifyPaymentSignature(orderID, paymentID, signature string) error
}

type paymentServiceImpl struct {
	client *razorpay.Client
	secret string
}

func NewPaymentService() PaymentService {
	keyID := os.Getenv("RAZORPAY_KEY_ID")
	keySecret := os.Getenv("RAZORPAY_SECRET")

	client := razorpay.NewClient(keyID, keySecret)

	return &paymentServiceImpl{
		client: client,
		secret: keySecret,
	}
}

// CreateOrder asks Razorpay for a new Order ID. 
func (s *paymentServiceImpl) CreateOrder(amount float64, receiptID string) (string, error) {
	// Razorpay STRICTLY requires the amount in the smallest currency sub-unit (Paise)
	// Example: ₹1532.00 becomes 153200 paise
	amountInPaise := int(amount * 100)

	data := map[string]interface{}{
		"amount":   amountInPaise,
		"currency": "INR",
		"receipt":  receiptID,
	}

	body, err := s.client.Order.Create(data, nil)
	if err != nil {
		return "", err
	}

	orderID, ok := body["id"].(string)
	if !ok {
		return "", errors.New("failed to extract order ID from Razorpay response")
	}

	return orderID, nil
}

// VerifyPaymentSignature acts as our Vault Security. It mathematically proves the frontend wasn't hacked.
func (s *paymentServiceImpl) VerifyPaymentSignature(orderID, paymentID, signature string) error {
	data := orderID + "|" + paymentID

	// Create a SHA256 HMAC hash using your secret key
	h := hmac.New(sha256.New, []byte(s.secret))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Compare our calculated hash with the signature Razorpay sent to the frontend
	if expectedSignature != signature {
		return errors.New("invalid payment signature: potential fraud detected")
	}

	return nil
}