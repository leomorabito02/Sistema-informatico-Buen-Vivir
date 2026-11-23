package services

import (
	"context"
	"strings"

	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// EnrollmentService valida las reglas de negocio de inscripción
// (HU-18/RF-18) antes de delegar en el repository de Firestore.
type EnrollmentService interface {
	ListByCourse(ctx context.Context, cursoID string) ([]models.Enrollment, error)
	// ListByCourseConNombres es ListByCourse con nombre y apellido de
	// cada participante, resueltos contra auth-service con bearerToken.
	// Si auth-service falla, las inscripciones salen igual, sin nombre.
	ListByCourseConNombres(ctx context.Context, cursoID, bearerToken string) ([]models.Enrollment, error)
	// CountActive cuenta las inscripciones activas del curso: es el
	// aviso que se muestra antes de darlo de baja (HU-16).
	CountActive(ctx context.Context, cursoID string) (int, error)
	ListByParticipant(ctx context.Context, participantID string) ([]models.Enrollment, error)
	GetBulkUserCourses(ctx context.Context, userIDs []string) ([]models.UserCourseSummary, error)
	// Enroll inscribe a un participante en un curso publicado.
	// bearerToken es el JWT crudo del usuario autenticado en education
	// (ver middlewares.TokenKey), que se reenvía a auth-service para
	// confirmar que el participante existe y está aprobado. Reglas de
	// negocio, en este orden:
	//  1. El curso tiene que existir y estar publicado (CDA). Si no está
	//     publicado, se distingue "no existe" de "existe pero inactivo"
	//     consultando la CMA.
	//  2. El participante tiene que existir en auth-service y estar
	//     aprobado.
	//  3. El participante no puede tener ya una inscripción activa en
	//     ese curso.
	//  4. Si el curso tiene Capacity fijado, el cupo no puede estar
	//     completo (inscripciones activas >= capacity).
	// Un curso dado de baja (HU-16) sigue publicado pero no acepta
	// inscripciones nuevas.
	Enroll(ctx context.Context, participanteID, cursoID, bearerToken string) (*models.Enrollment, error)
}

type enrollmentService struct {
	repo       repositories.EnrollmentRepository
	courseRepo repositories.CourseRepository
	authClient *repositories.AuthServiceClient
	bajas      repositories.CursoBajaRepository
}

// NewEnrollmentService recibe opcionalmente el repositorio de bajas de
// cursos (HU-16): con él, no se inscribe en un curso dado de baja.
func NewEnrollmentService(repo repositories.EnrollmentRepository, courseRepo repositories.CourseRepository, authClient *repositories.AuthServiceClient, bajas ...repositories.CursoBajaRepository) EnrollmentService {
	s := &enrollmentService{repo: repo, courseRepo: courseRepo, authClient: authClient}
	if len(bajas) > 0 {
		s.bajas = bajas[0]
	}
	return s
}

func (s *enrollmentService) ListByCourse(ctx context.Context, cursoID string) ([]models.Enrollment, error) {
	return s.repo.ListByCourse(ctx, cursoID)
}

func (s *enrollmentService) ListByCourseConNombres(ctx context.Context, cursoID, bearerToken string) ([]models.Enrollment, error) {
	enrollments, err := s.repo.ListByCourse(ctx, cursoID)
	if err != nil || len(enrollments) == 0 {
		return enrollments, err
	}
	directorio, err := directorioParticipantes(ctx, s.authClient, bearerToken)
	if err != nil {
		// Un docente no puede listar todos los participantes en auth
		// (403): se piden los nombres de a uno, solo los de este curso.
		// Los nombres son un dato secundario: si falla, queda el id.
		s.nombresDeAUno(ctx, enrollments, bearerToken)
		return enrollments, nil
	}
	for i := range enrollments {
		if p, ok := directorio[enrollments[i].ParticipanteID]; ok {
			enrollments[i].Nombre = p.Nombre
			enrollments[i].Apellido = p.Apellido
		}
	}
	return enrollments, nil
}

// nombresDeAUno completa nombre y apellido pidiendo a auth cada
// participante por separado (GET /directory/participants/{id}, permitido
// al docente). Cada participante se pide una sola vez.
func (s *enrollmentService) nombresDeAUno(ctx context.Context, enrollments []models.Enrollment, bearerToken string) {
	if s.authClient == nil {
		return
	}
	cache := map[string]*models.AuthUser{}
	for i := range enrollments {
		id := enrollments[i].ParticipanteID
		u, visto := cache[id]
		if !visto {
			if got, err := s.authClient.GetParticipantName(ctx, id, bearerToken); err == nil {
				u = &got
			}
			cache[id] = u
		}
		if u != nil {
			enrollments[i].Nombre = u.FirstName
			enrollments[i].Apellido = u.LastName
		}
	}
}

func (s *enrollmentService) CountActive(ctx context.Context, cursoID string) (int, error) {
	if strings.TrimSpace(cursoID) == "" {
		return 0, NewValidationError("cursoId es obligatorio")
	}
	enrollments, err := s.repo.ListByCourse(ctx, cursoID)
	if err != nil {
		return 0, err
	}
	return ContarInscripcionesActivas(enrollments), nil
}

func (s *enrollmentService) ListByParticipant(ctx context.Context, participantID string) ([]models.Enrollment, error) {
	enrollments, err := s.repo.ListByParticipant(ctx, participantID)
	if err != nil {
		return nil, err
	}
	if len(enrollments) == 0 {
		enrollments = s.resolveEnrollmentsByUserID(ctx, participantID)
	}
	s.populateEnrollmentTitles(ctx, enrollments)
	return enrollments, nil
}

func (s *enrollmentService) resolveEnrollmentsByUserID(ctx context.Context, id string) []models.Enrollment {
	if s.authClient == nil {
		return nil
	}
	bearerToken, _ := ctx.Value(middlewares.TokenKey).(string)
	participant, err := s.authClient.GetParticipant(ctx, id, bearerToken)
	if err != nil || participant == nil || participant.ParticipantID == "" || participant.ParticipantID == id {
		return nil
	}
	resolved, err := s.repo.ListByParticipant(ctx, participant.ParticipantID)
	if err != nil {
		return nil
	}
	return resolved
}

func (s *enrollmentService) populateEnrollmentTitles(ctx context.Context, enrollments []models.Enrollment) {
	if len(enrollments) == 0 || s.courseRepo == nil {
		return
	}
	courseTitles := make(map[string]string)
	for i := range enrollments {
		if enrollments[i].CourseTitle == "" && enrollments[i].CursoID != "" {
			enrollments[i].CourseTitle = resolveCourseTitle(ctx, s.courseRepo, courseTitles, enrollments[i].CursoID)
		}
	}
}

func (s *enrollmentService) GetBulkUserCourses(ctx context.Context, userIDs []string) ([]models.UserCourseSummary, error) {
	if len(userIDs) == 0 {
		return []models.UserCourseSummary{}, nil
	}

	summaryMap := make(map[string][]string, len(userIDs))
	requestedUsers := make(map[string]bool, len(userIDs))
	for _, id := range userIDs {
		summaryMap[id] = make([]string, 0)
		requestedUsers[id] = true
	}

	lookupIDs, participantToUser := s.buildBulkLookup(ctx, userIDs, requestedUsers)
	s.addEnrollmentCourses(ctx, lookupIDs, requestedUsers, participantToUser, summaryMap)
	s.addTeacherCourses(ctx, userIDs, requestedUsers, summaryMap)

	result := make([]models.UserCourseSummary, 0, len(summaryMap))
	for id, courses := range summaryMap {
		result = append(result, models.UserCourseSummary{
			UserID:  id,
			Courses: courses,
		})
	}
	return result, nil
}

func (s *enrollmentService) buildBulkLookup(ctx context.Context, userIDs []string, requestedUsers map[string]bool) ([]string, map[string]string) {
	lookupIDs := append([]string(nil), userIDs...)
	participantToUser := make(map[string]string)
	if s.authClient == nil {
		return lookupIDs, participantToUser
	}

	bearerToken, _ := ctx.Value(middlewares.TokenKey).(string)
	participants, err := s.authClient.ListParticipants(ctx, bearerToken)
	if err != nil {
		return lookupIDs, participantToUser
	}

	seen := make(map[string]bool, len(lookupIDs))
	for _, id := range lookupIDs {
		seen[id] = true
	}
	for _, p := range participants {
		if !requestedUsers[p.UserID] || p.ParticipantID == "" {
			continue
		}
		participantToUser[p.ParticipantID] = p.UserID
		if !seen[p.ParticipantID] {
			seen[p.ParticipantID] = true
			lookupIDs = append(lookupIDs, p.ParticipantID)
		}
	}
	return lookupIDs, participantToUser
}

func (s *enrollmentService) addEnrollmentCourses(
	ctx context.Context,
	lookupIDs []string,
	requestedUsers map[string]bool,
	participantToUser map[string]string,
	summaryMap map[string][]string,
) {
	enrollments, err := s.repo.ListByParticipants(ctx, lookupIDs)
	if err != nil {
		return
	}
	courseTitles := make(map[string]string)
	for _, e := range enrollments {
		if e.Estado != models.EnrollmentEstadoActiva {
			continue
		}
		targetUserID := resolveTargetUserID(e.ParticipanteID, participantToUser)
		if !requestedUsers[targetUserID] {
			continue
		}
		title := resolveCourseTitle(ctx, s.courseRepo, courseTitles, e.CursoID)
		summaryMap[targetUserID] = appendUnique(summaryMap[targetUserID], title)
	}
}

func (s *enrollmentService) addTeacherCourses(
	ctx context.Context,
	userIDs []string,
	requestedUsers map[string]bool,
	summaryMap map[string][]string,
) {
	if s.courseRepo == nil {
		return
	}
	teacherCourses, err := s.courseRepo.ListByResponsibleTeachers(ctx, userIDs)
	if err != nil {
		return
	}
	for _, c := range teacherCourses {
		if requestedUsers[c.ResponsibleTeacher] {
			summaryMap[c.ResponsibleTeacher] = appendUnique(summaryMap[c.ResponsibleTeacher], c.Title)
		}
	}
}

func resolveTargetUserID(participanteID string, participantToUser map[string]string) string {
	if mappedUserID, ok := participantToUser[participanteID]; ok {
		return mappedUserID
	}
	return participanteID
}

func resolveCourseTitle(ctx context.Context, courseRepo repositories.CourseRepository, cache map[string]string, cursoID string) string {
	if title, ok := cache[cursoID]; ok {
		return title
	}
	title := "Curso " + cursoID
	if courseRepo != nil {
		if course, err := courseRepo.Get(ctx, cursoID); err == nil && course != nil && course.Title != "" {
			title = course.Title
		}
	}
	cache[cursoID] = title
	return title
}

func appendUnique(items []string, value string) []string {
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func (s *enrollmentService) Enroll(ctx context.Context, participanteID, cursoID, bearerToken string) (*models.Enrollment, error) {
	participanteID = strings.TrimSpace(participanteID)
	if participanteID == "" {
		return nil, NewValidationError("participanteId es obligatorio")
	}
	if strings.TrimSpace(cursoID) == "" {
		return nil, NewValidationError("cursoId es obligatorio")
	}

	course, err := s.validateEnrollableCourse(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	if err := s.validateApprovedParticipant(ctx, participanteID, bearerToken); err != nil {
		return nil, err
	}
	if err := s.ensureEnrollCapacity(ctx, course, participanteID); err != nil {
		return nil, err
	}

	return s.repo.Create(ctx, models.Enrollment{
		ParticipanteID: participanteID,
		CursoID:        cursoID,
		Estado:         models.EnrollmentEstadoActiva,
		CourseTitle:    course.Title,
		TeacherID:      course.ResponsibleTeacher,
	})
}

func (s *enrollmentService) validateEnrollableCourse(ctx context.Context, cursoID string) (*models.Course, error) {
	course, err := s.courseRepo.Get(ctx, cursoID)
	if err != nil {
		if !repositories.IsNotFound(err) {
			return nil, err
		}
		exists, existsErr := s.courseRepo.ExistsInManagement(ctx, cursoID)
		if existsErr != nil {
			return nil, existsErr
		}
		if exists {
			return nil, NewValidationError("El curso no está disponible (inactivo o no publicado)")
		}
		return nil, NewValidationError("El curso no existe")
	}
	if s.bajas == nil {
		return course, nil
	}
	dadoDeBaja, err := s.bajas.EstaDadoDeBaja(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	if dadoDeBaja {
		return nil, NewValidationError("El curso fue dado de baja y no acepta inscripciones")
	}
	return course, nil
}

func (s *enrollmentService) validateApprovedParticipant(ctx context.Context, participanteID, bearerToken string) error {
	participant, err := s.authClient.GetParticipant(ctx, participanteID, bearerToken)
	if err != nil {
		return err
	}
	if participant == nil {
		return NewValidationError("El participante no existe")
	}
	if !isParticipantApproved(participant.ApprovalStatus) {
		return NewValidationError("El participante no está activo")
	}
	return nil
}

func (s *enrollmentService) ensureEnrollCapacity(ctx context.Context, course *models.Course, participanteID string) error {
	existing, err := s.repo.ListByCourse(ctx, course.ID)
	if err != nil {
		return err
	}

	activeCount := 0
	for _, e := range existing {
		if e.Estado != models.EnrollmentEstadoActiva {
			continue
		}
		activeCount++
		if e.ParticipanteID == participanteID {
			return NewValidationError("El participante ya se encuentra inscripto en el curso")
		}
	}

	if course.Capacity != nil && activeCount >= *course.Capacity {
		return NewValidationError("El curso no tiene cupo disponible")
	}
	return nil
}
