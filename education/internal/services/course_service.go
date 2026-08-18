package services

import (
	"context"
	"strings"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// CourseService valida las reglas de negocio de "course" antes de
// delegar en el repository. Se expone como interfaz (no struct
// concreto) para poder testearla más adelante con un repository mock,
// aunque por ahora no se generan tests.
type CourseService interface {
	ListCourses(ctx context.Context) ([]models.Course, error)
	GetCourse(ctx context.Context, id string) (*models.Course, error)
	CreateCourse(ctx context.Context, input models.CourseInput) (*models.Course, error)
	UpdateCourse(ctx context.Context, id string, input models.CourseInput) (*models.Course, error)
	DeleteCourse(ctx context.Context, id string) error
	PublishCourse(ctx context.Context, id string) error
	UnpublishCourse(ctx context.Context, id string) error
}

type courseService struct {
	repo repositories.CourseRepository
}

func NewCourseService(repo repositories.CourseRepository) CourseService {
	return &courseService{repo: repo}
}

func (s *courseService) ListCourses(ctx context.Context) ([]models.Course, error) {
	return s.repo.List(ctx)
}

func (s *courseService) GetCourse(ctx context.Context, id string) (*models.Course, error) {
	if strings.TrimSpace(id) == "" {
		return nil, NewValidationError("el id del curso es requerido")
	}
	return s.repo.Get(ctx, id)
}

func (s *courseService) CreateCourse(ctx context.Context, input models.CourseInput) (*models.Course, error) {
	if err := validateCourseInput(input); err != nil {
		return nil, err
	}
	return s.repo.Create(ctx, input)
}

func (s *courseService) UpdateCourse(ctx context.Context, id string, input models.CourseInput) (*models.Course, error) {
	if strings.TrimSpace(id) == "" {
		return nil, NewValidationError("el id del curso es requerido")
	}
	if err := validateCourseInput(input); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, input)
}

func (s *courseService) DeleteCourse(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return NewValidationError("el id del curso es requerido")
	}
	return s.repo.Delete(ctx, id)
}

func (s *courseService) PublishCourse(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return NewValidationError("el id del curso es requerido")
	}
	return s.repo.Publish(ctx, id)
}

func (s *courseService) UnpublishCourse(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return NewValidationError("el id del curso es requerido")
	}
	return s.repo.Unpublish(ctx, id)
}

// validateCourseInput refleja los campos requeridos y los valores
// predefinidos del Content Type "course" en Contentful. "mode" se
// valida contra los valores en español ("Presencial", "Virtual",
// "Mixta") tal como están configurados en el Space.
func validateCourseInput(input models.CourseInput) error {
	if strings.TrimSpace(input.Title) == "" {
		return NewValidationError("el campo 'title' es requerido")
	}
	if len(input.Description) == 0 || string(input.Description) == "null" {
		return NewValidationError("el campo 'description' es requerido")
	}
	if strings.TrimSpace(input.ResponsibleTeacher) == "" {
		return NewValidationError("el campo 'responsibleTeacher' es requerido")
	}
	if strings.TrimSpace(input.Duration) == "" {
		return NewValidationError("el campo 'duration' es requerido")
	}
	if strings.TrimSpace(input.TrainingAreaID) == "" {
		return NewValidationError("el campo 'trainingAreaId' es requerido")
	}
	if len(input.MaterialIDs) == 0 {
		return NewValidationError("el campo 'materialIds' es requerido (al menos un material)")
	}
	mode := models.CourseMode(input.Mode)
	if input.Mode == "" || !mode.IsValid() {
		return NewValidationError("el campo 'mode' debe ser uno de: Presencial, Virtual, Mixta")
	}
	return nil
}
