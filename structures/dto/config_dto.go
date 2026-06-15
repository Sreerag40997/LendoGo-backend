package dto

// UpdateConfigReq is the payload expected from the Admin Panel
type UpdateConfigReq struct {
	ApplyLoanEnabled     bool `json:"apply_loan_enabled"`
	LoginEnabled         bool `json:"login_enabled"`
	RegisterEnabled      bool `json:"register_enabled"`
	ApplyJobEnabled      bool `json:"apply_job_enabled"`
	ProfileUpdateEnabled bool `json:"profile_update_enabled"`
	FeedbackEnabled      bool `json:"feedback_enabled"`
	LoanHistoryEnabled   bool `json:"loan_history_enabled"`
	RepayEnabled         bool `json:"repay_enabled"`
	AutoPayEnabled          bool `json:"auto_pay_enabled"`
	InternalScoreEnabled    bool `json:"internal_score_enabled"`
	CibilScoreEnabled       bool `json:"cibil_score_enabled"`
	BlogEnabled             bool `json:"blog_enabled"`
	ChatSupportEnabled      bool `json:"chat_support_enabled"`
	FreeConsultationEnabled bool `json:"free_consultation_enabled"`
	MinCreditScore          int     `json:"min_credit_score"`
	BaseInterestRate        float64 `json:"base_interest_rate"`
}