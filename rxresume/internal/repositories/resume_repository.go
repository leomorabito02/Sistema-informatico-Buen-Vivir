package repositories

import (
	"context"

	"buenvivir/rxresume/internal/models"
)

type ResumeRepository interface {
	FindByUserID(ctx context.Context, userID string) (*models.Resume, error)
}

type resumeRepository struct{}

func NewResumeRepository() ResumeRepository {
	return &resumeRepository{}
}

func (r *resumeRepository) FindByUserID(ctx context.Context, userID string) (*models.Resume, error) {
	return &models.Resume{
		ID:        "rsm-1",
		UserID:    userID,
		Title:     "Curriculum Vitae Buen Vivir",
		Skills:    []string{"Go", "Angular", "PWA", "GCP"},
		Education: "Capacitación Informática Buen Vivir",
	}, nil
}
