package services

import (
	"context"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

type EducationService interface {
	GetCourses(ctx context.Context) ([]*models.Course, error)
}

type educationService struct {
	repo repositories.EducationRepository
}

func NewEducationService(repo repositories.EducationRepository) EducationService {
	return &educationService{repo: repo}
}

func (s *educationService) GetCourses(ctx context.Context) ([]*models.Course, error) {
	return s.repo.FindCourses(ctx)
}
