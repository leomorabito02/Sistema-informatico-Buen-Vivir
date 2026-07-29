package repositories

import (
	"context"
	"time"

	"buenvivir/campaigns/internal/models"
)

type CampaignRepository interface {
	FindAll(ctx context.Context) ([]*models.Campaign, error)
}

type campaignRepository struct{}

func NewCampaignRepository() CampaignRepository {
	return &campaignRepository{}
}

func (r *campaignRepository) FindAll(ctx context.Context) ([]*models.Campaign, error) {
	return []*models.Campaign{
		{
			ID:          "cmp-1",
			Title:       "Campaña de Salud y Bienestar 2026",
			Description: "Jornadas comunitarias de vacunación y nutrición.",
			Status:      "ACTIVE",
			StartDate:   time.Now(),
			EndDate:     time.Now().AddDate(0, 1, 0),
		},
	}, nil
}
