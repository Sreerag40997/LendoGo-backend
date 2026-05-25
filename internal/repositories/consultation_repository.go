package repositories

import (
	"lendogo-backend/structures/models"

	"gorm.io/gorm"
)

type ConsultationRepository interface {
    Create(consultation *models.Consultation) error
}

type consultationRepositoryImpl struct {
    db *gorm.DB
}

func NewConsultationRepository(db *gorm.DB) ConsultationRepository {
    return &consultationRepositoryImpl{db: db}
}

func (r *consultationRepositoryImpl) Create(consultation *models.Consultation) error {
    return r.db.Create(consultation).Error
}