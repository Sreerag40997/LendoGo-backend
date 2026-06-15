package models

import "time"

// WebConfiguration stores global feature toggles. Only ONE row (ID=1) will ever exist.
type WebConfiguration struct {
	ID uint `gorm:"primaryKey;autoIncrement:false;default:1" json:"-"`

	// Core Features
	ApplyLoanEnabled     bool `gorm:"default:true" json:"apply_loan_enabled"`
	LoginEnabled         bool `gorm:"default:true" json:"login_enabled"`
	RegisterEnabled      bool `gorm:"default:true" json:"register_enabled"`
	ApplyJobEnabled      bool `gorm:"default:true" json:"apply_job_enabled"`
	
	// User Dashboard Features
	ProfileUpdateEnabled bool `gorm:"default:true" json:"profile_update_enabled"`
	FeedbackEnabled      bool `gorm:"default:true" json:"feedback_enabled"`
	LoanHistoryEnabled   bool `gorm:"default:true" json:"loan_history_enabled"`
	RepayEnabled         bool `gorm:"default:true" json:"repay_enabled"`

	// Coming Soon Features (Defaulted to false)
	AutoPayEnabled       bool `gorm:"default:false" json:"auto_pay_enabled"`
	InternalScoreEnabled bool `gorm:"default:false" json:"internal_score_enabled"`
	CibilScoreEnabled    bool `gorm:"default:false" json:"cibil_score_enabled"`

	// Support & Content Features
	BlogEnabled             bool `gorm:"default:true" json:"blog_enabled"`
	ChatSupportEnabled      bool `gorm:"default:true" json:"chat_support_enabled"`
	FreeConsultationEnabled bool `gorm:"default:true" json:"free_consultation_enabled"`

	// Financial Parameters
	MinCreditScore   int     `gorm:"default:650" json:"min_credit_score"`
	BaseInterestRate float64 `gorm:"default:14.0" json:"base_interest_rate"`

	UpdatedAt time.Time `json:"updated_at"`
}