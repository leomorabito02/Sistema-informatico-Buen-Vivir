package services

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"buenvivir/education/internal/events"
	"buenvivir/education/internal/events/producers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// CourseService valida las reglas de negocio de "course" antes de
// delegar en el repository. Se expone como interfaz (no struct
// concreto) para poder testearla con un repository mock (ver
// course_service_test.go).
type CourseService interface {
	ListCourses(ctx context.Context) ([]models.Course, error)
	GetCourse(ctx context.Context, id string) (*models.Course, error)
	CreateCourse(ctx context.Context, input models.CourseInput) (*models.Course, error)
	UpdateCourse(ctx context.Context, id string, input models.CourseInput) (*models.Course, error)
	// DeleteCourse es la baja lógica del curso (HU-16): no lo borra de
	// Contentful. Lo marca como dado de baja (sale del catálogo y no
	// acepta inscripciones), con su registro de auditoría, y lo deja
	// publicado para que quien lo cursó lo siga viendo en "Mis cursos".
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
	// ListAllDraftCourses lista TODOS los cursos en borrador (uso de
	// administración: la tabla de "Gestión de cursos" del admin, para
	// no perder de vista un curso creado y no publicado).
	ListAllDraftCourses(ctx context.Context) ([]models.Course, error)
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
	// VerifyExists confirma que el curso exista, publicado o en
	// borrador (consulta la CMA). Lo usan admin/voluntario antes de
	// crear actividades, que se pueden armar antes de publicar.
	VerifyExists(ctx context.Context, cursoID string) error
}

type courseService struct {
	repo          repositories.CourseRepository
	kvSync        KVSyncService
	events        *producers.Educacion
	bajas         repositories.CursoBajaRepository
	inscripciones repositories.EnrollmentRepository
}

func NewCourseService(repo repositories.CourseRepository, kvSync ...KVSyncService) CourseService {
	return NewCourseServiceWithEvents(repo, nil, kvSync...)
}

// NewCourseServiceWithEvents es NewCourseService publicando los eventos
// de cursos (curso.*, contenido_docente.*). events nil = no publica.
func NewCourseServiceWithEvents(repo repositories.CourseRepository, events *producers.Educacion, kvSync ...KVSyncService) CourseService {
	var sync KVSyncService
	if len(kvSync) > 0 {
		sync = kvSync[0]
	}
	return &courseService{repo: repo, kvSync: sync, events: events}
}

// NewCourseServiceConBaja suma la baja lógica (HU-16): bajas guarda la
// marca y inscripciones permite contar los inscriptos activos al darlo
// de baja. Sin bajas, DeleteCourse falla y el catálogo no filtra nada.
func NewCourseServiceConBaja(repo repositories.CourseRepository, events *producers.Educacion, bajas repositories.CursoBajaRepository, inscripciones repositories.EnrollmentRepository, kvSync ...KVSyncService) CourseService {
	s := NewCourseServiceWithEvents(repo, events, kvSync...).(*courseService)
	s.bajas = bajas
	s.inscripciones = inscripciones
	return s
}

// errBajaNoConfigurada: el servicio se armó sin el repositorio de bajas.
var errBajaNoConfigurada = errors.New("la baja de cursos no está configurada")

// sinBajas saca de la lista los cursos dados de baja (HU-16).
func (s *courseService) sinBajas(ctx context.Context, courses []models.Course, err error) ([]models.Course, error) {
	if err != nil || s.bajas == nil {
		return courses, err
	}
	bajas, err := s.bajas.IDs(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]models.Course, 0, len(courses))
	for _, c := range courses {
		if !bajas[c.ID] {
			out = append(out, c)
		}
	}
	return out, nil
}

// publish registra un evento de curso. Los cursos viven en Contentful,
// sin transacción compartida con el outbox: el evento se escribe
// después de que Contentful confirmó el cambio. Si esa escritura falla,
// el cambio ya está hecho y no se revierte: se deja en el log (con el
// correlation id) para poder reenviarlo. Ver docs/eventos/README.md.
func (s *courseService) publish(ctx context.Context, eventType, cursoID string, emit func() error) {
	if err := emit(); err != nil {
		slog.ErrorContext(ctx, "no se pudo registrar el evento de curso",
			slog.String("event_type", eventType),
			slog.String("curso_id", cursoID),
			slog.String("correlation_id", events.CorrelationIDFrom(ctx)),
			slog.String("resultado", "error"),
			slog.String("error", err.Error()))
	}
}

func (s *courseService) syncKV() {
	if s.kvSync != nil {
		s.kvSync.SyncCourseStatsAsync()
	}
}

func (s *courseService) ListCourses(ctx context.Context) ([]models.Course, error) {
	courses, err := s.repo.List(ctx)
	return s.sinBajas(ctx, courses, err)
}

const (
	errCourseIDRequired = "el id del curso es requerido"
	errTeacherIDMissing = "falta el id del docente autenticado"
)

func (s *courseService) GetCourse(ctx context.Context, id string) (*models.Course, error) {
	if strings.TrimSpace(id) == "" {
		return nil, NewValidationError(errCourseIDRequired)
	}
	return s.repo.Get(ctx, id)
}

func (s *courseService) CreateCourse(ctx context.Context, input models.CourseInput) (*models.Course, error) {
	if err := validateCourseInput(input); err != nil {
		return nil, err
	}
	course, err := s.repo.Create(ctx, input)
	if err == nil {
		s.syncKV()
		s.publish(ctx, events.CursoCreado, course.ID, func() error { return s.events.CursoCreado(ctx, course) })
	}
	return course, err
}

