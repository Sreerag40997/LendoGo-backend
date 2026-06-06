package services

import (
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
	
	// 👇 NEW: Fetch User Balance
	GetUserBalance(userID string) (float64, error)
}

type walletServiceImpl struct {
	repo repositories.WalletRepository
}

func NewWalletService(repo repositories.WalletRepository) WalletService {
	return &walletServiceImpl{repo: repo}
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
	return s.repo.CreditSystemWallet(amount)
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
	return s.repo.ExecuteDisbursal(loanUUID, userUUID, req.NetPayout)
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