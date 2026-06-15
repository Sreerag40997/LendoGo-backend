package services

import (
	"lendogo-backend/internal/repositories"
	"lendogo-backend/structures/models"
)

type ConfigService interface {
	GetConfig() (*models.WebConfiguration, error)
	UpdateConfig(req models.WebConfiguration) (*models.WebConfiguration, error)
}

type configService struct {
	repo repositories.ConfigRepository
}

func NewConfigService(repo repositories.ConfigRepository) ConfigService {
	return &configService{repo: repo}
}

func (s *configService) GetConfig() (*models.WebConfiguration, error) {
	return s.repo.GetConfig()
}

func (s *configService) UpdateConfig(req models.WebConfiguration) (*models.WebConfiguration, error) {
	err := s.repo.UpdateConfig(&req)
	if err != nil {
		return nil, err
	}
	return s.GetConfig()
}