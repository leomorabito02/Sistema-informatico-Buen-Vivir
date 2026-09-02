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
	RegisterBatch(ctx context.Context, actividadID string, inputs []models.AttendanceInput) error
}

type attendanceService struct {
	repo repositories.AttendanceRepository
}

func NewAttendanceService(repo repositories.AttendanceRepository) AttendanceService {
	return &attendanceService{repo: repo}
}

func (s *attendanceService) ListByActivity(ctx context.Context, actividadID string) ([]models.Attendance, error) {
	if strings.TrimSpace(actividadID) == "" {
		return nil, NewValidationError("el id de la actividad es requerido")
	}
	return s.repo.ListByActivity(ctx, actividadID)
}

func (s *attendanceService) RegisterBatch(ctx context.Context, actividadID string, inputs []models.AttendanceInput) error {
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
	return s.repo.RegisterBatch(ctx, actividadID, inputs)
}
