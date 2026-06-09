package repositories

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"lendogo-backend/structures/models"
)

type PaymentRepository interface {
	ExecuteRepaymentTx(loanID uuid.UUID, userID uuid.UUID, amountPaid float64, razorpayPaymentID string) error
}

type paymentRepositoryImpl struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepositoryImpl{db: db}
}

// ExecuteRepaymentTx handles the massive "Waterfall" payment algorithm
func (r *paymentRepositoryImpl) ExecuteRepaymentTx(loanID uuid.UUID, userID uuid.UUID, amountPaid float64, razorpayPaymentID string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		// 1. Lock and Update the Admin System Wallet (The business receives the money!)
		var sysWallet models.SystemWallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("wallet_name = ?", "capital_disbursement").First(&sysWallet).Error; err != nil {
			return errors.New("critical error: admin system wallet not found")
		}
		if err := tx.Model(&sysWallet).UpdateColumn("balance", gorm.Expr("balance + ?", amountPaid)).Error; err != nil {
			return err
		}

		// 2. Create the Ledger Entry (Payment History for the Admin Dashboard)
		ledger := models.LedgerEntry{
			WalletID:        sysWallet.ID,
			Amount:          amountPaid,
			TransactionType: "LOAN_REPAYMENT",
			ReferenceID:     razorpayPaymentID, // We save the exact Razorpay ID as proof!
		}
		if err := tx.Create(&ledger).Error; err != nil {
			return err
		}

		// 3. Lock the Loan Application
		var loan models.LoanApplication
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", loanID).First(&loan).Error; err != nil {
			return errors.New("loan not found")
		}

		// 4. Fetch and Lock all PENDING EMIs, ordered by oldest first
		var pendingEMIs []models.EMISchedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("loan_id = ? AND status = ?", loanID, "PENDING").
			Order("installment_no asc").Find(&pendingEMIs).Error; err != nil {
			return err
		}

		// 5. 🌊 THE WATERFALL ALGORITHM 🌊
		// Distribute the user's money across the EMIs
		remainingAmount := amountPaid

		for i := range pendingEMIs {
			if remainingAmount <= 0 {
				break // No money left to distribute
			}

			if remainingAmount >= pendingEMIs[i].EMI {
				// The user paid enough to completely clear this EMI month
				pendingEMIs[i].Status = "PAID"
				remainingAmount -= pendingEMIs[i].EMI
			} else {
				// The user paid a CUSTOM partial amount. 
				// We reduce what they owe for this month, but it stays PENDING!
				pendingEMIs[i].EMI -= remainingAmount
				remainingAmount = 0
			}

			// Save the updated EMI row
			if err := tx.Save(&pendingEMIs[i]).Error; err != nil {
				return err
			}
		}

		// 6. Did they pay off the Entire Loan?
		var remainingPending int64
		tx.Model(&models.EMISchedule{}).Where("loan_id = ? AND status = ?", loanID, "PENDING").Count(&remainingPending)

		// If 0 EMIs are pending, the loan is officially over!
		if remainingPending == 0 {
			
			// Close the Loan Application
			if err := tx.Model(&loan).UpdateColumn("application_status", "CLOSED").Error; err != nil {
				return err
			}

			// 🏆 Add +50 points to their Trust Score so they can borrow again!
			var profile models.UserProfile
			result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).Limit(1).Find(&profile)
			if result.Error == nil && result.RowsAffected > 0 {
				tx.Model(&profile).UpdateColumn("trust_score", gorm.Expr("trust_score + ?", 50))
			}
		}

		return nil // Commit the entire transaction!
	})
}