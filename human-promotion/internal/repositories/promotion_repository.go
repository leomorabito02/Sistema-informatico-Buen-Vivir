package repositories

import (
	"context"

	"buenvivir/human-promotion/internal/models"
)

type PromotionRepository interface {
	FindPrograms(ctx context.Context) ([]*models.Program, error)
}

type promotionRepository struct{}

func NewPromotionRepository() PromotionRepository {
	return &promotionRepository{}
}

func (r *promotionRepository) FindPrograms(ctx context.Context) ([]*models.Program, error) {
	return []*models.Program{
		{ID: "prg-1", Name: "Programa de Apoyo Social y Familiar", TargetGroup: "Familias Vulnerables", BeneficiariesCount: 150},
	}, nil
}
