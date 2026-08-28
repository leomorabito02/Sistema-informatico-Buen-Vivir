package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"buenvivir/education/internal/models"
)

const attendancesCollection = "asistencias"

// AttendanceRepository encapsula el acceso a la colección Firestore
// "asistencias".
type AttendanceRepository interface {
	ListByActivity(ctx context.Context, actividadID string) ([]models.Attendance, error)
	ListByParticipant(ctx context.Context, participanteID string) ([]models.Attendance, error)
	// RegisterBatch registra un documento de asistencia por cada
	// input, con id determinístico {actividadId}_{participanteId}: si
	// el docente corrige y vuelve a mandar la misma actividad para el
	// mismo participante, se sobrescribe en vez de duplicarse.
	RegisterBatch(ctx context.Context, actividadID string, inputs []models.AttendanceInput) error
}

type attendanceRepository struct {
	client *firestore.Client
}

func NewAttendanceRepository(client *firestore.Client) AttendanceRepository {
	return &attendanceRepository{client: client}
}

func (r *attendanceRepository) ListByActivity(ctx context.Context, actividadID string) ([]models.Attendance, error) {
	return r.list(ctx, "actividad_id", actividadID)
}

func (r *attendanceRepository) ListByParticipant(ctx context.Context, participanteID string) ([]models.Attendance, error) {
	return r.list(ctx, "participante_id", participanteID)
}

func (r *attendanceRepository) list(ctx context.Context, field, value string) ([]models.Attendance, error) {
	iter := r.client.Collection(attendancesCollection).Where(field, "==", value).Documents(ctx)
	defer iter.Stop()

	attendances := make([]models.Attendance, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listando asistencias: %w", err)
		}
		var a models.Attendance
		if err := doc.DataTo(&a); err != nil {
			return nil, fmt.Errorf("error leyendo asistencia: %w", err)
		}
		a.ID = doc.Ref.ID
		attendances = append(attendances, a)
	}
	return attendances, nil
}

func (r *attendanceRepository) RegisterBatch(ctx context.Context, actividadID string, inputs []models.AttendanceInput) error {
	if len(inputs) == 0 {
		return nil
	}

	now := time.Now().UTC()
	batch := r.client.Batch()
	for _, in := range inputs {
		docID := actividadID + "_" + in.ParticipanteID
		a := models.Attendance{
			ParticipanteID: in.ParticipanteID,
			ActividadID:    actividadID,
			Presente:       in.Presente,
			Fecha:          now,
		}
		batch.Set(r.client.Collection(attendancesCollection).Doc(docID), a)
	}

	if _, err := batch.Commit(ctx); err != nil {
		return fmt.Errorf("error registrando asistencia: %w", err)
	}
	return nil
}
