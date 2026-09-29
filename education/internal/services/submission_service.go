package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

const (
	maxSubmissionFiles     = 5
	maxSubmissionFileBytes = 10 << 20
	maxSubmissionTextRunes = 10000
)

// allowedSubmissionTypes son los tipos de adjunto que acepta una
// entrega: documentos de oficina, PDF, texto e imágenes.
var allowedSubmissionTypes = map[string]bool{
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"application/vnd.ms-excel": true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"application/vnd.ms-powerpoint":                                             true,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
	"text/plain": true,
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

var unsafeFileNameChars = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// ParticipantDirectory es lo que SubmissionService necesita de
// auth-service (lo implementa *repositories.AuthServiceClient).
type ParticipantDirectory interface {
	GetMyParticipantID(ctx context.Context, bearerToken string) (string, error)
	ListParticipants(ctx context.Context, bearerToken string) ([]models.AuthParticipant, error)
	ListUsers(ctx context.Context, bearerToken string) ([]models.AuthUser, error)
}

// SubmissionService implementa las entregas de actividades y la vista
// del curso del rol participante. Una entrega es única e inmutable, y
// genera automáticamente la Asistencia del participante en esa
// actividad (presente=true, fecha = fecha de la actividad), de la que
// sale el progreso (HU-19, HU-20).
type SubmissionService interface {
	// ResolveParticipantID traduce el token del usuario autenticado a su
	// participant_id de auth-service. Forbidden si no tiene perfil de
	// participante.
	ResolveParticipantID(ctx context.Context, bearerToken string) (string, error)
	// Submit registra la entrega de participanteID en actividadID. El
	// participante tiene que tener una inscripción activa en el curso
	// de la actividad y no haberla entregado antes.
	Submit(ctx context.Context, actividadID, participanteID, texto string, uploads []models.SubmissionUpload) (*models.Submission, error)
	// ListByActivity devuelve una fila por participante con inscripción
	// activa en el curso de la actividad (más quien haya entregado
	// aunque ya no esté activo), con su estado de entrega.
	ListByActivity(ctx context.Context, activity *models.Activity, bearerToken string) ([]models.SubmissionStatusRow, error)
	GetSubmission(ctx context.Context, id string) (*models.Submission, error)
	// OpenFile abre el adjunto número index de la entrega. El control de
	// acceso lo hace el controller antes de llamar acá.
	OpenFile(ctx context.Context, sub *models.Submission, index int) (models.SubmissionFile, io.ReadCloser, error)
	// ListParticipantCourses devuelve los cursos publicados en los que
	// el participante está inscripto (inscripción activa o finalizada).
	ListParticipantCourses(ctx context.Context, participanteID string) ([]models.Course, error)
	// GetParticipantCourse devuelve el curso con sus actividades, el
	// estado de entrega de cada una y el progreso del participante.
	// Forbidden si no está inscripto en ese curso.
	GetParticipantCourse(ctx context.Context, participanteID, cursoID string) (*models.ParticipantCourseDetail, error)
	// VerifyParticipantEnrolled devuelve Forbidden si el participante no
	// tiene una inscripción (activa o finalizada) en el curso.
	VerifyParticipantEnrolled(ctx context.Context, participanteID, cursoID string) error
}

type submissionService struct {
	repo           repositories.SubmissionRepository
	activityRepo   repositories.ActivityRepository
	enrollmentRepo repositories.EnrollmentRepository
	courseRepo     repositories.CourseRepository
	storage        repositories.FileStorage
	directory      ParticipantDirectory
	now            func() time.Time
}

// NewSubmissionService arma el service. storage puede ser nil (bucket
// sin configurar): en ese caso solo se aceptan entregas sin adjuntos.
func NewSubmissionService(
	repo repositories.SubmissionRepository,
	activityRepo repositories.ActivityRepository,
	enrollmentRepo repositories.EnrollmentRepository,
	courseRepo repositories.CourseRepository,
	storage repositories.FileStorage,
	directory ParticipantDirectory,
) SubmissionService {
	return &submissionService{
		repo:           repo,
		activityRepo:   activityRepo,
		enrollmentRepo: enrollmentRepo,
		courseRepo:     courseRepo,
		storage:        storage,
		directory:      directory,
		now:            func() time.Time { return time.Now().UTC() },
	}
}

func (s *submissionService) ResolveParticipantID(ctx context.Context, bearerToken string) (string, error) {
	id, err := s.directory.GetMyParticipantID(ctx, bearerToken)
	if err != nil {
		return "", err
	}
	if id == "" {
		return "", NewForbiddenError("tu usuario no tiene un perfil de participante")
	}
	return id, nil
}

func (s *submissionService) Submit(ctx context.Context, actividadID, participanteID, texto string, uploads []models.SubmissionUpload) (*models.Submission, error) {
	texto = strings.TrimSpace(texto)
	if len([]rune(texto)) > maxSubmissionTextRunes {
		return nil, NewValidationError(fmt.Sprintf("el texto de la entrega no puede superar los %d caracteres", maxSubmissionTextRunes))
	}
	if err := s.validateUploads(uploads); err != nil {
		return nil, err
	}

	activity, err := s.activityRepo.Get(ctx, actividadID)
	if err != nil {
		return nil, err
	}
	active, err := s.enrollmentStatus(ctx, participanteID, activity.CursoID)
	if err != nil {
		return nil, err
	}
	if active != models.EnrollmentEstadoActiva {
		return nil, NewForbiddenError("no tenés una inscripción activa en el curso de esta actividad")
	}

	// Chequeo previo para no subir archivos al bucket de una entrega que
	// igual va a ser rechazada. La garantía real de unicidad es el
	// Create de la transacción (ver CreateWithAttendance).
	existing, err := s.repo.GetForParticipant(ctx, participanteID, []string{actividadID})
	if err != nil {
		return nil, err
	}
	if existing[actividadID] != nil {
		return nil, NewValidationError("Ya entregaste esta actividad")
	}

	now := s.now()
	uploadPrefix := fmt.Sprintf("%s%s/%d-", submissionActivityPrefix(actividadID), participanteID, now.UnixNano())
	files := make([]models.SubmissionFile, 0, len(uploads))
	for i, up := range uploads {
		objectPath := fmt.Sprintf("%s%d-%s", uploadPrefix, i, safeFileName(up.Nombre))
		if err := s.storage.Upload(ctx, objectPath, up.Tipo, up.Data); err != nil {
			s.cleanupUploads(ctx, uploadPrefix, len(files) > 0)
			return nil, err
		}
		files = append(files, models.SubmissionFile{
			Nombre: path.Base(strings.ReplaceAll(up.Nombre, "\\", "/")),
			Tipo:   up.Tipo,
			Tamano: int64(len(up.Data)),
			Ruta:   objectPath,
		})
	}

	sub, err := s.repo.CreateWithAttendance(ctx, activity, models.Submission{
		ParticipanteID: participanteID,
		Texto:          texto,
		Archivos:       files,
		FechaEntrega:   now,
	})
	if err != nil {
		s.cleanupUploads(ctx, uploadPrefix, len(files) > 0)
		if err == repositories.ErrAlreadySubmitted {
			return nil, NewValidationError("Ya entregaste esta actividad")
		}
		return nil, err
	}
	return sub, nil
}

func (s *submissionService) validateUploads(uploads []models.SubmissionUpload) error {
	if len(uploads) == 0 {
		return nil
	}
	if s.storage == nil {
		return NewValidationError("la carga de archivos no está disponible en este momento; podés entregar solo con texto")
	}
	if len(uploads) > maxSubmissionFiles {
		return NewValidationError(fmt.Sprintf("podés adjuntar hasta %d archivos", maxSubmissionFiles))
	}
	for _, up := range uploads {
		if strings.TrimSpace(up.Nombre) == "" {
			return NewValidationError("cada archivo necesita un nombre")
		}
		if !allowedSubmissionTypes[up.Tipo] {
			return NewValidationError(fmt.Sprintf("el archivo %q tiene un formato no permitido (PDF, Word, Excel, PowerPoint, texto o imagen)", up.Nombre))
		}
		if len(up.Data) == 0 {
			return NewValidationError(fmt.Sprintf("el archivo %q está vacío", up.Nombre))
		}
		if len(up.Data) > maxSubmissionFileBytes {
			return NewValidationError(fmt.Sprintf("el archivo %q supera los 10 MB", up.Nombre))
		}
	}
	return nil
}

func (s *submissionService) cleanupUploads(ctx context.Context, prefix string, any bool) {
	if !any {
		return
	}
	if err := s.storage.DeletePrefix(ctx, prefix); err != nil {
		log.Printf("no se pudieron limpiar los adjuntos de una entrega fallida (%s): %v", prefix, err)
	}
}

func (s *submissionService) ListByActivity(ctx context.Context, activity *models.Activity, bearerToken string) ([]models.SubmissionStatusRow, error) {
	enrollments, err := s.enrollmentRepo.ListByCourse(ctx, activity.CursoID)
	if err != nil {
		return nil, err
	}
	submissions, err := s.repo.ListByActivity(ctx, activity.ID)
	if err != nil {
		return nil, err
	}

	byParticipant := make(map[string]*models.Submission, len(submissions))
	for i := range submissions {
		byParticipant[submissions[i].ParticipanteID] = &submissions[i]
	}

	rows := make([]models.SubmissionStatusRow, 0, len(enrollments))
	seen := make(map[string]bool, len(enrollments))
	addRow := func(participanteID string) {
		if seen[participanteID] {
			return
		}
		seen[participanteID] = true
		row := models.SubmissionStatusRow{ParticipanteID: participanteID, Estado: models.SubmissionEstadoSinEntregar}
		if sub := byParticipant[participanteID]; sub != nil {
			row.Estado = models.SubmissionEstadoEntregado
			row.Entrega = sub
		}
		rows = append(rows, row)
	}
	for _, e := range enrollments {
		if e.Estado == models.EnrollmentEstadoActiva {
			addRow(e.ParticipanteID)
		}
	}
	for _, sub := range submissions {
		addRow(sub.ParticipanteID)
	}

	s.fillNames(ctx, rows, bearerToken)
	sort.SliceStable(rows, func(i, j int) bool {
		return strings.ToLower(rowSortKey(rows[i])) < strings.ToLower(rowSortKey(rows[j]))
	})
	return rows, nil
}

// fillNames resuelve nombre/apellido contra auth-service. No es fatal:
// si falla (o el rol que llama no puede listar participantes en auth),
// las filas quedan con el participanteId y el frontend lo muestra así.
func (s *submissionService) fillNames(ctx context.Context, rows []models.SubmissionStatusRow, bearerToken string) {
	if len(rows) == 0 || s.directory == nil {
		return
	}
	participants, err := s.directory.ListParticipants(ctx, bearerToken)
	if err != nil {
		return
	}
	users, err := s.directory.ListUsers(ctx, bearerToken)
	if err != nil {
		return
	}
	usersByID := make(map[string]models.AuthUser, len(users))
	for _, u := range users {
		usersByID[u.UserID] = u
	}
	userByParticipant := make(map[string]models.AuthUser, len(participants))
	for _, p := range participants {
		if u, ok := usersByID[p.UserID]; ok {
			userByParticipant[p.ParticipantID] = u
		}
	}
	for i := range rows {
		if u, ok := userByParticipant[rows[i].ParticipanteID]; ok {
			rows[i].Nombre = u.FirstName
			rows[i].Apellido = u.LastName
		}
	}
}

func rowSortKey(r models.SubmissionStatusRow) string {
	if name := strings.TrimSpace(r.Apellido + " " + r.Nombre); name != "" {
		return name
	}
	return r.ParticipanteID
}

func (s *submissionService) GetSubmission(ctx context.Context, id string) (*models.Submission, error) {
	if strings.TrimSpace(id) == "" {
		return nil, NewValidationError("el id de la entrega es requerido")
	}
	return s.repo.Get(ctx, id)
}

func (s *submissionService) OpenFile(ctx context.Context, sub *models.Submission, index int) (models.SubmissionFile, io.ReadCloser, error) {
	if index < 0 || index >= len(sub.Archivos) {
		return models.SubmissionFile{}, nil, repositories.ErrNotFound
	}
	if s.storage == nil {
		return models.SubmissionFile{}, nil, repositories.ErrFileStorageNotConfigured
	}
	file := sub.Archivos[index]
	rc, err := s.storage.Open(ctx, file.Ruta)
	if err != nil {
		return models.SubmissionFile{}, nil, err
	}
	return file, rc, nil
}

func (s *submissionService) ListParticipantCourses(ctx context.Context, participanteID string) ([]models.Course, error) {
	enrollments, err := s.enrollmentRepo.ListByParticipant(ctx, participanteID)
	if err != nil {
		return nil, err
	}
	courses := make([]models.Course, 0, len(enrollments))
	seen := make(map[string]bool, len(enrollments))
	for _, e := range enrollments {
		if e.Estado == models.EnrollmentEstadoBaja || seen[e.CursoID] {
			continue
		}
		seen[e.CursoID] = true
		course, err := s.courseRepo.Get(ctx, e.CursoID)
		if err != nil {
			// Un curso despublicado o dado de baja deja de estar en la
			// CDA: se omite en vez de romper el listado entero.
			if repositories.IsNotFound(err) {
				continue
			}
			return nil, err
		}
		courses = append(courses, *course)
	}
	sort.SliceStable(courses, func(i, j int) bool {
		return strings.ToLower(courses[i].Title) < strings.ToLower(courses[j].Title)
	})
	return courses, nil
}

func (s *submissionService) GetParticipantCourse(ctx context.Context, participanteID, cursoID string) (*models.ParticipantCourseDetail, error) {
	if err := s.VerifyParticipantEnrolled(ctx, participanteID, cursoID); err != nil {
		return nil, err
	}
	course, err := s.courseRepo.Get(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	activities, err := s.activityRepo.ListByCourse(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	sortActivities(activities)

	ids := make([]string, len(activities))
	for i, a := range activities {
		ids[i] = a.ID
	}
	submissions, err := s.repo.GetForParticipant(ctx, participanteID, ids)
	if err != nil {
		return nil, err
	}

	items := make([]models.ParticipantActivity, len(activities))
	for i, a := range activities {
		items[i] = models.ParticipantActivity{Activity: a, Entrega: submissions[a.ID]}
	}

	return &models.ParticipantCourseDetail{
		Curso:       *course,
		Actividades: items,
		Progreso:    buildProgress(participanteID, len(activities), len(submissions)),
	}, nil
}

func (s *submissionService) VerifyParticipantEnrolled(ctx context.Context, participanteID, cursoID string) error {
	estado, err := s.enrollmentStatus(ctx, participanteID, cursoID)
	if err != nil {
		return err
	}
	if estado == "" || estado == models.EnrollmentEstadoBaja {
		return NewForbiddenError("no estás inscripto en este curso")
	}
	return nil
}

// enrollmentStatus devuelve el estado de la inscripción del
// participante en el curso ("" si no tiene). Si hay más de una (ej. una
// baja y una nueva activa), gana la activa.
func (s *submissionService) enrollmentStatus(ctx context.Context, participanteID, cursoID string) (string, error) {
	enrollments, err := s.enrollmentRepo.ListByParticipant(ctx, participanteID)
	if err != nil {
		return "", err
	}
	estado := ""
	for _, e := range enrollments {
		if e.CursoID != cursoID {
			continue
		}
		if e.Estado == models.EnrollmentEstadoActiva {
			return e.Estado, nil
		}
		if estado == "" || estado == models.EnrollmentEstadoBaja {
			estado = e.Estado
		}
	}
	return estado, nil
}

// buildProgress calcula el progreso formativo (HU-20): actividades
// entregadas / total de actividades del curso.
func buildProgress(participanteID string, total, entregadas int) models.ParticipantProgress {
	pct := 0.0
	if total > 0 {
		pct = float64(entregadas) / float64(total) * 100
	}
	return models.ParticipantProgress{
		ParticipanteID:       participanteID,
		ActividadesTotal:     total,
		ActividadesAsistidas: entregadas,
		PorcentajeAsistencia: pct,
	}
}

// submissionActivityPrefix es el prefijo en el bucket de todos los
// adjuntos de una actividad: entregas/{actividadId}/.
func submissionActivityPrefix(actividadID string) string {
	return "entregas/" + actividadID + "/"
}

func safeFileName(name string) string {
	base := path.Base(strings.ReplaceAll(name, "\\", "/"))
	base = unsafeFileNameChars.ReplaceAllString(base, "_")
	if len(base) > 100 {
		base = base[len(base)-100:]
	}
	if base == "" || base == "." || base == ".." {
		base = "archivo"
	}
	return base
}
