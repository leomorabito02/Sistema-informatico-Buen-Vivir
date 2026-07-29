package services

import (
	"context"

	"buenvivir/rxresume/internal/models"
	"buenvivir/rxresume/internal/repositories"
)

type ResumeService interface {
	GetResume(ctx context.Context, userID string) (*models.Resume, error)
}

type resumeService struct {
	repo repositories.ResumeRepository
}

func NewResumeService(repo repositories.ResumeRepository) ResumeService {
	return &resumeService{repo: repo}
}

func (s *resumeService) GetResume(ctx context.Context, userID string) (*models.Resume, error) {
	return s.repo.FindByUserID(ctx, userID)
}
