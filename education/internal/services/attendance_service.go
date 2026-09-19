package services

import (
	"context"
	"strings"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// AttendanceService valida las reglas de negocio de "asistencias"
// antes de delegar en el repository (Firestore).
type AttendanceService interface {
	ListByActivity(ctx context.Context, actividadID string) ([]models.Attendance, error)
	// RegisterBatch registra el lote de asistencia de una actividad.
	// cursoID lo resuelve el controller a partir de la actividad (ya la
	// tiene que haber leído para validar que existe/ownership), así este
	// service no necesita volver a leerla de Firestore. bearerToken es
	// el JWT crudo del usuario autenticado (ver middlewares.TokenKey),
	// reenviado a auth-service para validar cada participante. Por cada
	// item del lote, en este orden:
	//  1. el participante existe en auth-service y está aprobado;
	//  2. el participante tiene una inscripción activa en cursoID.
	RegisterBatch(ctx context.Context, actividadID, cursoID string, inputs []models.AttendanceInput, bearerToken string) error
}

type attendanceService struct {
	repo           repositories.AttendanceRepository
	enrollmentRepo repositories.EnrollmentRepository
	authClient     *repositories.AuthServiceClient
}

func NewAttendanceService(repo repositories.AttendanceRepository, enrollmentRepo repositories.EnrollmentRepository, authClient *repositories.AuthServiceClient) AttendanceService {
	return &attendanceService{repo: repo, enrollmentRepo: enrollmentRepo, authClient: authClient}
}

func (s *attendanceService) ListByActivity(ctx context.Context, actividadID string) ([]models.Attendance, error) {
	if strings.TrimSpace(actividadID) == "" {
		return nil, NewValidationError("el id de la actividad es requerido")
	}
	return s.repo.ListByActivity(ctx, actividadID)
}

func (s *attendanceService) RegisterBatch(ctx context.Context, actividadID, cursoID string, inputs []models.AttendanceInput, bearerToken string) error {
	if strings.TrimSpace(actividadID) == "" {
		return NewValidationError("el id de la actividad es requerido")
	}
	if len(inputs) == 0 {
		return NewValidationError("se requiere al menos un registro de asistencia")
	}
	for _, in := range inputs {
		if strings.TrimSpace(in.ParticipanteID) == "" {
			return NewValidationError("cada registro de asistencia requiere 'participanteId'")
		}
	}

	// Inscriptos activos del curso, para validar "participante no
	// inscripto" sin pegarle a Firestore una vez por cada item del lote.
	enrollments, err := s.enrollmentRepo.ListByCourse(ctx, cursoID)
	if err != nil {
		return err
	}
	enrolledActive := make(map[string]bool, len(enrollments))
	for _, e := range enrollments {
		if e.Estado == models.EnrollmentEstadoActiva {
			enrolledActive[e.ParticipanteID] = true
		}
	}

	for _, in := range inputs {
		participant, err := s.authClient.GetParticipant(ctx, in.ParticipanteID, bearerToken)
		if err != nil {
			return err
		}
		if participant == nil {
			return NewValidationError("El participante " + in.ParticipanteID + " no existe")
		}
		if participant.ApprovalStatus != "" && participant.ApprovalStatus != participantApprovalStatusAccepted {
			return NewValidationError("El participante " + in.ParticipanteID + " no está activo")
		}
		if !enrolledActive[in.ParticipanteID] {
			return NewValidationError("El participante " + in.ParticipanteID + " no está inscripto en el curso")
		}
	}

	return s.repo.RegisterBatch(ctx, actividadID, inputs)
}
