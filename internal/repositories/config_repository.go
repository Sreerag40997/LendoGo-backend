package repositories

import (
	"errors"
	"lendogo-backend/structures/models"
	"gorm.io/gorm"
)

type ConfigRepository interface {
	GetConfig() (*models.WebConfiguration, error)
	UpdateConfig(config *models.WebConfiguration) error
}
type configRepository struct {
	db *gorm.DB
}
func NewConfigRepository(db *gorm.DB) ConfigRepository {
	return &configRepository{db: db}
}
func (r *configRepository) GetConfig() (*models.WebConfiguration, error) {
	var config models.WebConfiguration
	err := r.db.First(&config, 1).Error
	
	// If the row doesn't exist yet, seed the default settings automatically
	if errors.Is(err, gorm.ErrRecordNotFound) {
		config.ID = 1
		r.db.Create(&config)
		return &config, nil
	}
	
	return &config, err
}

func (r *configRepository) UpdateConfig(config *models.WebConfiguration) error {
	config.ID = 1 // Force ID 1 to prevent creating new rows
	return r.db.Save(config).Error
}