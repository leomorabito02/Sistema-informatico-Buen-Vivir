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

	// ListCoursesByTeacher lista los cursos YA PUBLICADOS donde
	// teacherID figura como responsibleTeacher (ver comentario en
	// CourseRepository.ListByResponsibleTeacher).
	ListCoursesByTeacher(ctx context.Context, teacherID string) ([]models.Course, error)
	// ListDraftCoursesByTeacher lista los cursos de ese docente que
	// todavía NO están publicados (borradores), leídos de la CMA.
	// Complementa a ListCoursesByTeacher, que solo ve los publicados.
	ListDraftCoursesByTeacher(ctx context.Context, teacherID string) ([]models.Course, error)
	// CreateCourseAsTeacher es el POST para el docente responsable:
	// crea el curso en borrador (igual que el POST de admin, sin
	// auto-publicar) forzando responsibleTeacher al id del propio
	// docente autenticado. Queda pendiente de que un role_admin lo
	// revise y publique.
	CreateCourseAsTeacher(ctx context.Context, teacherID string, input models.CourseInput) (*models.Course, error)
	// UpdateCourseAsTeacher es el PUT restringido para el docente
	// responsable: verifica ownership antes de delegar y no permite
	// reasignar responsibleTeacher a otro id.
	UpdateCourseAsTeacher(ctx context.Context, id, teacherID string, input models.CourseInput) (*models.Course, error)
	// VerifyOwnership confirma que cursoID esté publicado y que
	// teacherID sea su responsibleTeacher. La usan también los
	// endpoints de actividades/asistencia/inscriptos/progreso del
	// docente antes de operar sobre un curso.
	VerifyOwnership(ctx context.Context, cursoID, teacherID string) error
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

func (s *courseService) ListCoursesByTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	if strings.TrimSpace(teacherID) == "" {
		return nil, NewValidationError("falta el id del docente autenticado")
	}
	return s.repo.ListByResponsibleTeacher(ctx, teacherID)
}

func (s *courseService) ListDraftCoursesByTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	if strings.TrimSpace(teacherID) == "" {
		return nil, NewValidationError("falta el id del docente autenticado")
	}
	return s.repo.ListDraftsByResponsibleTeacher(ctx, teacherID)
}

// VerifyOwnership usa Get (Content Delivery API, mismo método que el
// GET público) para chequear ownership: si el curso está en borrador
// o no existe, la CDA no lo devuelve y el error se propaga como 404
// (repositories.IsNotFound), no como 403 — desde la perspectiva del
// docente, un curso no publicado directamente "no existe" todavía en
// el catálogo. El 403 queda reservado para cuando el curso sí está
// publicado pero pertenece a otro docente.
func (s *courseService) VerifyOwnership(ctx context.Context, cursoID, teacherID string) error {
	if strings.TrimSpace(cursoID) == "" {
		return NewValidationError("el id del curso es requerido")
	}
	course, err := s.repo.Get(ctx, cursoID)
	if err != nil {
		return err
	}
	if course.ResponsibleTeacher != teacherID {
		return NewForbiddenError("no sos el docente responsable de este curso")
	}
	return nil
}

func (s *courseService) CreateCourseAsTeacher(ctx context.Context, teacherID string, input models.CourseInput) (*models.Course, error) {
	if strings.TrimSpace(teacherID) == "" {
		return nil, NewValidationError("falta el id del docente autenticado")
	}
	// El docente solo puede crear cursos a su propio nombre: se ignora
	// cualquier valor que haya mandado el cliente en este campo.
	input.ResponsibleTeacher = teacherID
	if err := validateCourseInput(input); err != nil {
		return nil, err
	}
	// repo.Create deja la entry en borrador (no auto-publica): el curso
	// no aparece en el catálogo hasta que un role_admin lo publique.
	return s.repo.Create(ctx, input)
}

func (s *courseService) UpdateCourseAsTeacher(ctx context.Context, id, teacherID string, input models.CourseInput) (*models.Course, error) {
	if err := s.VerifyOwnership(ctx, id, teacherID); err != nil {
		return nil, err
	}
	// El docente no puede reasignar el curso a otro responsable: se
	// ignora cualquier valor que haya mandado el cliente en este campo.
	input.ResponsibleTeacher = teacherID
	if err := validateCourseInput(input); err != nil {
		return nil, err
	}
	return s.repo.Update(ctx, id, input)
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
