package services

import (
	"context"

	"buenvivir/dashboard/internal/models"
	"buenvivir/dashboard/internal/repositories"
)

type DashboardService interface {
	GetMetrics(ctx context.Context) (*models.Metrics, error)
}

type dashboardService struct {
	repo repositories.DashboardRepository
}

func NewDashboardService(repo repositories.DashboardRepository) DashboardService {
	return &dashboardService{repo: repo}
}

func (s *dashboardService) GetMetrics(ctx context.Context) (*models.Metrics, error) {
	return s.repo.GetMetrics(ctx)
}
