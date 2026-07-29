package services

import (
	"context"

	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/repositories"
)

type PromotionService interface {
	GetPrograms(ctx context.Context) ([]*models.Program, error)
}

type promotionService struct {
	repo repositories.PromotionRepository
}

func NewPromotionService(repo repositories.PromotionRepository) PromotionService {
	return &promotionService{repo: repo}
}

func (s *promotionService) GetPrograms(ctx context.Context) ([]*models.Program, error) {
	return s.repo.FindPrograms(ctx)
}
