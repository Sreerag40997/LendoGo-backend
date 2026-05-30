package services

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"lendogo-backend/internal/repositories"
	"lendogo-backend/structures/models"
)

// LoanService defines the business logic interface for loan operations.
type LoanService interface {
	ProcessApplication(loan *models.LoanApplication) error
	ApproveAndDisburse(loanID uuid.UUID) error // Added the new disbursement contract
}

type loanServiceImpl struct {
	repo repositories.LoanRepository
}

// NewLoanService injects the repository dependency.
func NewLoanService(repo repositories.LoanRepository) LoanService {
	return &loanServiceImpl{repo: repo}
}

// ProcessApplication calculates EMI and initializes the loan state.
func (s *loanServiceImpl) ProcessApplication(loan *models.LoanApplication) error {
	// 1. Generate a random Reference Number (e.g., LG-807191)
	rand.Seed(time.Now().UnixNano())
	loan.ReferenceNumber = fmt.Sprintf("LG-%06d", rand.Intn(1000000))

	// 2. Set default status and interest rate
	loan.ApplicationStatus = "UNDER_REVIEW"
	loan.InterestRate = 14.0 // 14% Annual Interest

	// 3. Calculate Estimated EMI using Standard Formula
	if loan.TenureMonths > 0 && loan.PrincipalAmount > 0 {
		monthlyRate := (loan.InterestRate / 100) / 12
		tenure := float64(loan.TenureMonths)

		// EMI = [P x R x (1+R)^N] / [(1+R)^N-1]
		numerator := loan.PrincipalAmount * monthlyRate * math.Pow(1+monthlyRate, tenure)
		denominator := math.Pow(1+monthlyRate, tenure) - 1

		loan.EstimatedEMI = numerator / denominator
	}

	// 4. Save to Database
	return s.repo.CreateApplication(loan)
}

// ApproveAndDisburse orchestrates the atomic ledger transaction.
func (s *loanServiceImpl) ApproveAndDisburse(loanID uuid.UUID) error {
	// Delegate the atomic transaction to the data access layer
	return s.repo.DisburseLoanTx(loanID)
}