func (s *courseService) UpdateCourse(ctx context.Context, id string, input models.CourseInput) (*models.Course, error) {
	if strings.TrimSpace(id) == "" {
		return nil, NewValidationError(errCourseIDRequired)
	}
	if err := validateCourseInput(input); err != nil {
		return nil, err
	}
	course, err := s.repo.Update(ctx, id, input)
	if err == nil {
		s.syncKV()
		s.publish(ctx, events.CursoActualizado, course.ID, func() error { return s.events.CursoActualizado(ctx, course) })
	}
	return course, err
}

func (s *courseService) DeleteCourse(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return NewValidationError(errCourseIDRequired)
	}
	if s.bajas == nil || s.inscripciones == nil {
		return errBajaNoConfigurada
	}
	exists, err := s.repo.ExistsInManagement(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return repositories.ErrNotFound
	}
	inscripciones, err := s.inscripciones.ListByCourse(ctx, id)
	if err != nil {
		return err
	}
	// El título queda en la baja para leerla sin Contentful. Un borrador
	// no está en la CDA: queda vacío.
	titulo := ""
	if c, err := s.repo.Get(ctx, id); err == nil {
		titulo = c.Title
	}
	err = s.bajas.DarDeBaja(ctx, models.CursoBaja{
		CursoID:           id,
		Titulo:            titulo,
		FechaBaja:         time.Now().UTC(),
		UsuarioID:         events.ActorFrom(ctx),
		InscriptosActivos: ContarInscripcionesActivas(inscripciones),
	})
	if errors.Is(err, repositories.ErrCursoYaDadoDeBaja) {
		return NewValidationError("El curso ya fue dado de baja")
	}
	if err != nil {
		return err
	}
	s.syncKV()
	s.publish(ctx, events.CursoEliminado, id, func() error { return s.events.CursoEliminado(ctx, id) })
	return nil
}

// ContarInscripcionesActivas cuenta las inscripciones en estado
// "activa": es el aviso de HU-16 antes de dar de baja un curso.
func ContarInscripcionesActivas(inscripciones []models.Enrollment) int {
	n := 0
	for _, e := range inscripciones {
		if e.Estado == models.EnrollmentEstadoActiva {
			n++
		}
	}
	return n
}

func (s *courseService) PublishCourse(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return NewValidationError(errCourseIDRequired)
	}
	err := s.repo.Publish(ctx, id)
	if err == nil {
		s.syncKV()
		// Publicar es la moderación editorial del curso: lo aprueba para
		// el catálogo (HU-23). No existe todavía un "rechazo" explícito.
		s.publish(ctx, events.ContenidoDocenteModerado, id, func() error {
			return s.events.ContenidoDocenteModerado(ctx, id, events.ModeracionAprobado)
		})
	}
	return err
}

func (s *courseService) UnpublishCourse(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return NewValidationError(errCourseIDRequired)
	}
	err := s.repo.Unpublish(ctx, id)
	if err == nil {
		s.syncKV()
	}
	return err
}

func (s *courseService) ListCoursesByTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	if strings.TrimSpace(teacherID) == "" {
		return nil, NewValidationError(errTeacherIDMissing)
	}
	return s.repo.ListByResponsibleTeacher(ctx, teacherID)
}

func (s *courseService) ListDraftCoursesByTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	if strings.TrimSpace(teacherID) == "" {
		return nil, NewValidationError(errTeacherIDMissing)
	}
	return s.repo.ListDraftsByResponsibleTeacher(ctx, teacherID)
}

func (s *courseService) ListAllDraftCourses(ctx context.Context) ([]models.Course, error) {
	courses, err := s.repo.ListAllDrafts(ctx)
	return s.sinBajas(ctx, courses, err)
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
		return NewValidationError(errCourseIDRequired)
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

func (s *courseService) VerifyExists(ctx context.Context, cursoID string) error {
	if strings.TrimSpace(cursoID) == "" {
		return NewValidationError(errCourseIDRequired)
	}
	exists, err := s.repo.ExistsInManagement(ctx, cursoID)
	if err != nil {
		return err
	}
	if !exists {
		return repositories.ErrNotFound
	}
	return nil
}

func (s *courseService) CreateCourseAsTeacher(ctx context.Context, teacherID string, input models.CourseInput) (*models.Course, error) {
	if strings.TrimSpace(teacherID) == "" {
		return nil, NewValidationError(errTeacherIDMissing)
	}
	// El docente solo puede crear cursos a su propio nombre: se ignora
	// cualquier valor que haya mandado el cliente en este campo.
	input.ResponsibleTeacher = teacherID
	if err := validateCourseInput(input); err != nil {
		return nil, err
	}
	// repo.Create deja la entry en borrador (no auto-publica): el curso
	// no aparece en el catálogo hasta que un role_admin lo publique.
	course, err := s.repo.Create(ctx, input)
	if err == nil {
		s.syncKV()
		s.publish(ctx, events.CursoCreado, course.ID, func() error { return s.events.CursoCreado(ctx, course) })
		s.publish(ctx, events.ContenidoDocenteSubido, course.ID, func() error { return s.events.ContenidoDocenteSubido(ctx, course) })
	}
	return course, err
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
	course, err := s.repo.Update(ctx, id, input)
	if err == nil {
		s.syncKV()
		s.publish(ctx, events.CursoActualizado, course.ID, func() error { return s.events.CursoActualizado(ctx, course) })
	}
	return course, err
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
