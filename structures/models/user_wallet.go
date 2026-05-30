package models

import (
	"time"
	"github.com/google/uuid"
)

// UserWallet holds the borrower's actual money
type UserWallet struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"` // Matches your user ID type
	Balance   float64   `gorm:"type:numeric(15,2);not null;default:0.00" json:"balance"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LedgerEntry is the permanent, immutable receipt
type LedgerEntry struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	WalletID        uuid.UUID `gorm:"type:uuid;index;not null" json:"wallet_id"` 
	Amount          float64   `gorm:"type:numeric(15,2);not null" json:"amount"`
	TransactionType string    `gorm:"type:varchar(50);not null" json:"transaction_type"` 
	ReferenceID     string    `gorm:"type:varchar(100);index" json:"reference_id"`     
	CreatedAt       time.Time `gorm:"index" json:"created_at"`
}