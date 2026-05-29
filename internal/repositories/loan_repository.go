package repositories

import (
	"lendogo-backend/structures/models"
	"gorm.io/gorm"
)

type LoanRepository interface {
	CreateApplication(loan *models.LoanApplication) error
}

type loanRepositoryImpl struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) LoanRepository {
	return &loanRepositoryImpl{db: db}
}

func (r *loanRepositoryImpl) CreateApplication(loan *models.LoanApplication) error {
	return r.db.Create(loan).Error
}