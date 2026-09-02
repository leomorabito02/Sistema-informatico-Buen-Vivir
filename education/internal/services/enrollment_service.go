package services

import (
	"context"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// EnrollmentService expone lectura de inscripciones (Firestore). No
// tiene reglas de negocio propias todavía: es un passthrough al
// repository, dejado como interfaz para mockear en tests futuros.
type EnrollmentService interface {
	ListByCourse(ctx context.Context, cursoID string) ([]models.Enrollment, error)
}

type enrollmentService struct {
	repo repositories.EnrollmentRepository
}

func NewEnrollmentService(repo repositories.EnrollmentRepository) EnrollmentService {
	return &enrollmentService{repo: repo}
}

func (s *enrollmentService) ListByCourse(ctx context.Context, cursoID string) ([]models.Enrollment, error) {
	return s.repo.ListByCourse(ctx, cursoID)
}
