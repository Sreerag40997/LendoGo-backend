package jobs

import (
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	
	// Import your services so the job can talk to the database!
	"lendogo-backend/internal/services"
)

type EMICheckerJob struct {
	cron        *cron.Cron
	LoanService services.LoanService
}

func NewEMICheckerJob(loanService services.LoanService) *EMICheckerJob {
	// Create a new cron instance based on the server's local time
	c := cron.New(cron.WithLocation(time.Local))
	return &EMICheckerJob{
		cron:        c,
		LoanService: loanService,
	}
}

func (j *EMICheckerJob) Start() {
	// ⏰ THE CRON EXPRESSION: "0 0 * * *" means run exactly at Midnight (12:00 AM) every day.
	// For testing purposes, we can change this to "* * * * *" to run every 1 minute!
	_, err := j.cron.AddFunc("0 0 * * *", func() {
		fmt.Println("⏰ [CRON WAKING UP] Checking database for due EMIs...")

		// ==========================================
		// TODO: THE LOGIC HAPPENS HERE
		// ==========================================
		// 1. Fetch all pending EMIs due today
		// 2. Send "Payment Due" notifications to the users
		// 3. Mark heavily overdue loans with a late fee

		fmt.Println("🏁 [CRON FINISHED] All EMI reminders sent successfully!")
	})

	if err != nil {
		log.Fatalf("❌ Failed to start EMI Checker Cron Job: %v", err)
	}

	j.cron.Start()
	log.Println("⏱️ EMI Checker Job Scheduled! (Runs daily at Midnight)")
}