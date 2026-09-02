package services

import (
	"context"
	"strings"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// ActivityService valida las reglas de negocio de "actividades" antes
// de delegar en el repository (Firestore). El chequeo de ownership del
// curso (¿este docente puede ver/crear actividades acá?) vive en
// CourseService.VerifyOwnership y lo hace el controller antes de
// llamar a este service, no este service.
type ActivityService interface {
	ListByCourse(ctx context.Context, cursoID string) ([]models.Activity, error)
	GetActivity(ctx context.Context, id string) (*models.Activity, error)
	CreateActivity(ctx context.Context, cursoID string, input models.ActivityInput) (*models.Activity, error)
}

type activityService struct {
	repo repositories.ActivityRepository
}

func NewActivityService(repo repositories.ActivityRepository) ActivityService {
	return &activityService{repo: repo}
}

func (s *activityService) ListByCourse(ctx context.Context, cursoID string) ([]models.Activity, error) {
	return s.repo.ListByCourse(ctx, cursoID)
}

func (s *activityService) GetActivity(ctx context.Context, id string) (*models.Activity, error) {
	if strings.TrimSpace(id) == "" {
		return nil, NewValidationError("el id de la actividad es requerido")
	}
	return s.repo.Get(ctx, id)
}

func (s *activityService) CreateActivity(ctx context.Context, cursoID string, input models.ActivityInput) (*models.Activity, error) {
	if strings.TrimSpace(input.Titulo) == "" {
		return nil, NewValidationError("el campo 'titulo' es requerido")
	}
	if input.Fecha.IsZero() {
		return nil, NewValidationError("el campo 'fecha' es requerido")
	}
	return s.repo.Create(ctx, cursoID, input)
}
