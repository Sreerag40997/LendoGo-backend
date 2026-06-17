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
	ProcessDueEMIs(ctx context.Context) error 

	// Frontend API Services
	GetMyLoans(userID string) ([]models.LoanApplication, error)
	GetMyRepayments(loanID string) ([]models.EMISchedule, error)

	// 👇 NEW: Kafka Payment Event Services
	MarkEMIPaid(ctx context.Context, scheduleID string) error
	ApplyPenalty(ctx context.Context, loanID string) error
}

type loanServiceImpl struct {
	repo         repositories.LoanRepository
	notifService NotificationService
}

// NewLoanService injects the repository dependency.
func NewLoanService(repo repositories.LoanRepository, notifService NotificationService) LoanService {
	return &loanServiceImpl{repo: repo, notifService: notifService}
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
		// Actually mark the EMI as overdue in the database!
		err := s.repo.UpdateEMIStatus(emi.ID, "OVERDUE")
		if err != nil {
			fmt.Printf("❌ Failed to update EMI status for %v: %v\n", emi.ID, err)
		} else {
			fmt.Printf("🔔 DISPATCH NOTIFICATION: User %v owes ₹%.2f for Loan %v (Installment %d) -> Marked OVERDUE\n",
				emi.UserID, emi.EMI, emi.LoanID, emi.InstallmentNo)

			// Dispatch In-App Notification
			notifMsg := fmt.Sprintf("Repayment due: EMI of ₹%.2f on %s.", emi.EMI, emi.DueDate.Format("Jan 02, 2006"))
			s.notifService.SendNotification(emi.UserID, notifMsg, "WARNING", "repayment")
		}
	}

	return nil
}

// ==========================================
// FRONTEND API SERVICES
// ==========================================

func (s *loanServiceImpl) GetMyLoans(userID string) ([]models.LoanApplication, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %v", err)
	}
	return s.repo.GetLoansByUserID(userUUID)
}

func (s *loanServiceImpl) GetMyRepayments(loanID string) ([]models.EMISchedule, error) {
	loanUUID, err := uuid.Parse(loanID)
	if err != nil {
		return nil, fmt.Errorf("invalid loan ID: %v", err)
	}
	return s.repo.GetEMIsByLoanID(loanUUID)
}

// ==========================================
// 👇 KAFKA PAYMENT EVENT SERVICES (NEW) 👇
// ==========================================

func (s *loanServiceImpl) MarkEMIPaid(ctx context.Context, scheduleID string) error {
	return s.repo.MarkEMIPaid(ctx, scheduleID)
}

func (s *loanServiceImpl) ApplyPenalty(ctx context.Context, loanID string) error {
	return s.repo.ApplyPenalty(ctx, loanID)
}