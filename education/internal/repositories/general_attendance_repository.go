package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"buenvivir/education/internal/models"
)

const generalAttendanceCollection = "asistencia_general"

// GeneralAttendanceRepository encapsula el acceso a la colección
// Firestore "asistencia_general" (HU-19bis/RF-64).
type GeneralAttendanceRepository interface {
	ListByDate(ctx context.Context, fecha string) ([]models.GeneralAttendance, error)
	// RegisterBatch crea o actualiza (upsert) un documento por cada
	// input, con id determinístico "{fecha}_{participanteId}": volver a
	// cargar al mismo participante en la misma fecha sobrescribe el
	// registro existente en vez de duplicarlo (mismo criterio que
	// AttendanceRepository.RegisterBatch).
	RegisterBatch(ctx context.Context, fecha string, inputs []models.GeneralAttendanceInput) error
}

type generalAttendanceRepository struct {
	client *firestore.Client
}

func NewGeneralAttendanceRepository(client *firestore.Client) GeneralAttendanceRepository {
	return &generalAttendanceRepository{client: client}
}

func (r *generalAttendanceRepository) ListByDate(ctx context.Context, fecha string) ([]models.GeneralAttendance, error) {
	iter := r.client.Collection(generalAttendanceCollection).Where("fecha", "==", fecha).Documents(ctx)
	defer iter.Stop()

	out := make([]models.GeneralAttendance, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listando asistencia general: %w", err)
		}
		var a models.GeneralAttendance
		if err := doc.DataTo(&a); err != nil {
			return nil, fmt.Errorf("error leyendo asistencia general: %w", err)
		}
		a.ID = doc.Ref.ID
		out = append(out, a)
	}
	return out, nil
}

func (r *generalAttendanceRepository) RegisterBatch(ctx context.Context, fecha string, inputs []models.GeneralAttendanceInput) error {
	if len(inputs) == 0 {
		return nil
	}

	now := time.Now().UTC()
	batch := r.client.Batch() //nolint:staticcheck // WriteBatch alcanza para este lote chico y atómico, mismo criterio que AttendanceRepository
	for _, in := range inputs {
		docID := fecha + "_" + in.ParticipanteID
		a := models.GeneralAttendance{
			ParticipanteID: in.ParticipanteID,
			Fecha:          fecha,
			Presente:       in.Presente,
			Observacion:    in.Observacion,
			UpdatedAt:      now,
		}
		batch.Set(r.client.Collection(generalAttendanceCollection).Doc(docID), a)
	}

	if _, err := batch.Commit(ctx); err != nil {
		return fmt.Errorf("error registrando asistencia general: %w", err)
	}
	return nil
}
