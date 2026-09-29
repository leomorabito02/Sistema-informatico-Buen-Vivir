package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"buenvivir/education/internal/models"
)

const enrollmentsCollection = "inscripciones"

// EnrollmentRepository encapsula el acceso a la colección Firestore
// "inscripciones".
type EnrollmentRepository interface {
	ListByParticipants(ctx context.Context, participantIDs []string) ([]models.Enrollment, error)
	ListByCourse(ctx context.Context, cursoID string) ([]models.Enrollment, error)
	// ListByParticipant devuelve todas las inscripciones (de cualquier
	// estado) de un participante: es la base de "mis cursos" del rol
	// participante.
	ListByParticipant(ctx context.Context, participanteID string) ([]models.Enrollment, error)
	// Create persiste una nueva inscripción. No valida duplicados ni
	// cupo — eso es responsabilidad de EnrollmentService.Enroll, que
	// lista las inscripciones activas del curso antes de llamar acá.
	// Fuerza fecha_inscripcion a "ahora" y, si Estado viene vacío, lo
	// deja en "activa".
	Create(ctx context.Context, e models.Enrollment) (*models.Enrollment, error)
}

type enrollmentRepository struct {
	client *firestore.Client
}

func NewEnrollmentRepository(client *firestore.Client) EnrollmentRepository {
	return &enrollmentRepository{client: client}
}

func (r *enrollmentRepository) ListByCourse(ctx context.Context, cursoID string) ([]models.Enrollment, error) {
	return r.list(ctx, "curso_id", cursoID)
}

func (r *enrollmentRepository) ListByParticipant(ctx context.Context, participanteID string) ([]models.Enrollment, error) {
	return r.list(ctx, "participante_id", participanteID)
}

func (r *enrollmentRepository) list(ctx context.Context, field, value string) ([]models.Enrollment, error) {
	iter := r.client.Collection(enrollmentsCollection).Where(field, "==", value).Documents(ctx)
	defer iter.Stop()

	enrollments := make([]models.Enrollment, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listando inscripciones: %w", err)
		}
		var e models.Enrollment
		if err := doc.DataTo(&e); err != nil {
			return nil, fmt.Errorf("error leyendo inscripcion: %w", err)
		}
		e.ID = doc.Ref.ID
		enrollments = append(enrollments, e)
	}
	return enrollments, nil
}

func (r *enrollmentRepository) ListByParticipants(ctx context.Context, participantIDs []string) ([]models.Enrollment, error) {
	if len(participantIDs) == 0 {
		return []models.Enrollment{}, nil
	}
	enrollments := make([]models.Enrollment, 0)
	for i := 0; i < len(participantIDs); i += 30 {
		end := i + 30
		if end > len(participantIDs) {
			end = len(participantIDs)
		}
		chunk := participantIDs[i:end]
		iter := r.client.Collection(enrollmentsCollection).Where("participante_id", "in", chunk).Documents(ctx)
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				iter.Stop()
				return nil, fmt.Errorf("error listando inscripciones bulk: %w", err)
			}
			var e models.Enrollment
			if err := doc.DataTo(&e); err != nil {
				iter.Stop()
				return nil, fmt.Errorf("error leyendo inscripcion bulk: %w", err)
			}
			e.ID = doc.Ref.ID
			enrollments = append(enrollments, e)
		}
		iter.Stop()
	}
	return enrollments, nil
}

func (r *enrollmentRepository) Create(ctx context.Context, e models.Enrollment) (*models.Enrollment, error) {
	e.FechaInscripcion = time.Now().UTC()
	if e.Estado == "" {
		e.Estado = models.EnrollmentEstadoActiva
	}

	ref, _, err := r.client.Collection(enrollmentsCollection).Add(ctx, e)
	if err != nil {
		return nil, fmt.Errorf("error creando inscripcion: %w", err)
	}
	e.ID = ref.ID
	return &e, nil
}
