package services

import (
	"context"
	"strings"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// TrainingAreaService valida las reglas de negocio de "trainingArea".
type TrainingAreaService interface {
	ListTrainingAreas(ctx context.Context) ([]models.TrainingArea, error)
	CreateTrainingArea(ctx context.Context, input models.TrainingAreaInput) (*models.TrainingArea, error)
	UpdateTrainingArea(ctx context.Context, id string, input models.TrainingAreaInput) (*models.TrainingArea, error)
	DeleteTrainingArea(ctx context.Context, id string) error
}

type trainingAreaService struct {
	repo repositories.TrainingAreaRepository
}

func NewTrainingAreaService(repo repositories.TrainingAreaRepository) TrainingAreaService {
	return &trainingAreaService{repo: repo}
}

func (s *trainingAreaService) ListTrainingAreas(ctx context.Context) ([]models.TrainingArea, error) {
	return s.repo.List(ctx)
}

func (s *trainingAreaService) CreateTrainingArea(ctx context.Context, input models.TrainingAreaInput) (*models.TrainingArea, error) {
	if err := validateTrainingAreaInput(input); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, input)
}

func (s *trainingAreaService) UpdateTrainingArea(ctx context.Context, id string, input models.TrainingAreaInput) (*models.TrainingArea, error) {
	if strings.TrimSpace(id) == "" {
		return nil, NewValidationError("el id del área formativa es requerido")
	}
	if err := validateTrainingAreaInput(input); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, input)
}

func (s *trainingAreaService) DeleteTrainingArea(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return NewValidationError("el id del área formativa es requerido")
	}
	return s.repo.Delete(ctx, id)
}

func validateTrainingAreaInput(input models.TrainingAreaInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return NewValidationError("el campo 'name' es requerido")
	}
	return nil
}
