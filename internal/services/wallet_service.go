package services

import (
	"errors"
	"fmt"
	"time"

	"lendogo-backend/internal/repositories"
	"lendogo-backend/utils"
)

type WalletService interface {
	GetBalance() (float64, error)
	GenerateRechargeOrder(amount float64) (map[string]interface{}, error)
	ProcessPaymentVerification(orderID, paymentID, signature string, amount float64) error
		DirectFund(amount float64) error 

}

type walletServiceImpl struct {
	repo repositories.WalletRepository
}


func NewWalletService(repo repositories.WalletRepository) WalletService {
	return &walletServiceImpl{repo: repo}
}

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