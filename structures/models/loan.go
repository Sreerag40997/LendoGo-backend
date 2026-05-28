package models

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoanApplication represents the massive form the user submits
type LoanApplication struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID `gorm:"type:uuid;not null"` // Links to the logged-in user
	ReferenceNumber string    `gorm:"type:varchar(20);uniqueIndex"` // e.g., "LG-807191"

	// 1. Applicant Details (From Step 1)
	FullName     string `gorm:"type:varchar(100)"`
	DOB          string `gorm:"type:varchar(20)"`
	Email        string `gorm:"type:varchar(100)"`
	MobileNumber string `gorm:"type:varchar(20)"`
	Address      string
	City         string
	State        string
	Pincode      string `gorm:"type:varchar(20)"`

	// 2. Loan Configuration (From Step 2)
	LoanTrack       string  `gorm:"type:varchar(50)"` // "Micro-Credit Hub" or "Elite Asset Funding"
	ProductCategory string  `gorm:"type:varchar(50)"` // "Instant Personal Loan", etc.
	PrincipalAmount float64 `gorm:"not null"`
	TenureMonths    int     `gorm:"not null"`
	InterestRate    float64 //
	EstimatedEMI    float64

	// 3. KYC & Employment (From Step 3)
	EmploymentStatus string  `gorm:"type:varchar(50)"`
	MonthlyIncome    float64

	// 4. Document Links (These will store the AWS S3 URLs)
	LiveSelfiePath       string
	AadhaarFrontPath     string
	AadhaarBackPath      string
	PanCardPath          string
	BankStatementPath    string // Optional for low loans
	PropertyAgreemntPath string // Optional for low loans
	IncomeProofPath      string // Optional for low loans

	// 5. App Status
	ApplicationStatus string `gorm:"type:varchar(20);default:'UNDER_REVIEW'"` 
	
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Generate the UUID before saving to the database
func (l *LoanApplication) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return
}