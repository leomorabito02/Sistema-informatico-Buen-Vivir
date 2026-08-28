package services

import (
	"context"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// ProgressService calcula, para cada participante inscripto en un
// curso, el porcentaje de actividades del curso a las que asistió
// (presente=true). Compone directamente los repositories de
// actividades/inscripciones/asistencias en vez de sus services: es
// solo lectura y combinación, no necesita las validaciones de esas
// capas.
type ProgressService interface {
	CalculateCourseProgress(ctx context.Context, cursoID string) ([]models.ParticipantProgress, error)
}

type progressService struct {
	activityRepo   repositories.ActivityRepository
	enrollmentRepo repositories.EnrollmentRepository
	attendanceRepo repositories.AttendanceRepository
}

func NewProgressService(
	activityRepo repositories.ActivityRepository,
	enrollmentRepo repositories.EnrollmentRepository,
	attendanceRepo repositories.AttendanceRepository,
) ProgressService {
	return &progressService{
		activityRepo:   activityRepo,
		enrollmentRepo: enrollmentRepo,
		attendanceRepo: attendanceRepo,
	}
}

func (s *progressService) CalculateCourseProgress(ctx context.Context, cursoID string) ([]models.ParticipantProgress, error) {
	activities, err := s.activityRepo.ListByCourse(ctx, cursoID)
	if err != nil {
		return nil, err
	}
	activityIDs := make(map[string]bool, len(activities))
	for _, a := range activities {
		activityIDs[a.ID] = true
	}

	enrollments, err := s.enrollmentRepo.ListByCourse(ctx, cursoID)
	if err != nil {
		return nil, err
	}

	progress := make([]models.ParticipantProgress, 0, len(enrollments))
	for _, e := range enrollments {
		attendances, err := s.attendanceRepo.ListByParticipant(ctx, e.ParticipanteID)
		if err != nil {
			return nil, err
		}

		attended := 0
		for _, at := range attendances {
			if activityIDs[at.ActividadID] && at.Presente {
				attended++
			}
		}

		pct := 0.0
		if len(activities) > 0 {
			pct = float64(attended) / float64(len(activities)) * 100
		}

		progress = append(progress, models.ParticipantProgress{
			ParticipanteID:       e.ParticipanteID,
			ActividadesTotal:     len(activities),
			ActividadesAsistidas: attended,
			PorcentajeAsistencia: pct,
		})
	}

	return progress, nil
}
