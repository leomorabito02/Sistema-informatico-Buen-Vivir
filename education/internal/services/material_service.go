package services

import (
	"context"
	"strings"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// MaterialService valida las reglas de negocio de "material".
type MaterialService interface {
	ListMaterials(ctx context.Context) ([]models.Material, error)
	CreateMaterial(ctx context.Context, input models.MaterialInput) (*models.Material, error)
	UpdateMaterial(ctx context.Context, id string, input models.MaterialInput) (*models.Material, error)
	DeleteMaterial(ctx context.Context, id string) error
}

type materialService struct {
	repo repositories.MaterialRepository
}

func NewMaterialService(repo repositories.MaterialRepository) MaterialService {
	return &materialService{repo: repo}
}

func (s *materialService) ListMaterials(ctx context.Context) ([]models.Material, error) {
	return s.repo.List(ctx)
}

func (s *materialService) CreateMaterial(ctx context.Context, input models.MaterialInput) (*models.Material, error) {
	if err := validateMaterialInput(input); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, input)
}

func (s *materialService) UpdateMaterial(ctx context.Context, id string, input models.MaterialInput) (*models.Material, error) {
	if strings.TrimSpace(id) == "" {
		return nil, NewValidationError("el id del material es requerido")
	}
	if err := validateMaterialInput(input); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, input)
}

func (s *materialService) DeleteMaterial(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return NewValidationError("el id del material es requerido")
	}
	return s.repo.Delete(ctx, id)
}

func validateMaterialInput(input models.MaterialInput) error {
	if strings.TrimSpace(input.Title) == "" {
		return NewValidationError("el campo 'title' es requerido")
	}
	materialType := models.MaterialType(input.Type)
	if !materialType.IsValid() {
		return NewValidationError("el campo 'type' debe ser uno de: PDF, Video, Link")
	}
	return nil
}
