package repositories

import (
	"context"
	"errors"
	"time"

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

	// Background Worker Queries
	GetLoanByID(loanID uuid.UUID) (*models.LoanApplication, error)
	CreateEMISchedules(schedules []models.EMISchedule) error

	// Cronjob Queries
	GetDueEMIs(today time.Time) ([]models.EMISchedule, error)
	UpdateEMIStatus(emiID uuid.UUID, status string) error

	// Frontend API Queries
	GetLoansByUserID(userID uuid.UUID) ([]models.LoanApplication, error)
	GetEMIsByLoanID(loanID uuid.UUID) ([]models.EMISchedule, error)

	// Kafka Payment Event Queries
	MarkEMIPaid(ctx context.Context, scheduleID string) error
	ApplyPenalty(ctx context.Context, loanID string) error
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

		// 3. Mutate Master Ledger (Debit)
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

// ==========================================
// EMI BACKGROUND WORKER QUERIES
// ==========================================

func (r *loanRepositoryImpl) GetLoanByID(loanID uuid.UUID) (*models.LoanApplication, error) {
	var loan models.LoanApplication
	if err := r.db.Where("id = ?", loanID).First(&loan).Error; err != nil {
		return nil, err
	}
	return &loan, nil
}

func (r *loanRepositoryImpl) CreateEMISchedules(schedules []models.EMISchedule) error {
	return r.db.Create(&schedules).Error
}

// ==========================================
// CRONJOB EMI CHECKER QUERIES
// ==========================================

// GetDueEMIs finds all unpaid EMIs where the due date is today or earlier
func (r *loanRepositoryImpl) GetDueEMIs(today time.Time) ([]models.EMISchedule, error) {
	var emis []models.EMISchedule
	err := r.db.Where("status = ? AND due_date <= ?", "PENDING", today).Find(&emis).Error
	return emis, err
}

// UpdateEMIStatus changes an EMI from PENDING to PAID or OVERDUE
func (r *loanRepositoryImpl) UpdateEMIStatus(emiID uuid.UUID, status string) error {
	return r.db.Model(&models.EMISchedule{}).Where("id = ?", emiID).Update("status", status).Error
}

// ==========================================
// KAFKA PAYMENT EVENT QUERIES (NEW)
// ==========================================

// MarkEMIPaid updates a specific EMI schedule row to 'Paid' and sets the payment date
func (r *loanRepositoryImpl) MarkEMIPaid(ctx context.Context, scheduleID string) error {
	return r.db.WithContext(ctx).
		Model(&models.EMISchedule{}).
		Where("id = ?", scheduleID).
		Updates(map[string]interface{}{
			"status":    "Paid",
			"paid_date": time.Now(), // Assuming you have or will want a paid_date column
		}).Error
}

// ApplyPenalty updates the loan status when an EMI is missed.
// In a full enterprise system, you would also use gorm.Expr to add a fee to a "PenaltyAmount" column.
func (r *loanRepositoryImpl) ApplyPenalty(ctx context.Context, loanID string) error {
	return r.db.WithContext(ctx).
		Model(&models.LoanApplication{}).
		Where("id = ?", loanID).
		Update("application_status", "DEFAULTED").Error // Assuming standard status is DEFAULTED
}

// ==========================================
// FRONTEND API QUERIES
// ==========================================

func (r *loanRepositoryImpl) GetLoansByUserID(userID uuid.UUID) ([]models.LoanApplication, error) {
	var loans []models.LoanApplication
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Find(&loans).Error
	return loans, err
}

func (r *loanRepositoryImpl) GetEMIsByLoanID(loanID uuid.UUID) ([]models.EMISchedule, error) {
	var emis []models.EMISchedule
	err := r.db.Where("loan_id = ?", loanID).Order("installment_no asc").Find(&emis).Error
	return emis, err
}