package services

import (
	"context"
	"strings"

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
		return enrollments, nil // los nombres son un dato secundario
	}
	for i := range enrollments {
		if p, ok := directorio[enrollments[i].ParticipanteID]; ok {
			enrollments[i].Nombre = p.Nombre
			enrollments[i].Apellido = p.Apellido
		}
	}
	return enrollments, nil
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
	return s.repo.ListByParticipant(ctx, participantID)
}

func (s *enrollmentService) GetBulkUserCourses(ctx context.Context, userIDs []string) ([]models.UserCourseSummary, error) {
	if len(userIDs) == 0 {
		return []models.UserCourseSummary{}, nil
	}

	summaryMap := make(map[string][]string)
	for _, id := range userIDs {
		summaryMap[id] = make([]string, 0)
	}

	// 1. Get enrollments for all these users
	enrollments, err := s.repo.ListByParticipants(ctx, userIDs)
	if err == nil {
		// fetch course titles for these enrollments
		// cache course titles
		courseTitles := make(map[string]string)
		for _, e := range enrollments {
			if e.Estado != models.EnrollmentEstadoActiva {
				continue
			}
			title, ok := courseTitles[e.CursoID]
			if !ok {
				course, cerr := s.courseRepo.Get(ctx, e.CursoID)
				if cerr == nil && course != nil {
					title = course.Title
					courseTitles[e.CursoID] = title
				} else {
					title = "Curso " + e.CursoID
				}
			}
			summaryMap[e.ParticipanteID] = append(summaryMap[e.ParticipanteID], title)
		}
	}

	// 2. Get courses for these teachers
	teacherCourses, err := s.courseRepo.ListByResponsibleTeachers(ctx, userIDs)
	if err == nil {
		for _, c := range teacherCourses {
			summaryMap[c.ResponsibleTeacher] = append(summaryMap[c.ResponsibleTeacher], c.Title)
		}
	}

	var result []models.UserCourseSummary
	for id, courses := range summaryMap {
		result = append(result, models.UserCourseSummary{
			UserID:  id,
			Courses: courses,
		})
	}

	return result, nil
}

func (s *enrollmentService) Enroll(ctx context.Context, participanteID, cursoID, bearerToken string) (*models.Enrollment, error) {
	participanteID = strings.TrimSpace(participanteID)
	if participanteID == "" {
		return nil, NewValidationError("participanteId es obligatorio")
	}
	if strings.TrimSpace(cursoID) == "" {
		return nil, NewValidationError("cursoId es obligatorio")
	}

	// 1. Curso: solo se puede inscribir a uno publicado y visible en el
	// catálogo (la CDA, a diferencia de la CMA, no devuelve borradores).
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
	if s.bajas != nil {
		dadoDeBaja, err := s.bajas.EstaDadoDeBaja(ctx, cursoID)
		if err != nil {
			return nil, err
		}
		if dadoDeBaja {
			return nil, NewValidationError("El curso fue dado de baja y no acepta inscripciones")
		}
	}

	// 2. Participante: tiene que existir en auth-service y estar
	// aprobado.
	participant, err := s.authClient.GetParticipant(ctx, participanteID, bearerToken)
	if err != nil {
		return nil, err
	}
	if participant == nil {
		return nil, NewValidationError("El participante no existe")
	}
	if !isParticipantApproved(participant.ApprovalStatus) {
		return nil, NewValidationError("El participante no está activo")
	}

	// 3-4. Duplicado y cupo.
	existing, err := s.repo.ListByCourse(ctx, cursoID)
	if err != nil {
		return nil, err
	}

	activeCount := 0
	for _, e := range existing {
		if e.Estado != models.EnrollmentEstadoActiva {
			continue
		}
		activeCount++
		if e.ParticipanteID == participanteID {
			return nil, NewValidationError("El participante ya se encuentra inscripto en el curso")
		}
	}

	if course.Capacity != nil && activeCount >= *course.Capacity {
		return nil, NewValidationError("El curso no tiene cupo disponible")
	}

	return s.repo.Create(ctx, models.Enrollment{
		ParticipanteID: participanteID,
		CursoID:        cursoID,
		Estado:         models.EnrollmentEstadoActiva,
		CourseTitle:    course.Title,
	})
}
