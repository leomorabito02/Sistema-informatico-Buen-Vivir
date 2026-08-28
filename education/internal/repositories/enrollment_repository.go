package repositories

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"buenvivir/education/internal/models"
)

const enrollmentsCollection = "inscripciones"

// EnrollmentRepository encapsula el acceso a la colección Firestore
// "inscripciones". Solo implementa lectura: ningún endpoint de esta
// tarea necesita crear/editar inscripciones (esa colección la puebla
// otro módulo, fuera de este alcance) — se evita así código sin uso.
type EnrollmentRepository interface {
	ListByCourse(ctx context.Context, cursoID string) ([]models.Enrollment, error)
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
