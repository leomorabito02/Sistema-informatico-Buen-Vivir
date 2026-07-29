package repositories

import (
	"context"

	"buenvivir/dashboard/internal/models"
)

type DashboardRepository interface {
	GetMetrics(ctx context.Context) (*models.Metrics, error)
}

type dashboardRepository struct{}

func NewDashboardRepository() DashboardRepository {
	return &dashboardRepository{}
}

func (r *dashboardRepository) GetMetrics(ctx context.Context) (*models.Metrics, error) {
	return &models.Metrics{
		TotalUsers:       1250,
		ActiveCampaigns:  8,
		CompletedCourses: 430,
	}, nil
}
