package services

import (
	"context"

	"buenvivir/campaigns/internal/models"
	"buenvivir/campaigns/internal/repositories"
)

type CampaignService interface {
	GetCampaigns(ctx context.Context) ([]*models.Campaign, error)
}

type campaignService struct {
	repo repositories.CampaignRepository
}

func NewCampaignService(repo repositories.CampaignRepository) CampaignService {
	return &campaignService{repo: repo}
}

func (s *campaignService) GetCampaigns(ctx context.Context) ([]*models.Campaign, error) {
	return s.repo.FindAll(ctx)
}
