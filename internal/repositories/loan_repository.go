package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"lendogo-backend/structures/models"
)

var (
	ErrInsufficientCapital = errors.New("insufficient system capital for disbursement")
	ErrInvalidLoanState    = errors.New("invalid loan state for disbursement")
)

// LoanRepository defines the data access methods for loan applications.
type LoanRepository interface {
	CreateApplication(loan *models.LoanApplication) error
	DisburseLoanTx(loanID uuid.UUID) error
}

type loanRepositoryImpl struct {
	db *gorm.DB
}

// NewLoanRepository returns a new instance of LoanRepository.
func NewLoanRepository(db *gorm.DB) LoanRepository {
	return &loanRepositoryImpl{db: db}
}

// CreateApplication inserts a new loan application into the database.
func (r *loanRepositoryImpl) CreateApplication(loan *models.LoanApplication) error {
	return r.db.Create(loan).Error
}

// DisburseLoanTx executes an atomic, double-entry ledger update transferring principal
// from the master system wallet to the borrower's wallet. It enforces pessimistic 
// row-level locks (SELECT FOR UPDATE) to guarantee isolation during concurrent requests.
func (r *loanRepositoryImpl) DisburseLoanTx(loanID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var loan models.LoanApplication

		// 1. Lock the loan row to prevent race conditions during concurrent approval attempts
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", loanID).
			First(&loan).Error; err != nil {
			return err
		}

		if loan.ApplicationStatus != "PENDING" && loan.ApplicationStatus != "UNDER_REVIEW" {
			return ErrInvalidLoanState
		}

		// 2. Lock the Master Ledger (System Wallet)
		var sysWallet models.SystemWallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("wallet_name = ?", "capital_disbursement").
			First(&sysWallet).Error; err != nil {
			return err
		}

		if sysWallet.Balance < loan.PrincipalAmount {
			return ErrInsufficientCapital
		}

		// 3. Mutate Master Ledger (Debit). Using UpdateColumn to bypass ORM hooks and optimize execution
		if err := tx.Model(&sysWallet).
			UpdateColumn("balance", gorm.Expr("balance - ?", loan.PrincipalAmount)).Error; err != nil {
			return err
		}

		// 4. Lock or Create Borrower Ledger (Credit)
		var userWallet models.UserWallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", loan.UserID).
			FirstOrCreate(&userWallet, models.UserWallet{UserID: loan.UserID, Balance: 0}).Error; err != nil {
			return err
		}

		if err := tx.Model(&userWallet).
			UpdateColumn("balance", gorm.Expr("balance + ?", loan.PrincipalAmount)).Error; err != nil {
			return err
		}

		// 5. Append immutable Double-Entry Receipts to the Ledger table
		entries := []models.LedgerEntry{
			{
				WalletID:        sysWallet.ID,
				Amount:          -loan.PrincipalAmount,
				TransactionType: "DISBURSEMENT_DEBIT",
				ReferenceID:     loan.ID.String(),
			},
			{
				WalletID:        userWallet.ID,
				Amount:          loan.PrincipalAmount,
				TransactionType: "LOAN_CREDIT",
				ReferenceID:     loan.ID.String(),
			},
		}
		if err := tx.Create(&entries).Error; err != nil {
			return err
		}

		// 6. Advance State Machine
		return tx.Model(&loan).UpdateColumn("application_status", "APPROVED").Error
	})
}