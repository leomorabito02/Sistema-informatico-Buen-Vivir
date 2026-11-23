package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"buenvivir/education/internal/events/producers"
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
	// Create guarda una inscripción nueva (fecha = ahora, estado "activa" por defecto).
	// Los duplicados y el cupo los valida EnrollmentService.Enroll.
	Create(ctx context.Context, e models.Enrollment) (*models.Enrollment, error)
	// DeactivateByParticipant pasa a "baja" las inscripciones activas de un participante
	// dado de baja en auth-service, conservando el historial. Devuelve cuántas cambió.
	DeactivateByParticipant(ctx context.Context, participanteID, motivo string) (int, error)
}

type enrollmentRepository struct {
	client *firestore.Client
	events *producers.Educacion
}

// NewEnrollmentRepository recibe opcionalmente el productor de eventos:
// sin él (o nil) no se publica nada.
func NewEnrollmentRepository(client *firestore.Client, events ...*producers.Educacion) EnrollmentRepository {
	return &enrollmentRepository{client: client, events: firstProducer(events)}
}

// firstProducer resuelve el parámetro opcional de productor de eventos
// de los constructores de repositorios.
func firstProducer(ps []*producers.Educacion) *producers.Educacion {
	if len(ps) > 0 {
		return ps[0]
	}
	return nil
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

	// La inscripción y su evento (inscripcion.creada) se escriben en la
	// misma transacción: o quedan las dos cosas, o ninguna (outbox).
	ref := r.client.Collection(enrollmentsCollection).NewDoc()
	e.ID = ref.ID
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		if err := tx.Create(ref, e); err != nil {
			return err
		}
		return r.events.InscripcionCreada(ctx, tx, e)
	})
	if err != nil {
		return nil, fmt.Errorf("error creando inscripcion: %w", err)
	}
	r.events.Notify()
	return &e, nil
}

func (r *enrollmentRepository) DeactivateByParticipant(ctx context.Context, participanteID, motivo string) (int, error) {
	docs, err := r.client.Collection(enrollmentsCollection).
		Where("participante_id", "==", participanteID).
		Where("estado", "==", models.EnrollmentEstadoActiva).
		Documents(ctx).GetAll()
	if err != nil {
		return 0, fmt.Errorf("error listando inscripciones activas: %w", err)
	}
	if len(docs) == 0 {
		return 0, nil
	}
	now := time.Now().UTC()
	batch := r.client.Batch() //nolint:staticcheck // lote chico y atómico, mismo criterio que AttendanceRepository
	for _, d := range docs {
		batch.Update(d.Ref, []firestore.Update{
			{Path: "estado", Value: models.EnrollmentEstadoBaja},
			{Path: "motivo_baja", Value: motivo},
			{Path: "fecha_baja", Value: now},
		})
	}
	if _, err := batch.Commit(ctx); err != nil {
		return 0, fmt.Errorf("error dando de baja inscripciones: %w", err)
	}
	return len(docs), nil
}
