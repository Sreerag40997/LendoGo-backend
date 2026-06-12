package repositories

import (
	"errors"
	"lendogo-backend/structures/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletRepository interface {
	GetSystemBalance() (float64, error)
	CreditSystemWallet(amount float64) error
	ExecuteDisbursal(loanID uuid.UUID, userID uuid.UUID, netPayout float64) error

	// Fetch User Balance
	GetUserBalance(userID uuid.UUID) (float64, error)
}

type walletRepositoryImpl struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepositoryImpl{db: db}
}

func (r *walletRepositoryImpl) GetSystemBalance() (float64, error) {
	var wallet models.SystemWallet
	if err := r.db.Where("wallet_name = ?", "capital_disbursement").First(&wallet).Error; err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}

func (r *walletRepositoryImpl) CreditSystemWallet(amount float64) error {
	return r.db.Model(&models.SystemWallet{}).
		Where("wallet_name = ?", "capital_disbursement").
		UpdateColumn("balance", gorm.Expr("balance + ?", amount)).Error
}

// ExecuteDisbursal is a strictly locked, ACID-compliant database transaction.
func (r *walletRepositoryImpl) ExecuteDisbursal(loanID uuid.UUID, userID uuid.UUID, netPayout float64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		// 1. Lock and deduct from System Wallet
		var sysWallet models.SystemWallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("wallet_name = ?", "capital_disbursement").First(&sysWallet).Error; err != nil {
			return errors.New("critical error: system wallet not found")
		}
		if sysWallet.Balance < netPayout {
			return errors.New("insufficient system capital to disburse this loan")
		}
		if err := tx.Model(&sysWallet).UpdateColumn("balance", gorm.Expr("balance - ?", netPayout)).Error; err != nil {
			return err
		}

		// 2. Lock and Update Loan Status
		var loan models.LoanApplication
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", loanID).First(&loan).Error; err != nil {
			return errors.New("loan application not found")
		}
		if loan.ApplicationStatus == "DISBURSED" {
			return errors.New("fraud prevention: loan has already been disbursed")
		}
		if err := tx.Model(&loan).Update("application_status", "DISBURSED").Error; err != nil {
			return err
		}

		// 3. Lock and Credit User Wallet (Create if missing)
		var userWallet models.UserWallet
		// 👇 FIX: Use .Limit(1).Find() to silently check for the wallet without triggering GORM logs
		result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).Limit(1).Find(&userWallet)
		if result.Error != nil {
			return result.Error // Real DB connection error
		}

		if result.RowsAffected == 0 {
			// No wallet found! Create it silently.
			userWallet = models.UserWallet{UserID: userID, Balance: 0}
			if err := tx.Create(&userWallet).Error; err != nil {
				return err
			}
		}

		// Add funds safely using DB-level math
		if err := tx.Model(&userWallet).UpdateColumn("balance", gorm.Expr("balance + ?", netPayout)).Error; err != nil {
			return err
		}

		// 4. Create Immutable Ledger Entry
		ledger := models.LedgerEntry{
			WalletID:        userWallet.ID,
			Amount:          netPayout,
			TransactionType: "CREDIT_LOAN_DISBURSEMENT",
			ReferenceID:     loanID.String(),
		}
		if err := tx.Create(&ledger).Error; err != nil {
			return err
		}

		return nil // Commit! All 4 steps succeeded.
	})
}

// ==========================================
// USER WALLET LOGIC
// ==========================================

// GetUserBalance safely fetches the real money balance without spamming logs!
func (r *walletRepositoryImpl) GetUserBalance(userID uuid.UUID) (float64, error) {
	var wallet models.UserWallet
	
	// 👇 FIX: Use .Limit(1).Find() to prevent GORM from logging "record not found"
	result := r.db.Where("user_id = ?", userID).Limit(1).Find(&wallet)
	
	if result.Error != nil {
		return 0, result.Error
	}
	
	if result.RowsAffected == 0 {
		// If they have no wallet row yet, their balance is legally 0.00
		return 0, nil 
	}
	
	return wallet.Balance, nil
}