package services

import (
	"context"
	"log"
	"sort"
	"strings"
	"time"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// ActivityService valida las reglas de negocio de las actividades. El ownership del
// curso lo verifica el controller antes (CourseService.VerifyOwnership).
type ActivityService interface {
	// ListByCourse devuelve las actividades del curso ordenadas por
	// fecha (y por fecha de alta para desempatar).
	ListByCourse(ctx context.Context, cursoID string) ([]models.Activity, error)
	GetActivity(ctx context.Context, id string) (*models.Activity, error)
	CreateActivity(ctx context.Context, cursoID string, input models.ActivityInput) (*models.Activity, error)
	UpdateActivity(ctx context.Context, id string, input models.ActivityInput) (*models.Activity, error)
	// DeleteActivity borra la actividad con sus entregas y asistencias
	// (ver ActivityRepository.Delete) y después sus adjuntos del bucket.
	DeleteActivity(ctx context.Context, id string) error
}

type activityService struct {
	repo    repositories.ActivityRepository
	storage repositories.FileStorage
}

// NewActivityService recibe el FileStorage de entregas para poder
// limpiar los adjuntos al eliminar una actividad; puede ser nil si el
// bucket no está configurado.
func NewActivityService(repo repositories.ActivityRepository, storage repositories.FileStorage) ActivityService {
	return &activityService{repo: repo, storage: storage}
}

func (s *activityService) ListByCourse(ctx context.Context, cursoID string) ([]models.Activity, error) {
	activities, err := s.repo.ListByCourse(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	sortActivities(activities)
	return activities, nil
}

func (s *activityService) GetActivity(ctx context.Context, id string) (*models.Activity, error) {
	if strings.TrimSpace(id) == "" {
		return nil, NewValidationError("el id de la actividad es requerido")
	}
	return s.repo.Get(ctx, id)
}

func (s *activityService) CreateActivity(ctx context.Context, cursoID string, input models.ActivityInput) (*models.Activity, error) {
	input, err := normalizeActivityInput(input)
	if err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, cursoID, input)
}

func (s *activityService) UpdateActivity(ctx context.Context, id string, input models.ActivityInput) (*models.Activity, error) {
	if strings.TrimSpace(id) == "" {
		return nil, NewValidationError("el id de la actividad es requerido")
	}
	input, err := normalizeActivityInput(input)
	if err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, input)
}

func (s *activityService) DeleteActivity(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return NewValidationError("el id de la actividad es requerido")
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	// Los datos ya se borraron: si el bucket falla, los adjuntos quedan
	// huérfanos pero inaccesibles (ninguna entrega los referencia), así
	// que no se revierte la baja por esto.
	if s.storage != nil {
		if err := s.storage.DeletePrefix(ctx, submissionActivityPrefix(id)); err != nil {
			log.Printf("no se pudieron borrar los adjuntos de la actividad %s: %v", id, err)
		}
	}
	return nil
}

// normalizeActivityInput valida los campos obligatorios y lleva la
// fecha a medianoche UTC de su día: la actividad tiene fecha
// calendario, no hora, y esa misma fecha es la que toma la Asistencia
// de cada entrega.
func normalizeActivityInput(input models.ActivityInput) (models.ActivityInput, error) {
	input.Titulo = strings.TrimSpace(input.Titulo)
	input.Consigna = strings.TrimSpace(input.Consigna)
	if input.Titulo == "" {
		return input, NewValidationError("el campo 'titulo' es requerido")
	}
	if input.Consigna == "" {
		return input, NewValidationError("el campo 'consigna' es requerido")
	}
	if input.Fecha.IsZero() {
		return input, NewValidationError("el campo 'fecha' es requerido")
	}
	y, m, d := input.Fecha.UTC().Date()
	input.Fecha = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return input, nil
}

func sortActivities(activities []models.Activity) {
	sort.SliceStable(activities, func(i, j int) bool {
		if !activities[i].Fecha.Equal(activities[j].Fecha) {
			return activities[i].Fecha.Before(activities[j].Fecha)
		}
		return activities[i].CreatedAt.Before(activities[j].CreatedAt)
	})
}
