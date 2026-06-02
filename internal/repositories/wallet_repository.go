package repositories

import (
	"lendogo-backend/structures/models"
	"gorm.io/gorm"
)

type WalletRepository interface {
	GetSystemBalance() (float64, error)
	CreditSystemWallet(amount float64) error
}

type walletRepositoryImpl struct {
	db *gorm.DB
}

func NewWalletRepository(db *gorm.DB) WalletRepository {
	return &walletRepositoryImpl{db: db}
}

func (r *walletRepositoryImpl) GetSystemBalance() (float64, error) {
	var wallet models.SystemWallet
	if err := r.db.Where("wallet_name = ?", "capital_disbursement").First(&wallet).Error; err != nil {
		return 0, err
	}
	return wallet.Balance, nil
}

func (r *walletRepositoryImpl) CreditSystemWallet(amount float64) error {
	return r.db.Model(&models.SystemWallet{}).
		Where("wallet_name = ?", "capital_disbursement").
		UpdateColumn("balance", gorm.Expr("balance + ?", amount)).Error
}