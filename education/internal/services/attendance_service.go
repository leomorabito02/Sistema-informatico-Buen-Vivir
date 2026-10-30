package services

import (
	"context"
	"strings"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// AttendanceService expone la consulta de "asistencias" por actividad
// (HU-19). Es de solo lectura: la asistencia por curso ya no se carga a
// mano, la genera cada entrega (ver SubmissionService.Submit).
type AttendanceService interface {
	ListByActivity(ctx context.Context, actividadID string) ([]models.Attendance, error)
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
