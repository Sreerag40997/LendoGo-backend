package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 1. Core Application (Contains all the basic text data and loan terms)
type LoanApplication struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID `gorm:"type:uuid;not null"`
	ReferenceNumber string    `gorm:"type:varchar(20);uniqueIndex"`

	// Applicant Details
	FullName     string `gorm:"type:varchar(100)"`
	DOB          string `gorm:"type:varchar(20)"`
	Email        string `gorm:"type:varchar(100)"`
	MobileNumber string `gorm:"type:varchar(20)"`
	Address      string
	City         string
	State        string
	Pincode      string `gorm:"type:varchar(20)"`

	// Loan Configuration
	LoanTrack       string  `gorm:"type:varchar(50)"`
	ProductCategory string  `gorm:"type:varchar(50)"`
	PrincipalAmount float64 `gorm:"not null"`
	TenureMonths    int     `gorm:"not null"`
	InterestRate    float64
	EstimatedEMI    float64

	ApplicationStatus string `gorm:"type:varchar(20);default:'UNDER_REVIEW'"`

	// 👇 GORM MAGIC: These fields tell GORM to link the other two tables! 👇
	KYC           KYCDocuments     `gorm:"foreignKey:LoanID;constraint:OnDelete:CASCADE;" json:"kyc_documents"`
	FinancialDocs FinancialDetails `gorm:"foreignKey:LoanID;constraint:OnDelete:CASCADE;" json:"financial_details"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 2. KYC Table (Strictly Identity & Security)
type KYCDocuments struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"-"`
	LoanID           uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"loan_id"` // uniqueIndex forces 1:1 Match
	LiveSelfiePath   string    `json:"live_selfie_path"`
	AadhaarFrontPath string    `json:"aadhaar_front_path"`
	AadhaarBackPath  string    `json:"aadhaar_back_path"`
	PanCardPath      string    `json:"pan_card_path"`
}

// 3. Financial Table (Strictly Income & Bank files)
type FinancialDetails struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey" json:"-"`
	LoanID                uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"loan_id"` // uniqueIndex forces 1:1 Match
	EmploymentStatus      string    `gorm:"type:varchar(50)" json:"employment_status"`
	MonthlyIncome         float64   `json:"monthly_income"`
	BankStatementPath     string    `json:"bank_statement_path"`
	PropertyAgreemntPath  string    `json:"property_agreemnt_path"`
	IncomeProofPath       string    `json:"income_proof_path"`
}

// ==========================================
// UUID Generation Hooks for all 3 tables
// ==========================================

func (l *LoanApplication) BeforeCreate(tx *gorm.DB) (err error) {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return
}

func (k *KYCDocuments) BeforeCreate(tx *gorm.DB) (err error) {
	if k.ID == uuid.Nil {
		k.ID = uuid.New()
	}
	return
}

func (f *FinancialDetails) BeforeCreate(tx *gorm.DB) (err error) {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return
}