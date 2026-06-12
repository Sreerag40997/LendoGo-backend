package services

import (
	"context" // 👈 Added for Kafka background context
	"errors"
	"fmt"
	"math"
	"time"

	"lendogo-backend/internal/repositories"
	"lendogo-backend/structures/dto"
	"lendogo-backend/utils"

	"github.com/google/uuid"
)

type WalletService interface {
	GetBalance() (float64, error)
	GenerateRechargeOrder(amount float64) (map[string]interface{}, error)
	ProcessPaymentVerification(orderID, paymentID, signature string, amount float64) error
	DirectFund(amount float64) error

	// Loan Disbursal
	ProcessDisbursal(req dto.DisburseLoanRequest) error

	// Fetch User Balance
	GetUserBalance(userID string) (float64, error)
}

// 👇 1. ADDED THE PRODUCER TO THE STRUCT
type walletServiceImpl struct {
	repo     repositories.WalletRepository
	producer *utils.KafkaProducer
}

// 👇 2. UPDATED THE CONSTRUCTOR TO REQUIRE THE PRODUCER
func NewWalletService(repo repositories.WalletRepository, producer *utils.KafkaProducer) WalletService {
	return &walletServiceImpl{
		repo:     repo,
		producer: producer,
	}
}

// ==========================================
// 1. INBOUND CAPITAL (System Wallet Top-ups)
// ==========================================

func (s *walletServiceImpl) GetBalance() (float64, error) {
	return s.repo.GetSystemBalance()
}

func (s *walletServiceImpl) GenerateRechargeOrder(amount float64) (map[string]interface{}, error) {
	receipt := fmt.Sprintf("admin_rx_%d", time.Now().Unix())
	return utils.CreateRazorpayOrder(amount, receipt)
}

func (s *walletServiceImpl) ProcessPaymentVerification(orderID, paymentID, signature string, amount float64) error {
	// 1. Verify Crypto Signature using Utility
	isValid := utils.VerifyRazorpaySignature(orderID, paymentID, signature)
	if !isValid {
		return errors.New("invalid payment signature")
	}

	// 2. Add Money using Repository
	err := s.repo.CreditSystemWallet(amount)
	if err != nil {
		return err // If DB fails, don't send the Kafka event!
	}

	// 👇 3. KAFKA TRIGGER: BLAST THE EVENT AFTER SUCCESSFUL RECHARGE! 👇
	payload := map[string]interface{}{
		"type":           "SYSTEM_RECHARGE",
		"transaction_id": paymentID,
		"order_id":       orderID,
		"amount":         amount,
		"status":         "COMPLETED",
	}

	// Publish the event to the background workers (Takes ~2 milliseconds)
	kafkaErr := s.producer.PublishEvent(context.Background(), "telemetry.payments", "SYSTEM_RECHARGE_SUCCESS", payload)
	if kafkaErr != nil {
		fmt.Println("❌ KAFKA ERROR: Failed to publish system recharge event:", kafkaErr)
		// We don't return the error here because the payment actually succeeded in the DB.
		// We just log it so the admin still gets a success response.
	} else {
		fmt.Println("🚀 KAFKA EVENT PUBLISHED: SYSTEM_RECHARGE_SUCCESS")
	}
	// 👆 END OF KAFKA TRIGGER 👆

	return nil
}

func (s *walletServiceImpl) DirectFund(amount float64) error {
	// Bypasses Razorpay completely and just talks to the database
	return s.repo.CreditSystemWallet(amount)
}

// ==========================================
// 2. OUTBOUND CAPITAL (Loan Disbursements)
// ==========================================

func (s *walletServiceImpl) ProcessDisbursal(req dto.DisburseLoanRequest) error {
	// 1. Security Math Verification: Never trust the frontend math blindly!
	expectedNet := req.SanctionedAmt - req.ProcessingFee

	// FinTech Rule: Use math.Abs to prevent floating-point mismatch errors
	if math.Abs(req.NetPayout-expectedNet) > 0.01 {
		return errors.New("security alert: payout amount mismatch detected")
	}

	// 2. Parse String IDs to Secure UUIDs
	loanUUID, err := uuid.Parse(req.LoanID)
	if err != nil {
		return errors.New("invalid loan ID format")
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	// 3. Delegate to Repository for the ACID Transaction execution
	// 👇 WE CATCH THE ERROR HERE FIRST! 👇
	err = s.repo.ExecuteDisbursal(loanUUID, userUUID, req.NetPayout)
	if err != nil {
		return err // If the database fails or rolls back, WE STOP HERE. No Kafka event is sent!
	}

	// 👇 4. KAFKA TRIGGER: BLAST THE EVENT AFTER SUCCESSFUL DISBURSAL! 👇
	payload := map[string]interface{}{
		"loan_id":   req.LoanID,
		"user_id":   req.UserID,
		"amount":    req.NetPayout,
		"status":    "DISBURSED",
		"timestamp": time.Now().Unix(),
	}

	kafkaErr := s.producer.PublishEvent(context.Background(), "telemetry.loans", "LOAN_DISBURSED", payload)
	if kafkaErr != nil {
		fmt.Println("❌ KAFKA ERROR: Failed to publish loan disbursal event:", kafkaErr)
	} else {
		fmt.Println("🚀 KAFKA EVENT PUBLISHED: LOAN_DISBURSED")
	}
	// 👆 END OF KAFKA TRIGGER 👆

	return nil // Return success!
}

// ==========================================
// 3. USER WALLET LOGIC
// ==========================================

// GetUserBalance safely parses the string ID from the JWT token and fetches the balance
func (s *walletServiceImpl) GetUserBalance(userID string) (float64, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return 0, errors.New("invalid user ID format")
	}

	return s.repo.GetUserBalance(userUUID)
}