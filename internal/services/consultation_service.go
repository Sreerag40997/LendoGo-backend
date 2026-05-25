package services

import (
	"lendogo-backend/internal/repositories"
	"lendogo-backend/structures/dto"
	"lendogo-backend/structures/models"
)

type ConsultationService interface {
    RequestConsultation(req dto.ConsultationReq) error
}

type consultationServiceImpl struct {
    repo repositories.ConsultationRepository
}

func NewConsultationService(repo repositories.ConsultationRepository) ConsultationService {
    return &consultationServiceImpl{repo: repo}
}

func (s *consultationServiceImpl) RequestConsultation(req dto.ConsultationReq) error {
    newConsultation := &models.Consultation{
        FullName:    req.FullName,
        Email:       req.Email,
        PhoneNumber: req.PhoneNumber,
    }

    return s.repo.Create(newConsultation)
}