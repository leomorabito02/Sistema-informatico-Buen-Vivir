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
	ListByCourse(ctx context.Context, cursoID string) ([]models.Enrollment, error)
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
	iter := r.client.Collection(enrollmentsCollection).Where("curso_id", "==", cursoID).Documents(ctx)
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
