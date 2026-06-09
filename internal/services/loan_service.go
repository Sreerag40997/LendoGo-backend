package services

import (
	"context"
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
	ApproveAndDisburse(loanID uuid.UUID) error
	GenerateEMISchedule(ctx context.Context, loanIDStr string) error
	ProcessDueEMIs(ctx context.Context) error // 👈 Added contract for Cronjob
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
	rand.Seed(time.Now().UnixNano())
	loan.ReferenceNumber = fmt.Sprintf("LG-%06d", rand.Intn(1000000))

	loan.ApplicationStatus = "UNDER_REVIEW"
	loan.InterestRate = 14.0 // 14% Annual Interest

	if loan.TenureMonths > 0 && loan.PrincipalAmount > 0 {
		monthlyRate := (loan.InterestRate / 100) / 12
		tenure := float64(loan.TenureMonths)

		numerator := loan.PrincipalAmount * monthlyRate * math.Pow(1+monthlyRate, tenure)
		denominator := math.Pow(1+monthlyRate, tenure) - 1

		loan.EstimatedEMI = numerator / denominator
	}

	return s.repo.CreateApplication(loan)
}

// ApproveAndDisburse orchestrates the atomic ledger transaction.
func (s *loanServiceImpl) ApproveAndDisburse(loanID uuid.UUID) error {
	return s.repo.DisburseLoanTx(loanID)
}

// ==========================================
// BACKGROUND EMI MATH WORKER
// ==========================================

func (s *loanServiceImpl) GenerateEMISchedule(ctx context.Context, loanIDStr string) error {
	loanUUID, err := uuid.Parse(loanIDStr)
	if err != nil {
		return fmt.Errorf("invalid loan ID format: %v", err)
	}

	loan, err := s.repo.GetLoanByID(loanUUID)
	if err != nil {
		return fmt.Errorf("failed to fetch loan: %v", err)
	}

	monthlyRate := (loan.InterestRate / 100) / 12
	tenure := loan.TenureMonths
	currentPrincipal := loan.PrincipalAmount

	numerator := currentPrincipal * monthlyRate * math.Pow(1+monthlyRate, float64(tenure))
	denominator := math.Pow(1+monthlyRate, float64(tenure)) - 1
	monthlyEMI := numerator / denominator

	var schedules []models.EMISchedule

	for i := 1; i <= tenure; i++ {
		interestForMonth := currentPrincipal * monthlyRate
		principalForMonth := monthlyEMI - interestForMonth
		dueDate := time.Now().AddDate(0, i, 0)

		schedule := models.EMISchedule{
			LoanID:        loanUUID,
			UserID:        loan.UserID,
			InstallmentNo: i,
			DueDate:       dueDate,
			EMI:           math.Round(monthlyEMI*100) / 100,
			PrincipalPart: math.Round(principalForMonth*100) / 100,
			InterestPart:  math.Round(interestForMonth*100) / 100,
			Status:        "PENDING",
		}

		schedules = append(schedules, schedule)
		currentPrincipal -= principalForMonth
	}

	return s.repo.CreateEMISchedules(schedules)
}

// ==========================================
// CRONJOB WORKER
// ==========================================

// ProcessDueEMIs evaluates all pending schedules due on or before today.
func (s *loanServiceImpl) ProcessDueEMIs(ctx context.Context) error {
	today := time.Now()

	emis, err := s.repo.GetDueEMIs(today)
	if err != nil {
		return fmt.Errorf("failed to retrieve due EMIs: %w", err)
	}

	if len(emis) == 0 {
		fmt.Println("✅ [EMI CHECKER] Zero pending EMIs due today.")
		return nil
	}

	fmt.Printf("⚠️ [EMI CHECKER] Processing %d due EMIs.\n", len(emis))

	for _, emi := range emis {
		// Mock notification dispatch. 
		// In production, push to a Kafka topic (e.g., telemetry.notifications) or an SQS queue.
		fmt.Printf("🔔 DISPATCH NOTIFICATION: User %v owes ₹%.2f for Loan %v (Installment %d)\n", 
			emi.UserID, emi.EMI, emi.LoanID, emi.InstallmentNo)
	}

	return nil
}