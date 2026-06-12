package dto

// DisburseLoanRequest catches the data from the Admin React Modal
type DisburseLoanRequest struct {
	LoanID        string  `json:"loan_id" validate:"required"`
	UserID        string  `json:"user_id" validate:"required"`
	SanctionedAmt float64 `json:"sanctioned_amount" validate:"required,gt=0"`
	ProcessingFee float64 `json:"processing_fee" validate:"gte=0"`
	NetPayout     float64 `json:"net_payout" validate:"required,gt=0"`
}