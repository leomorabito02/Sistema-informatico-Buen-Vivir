package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"buenvivir/education/internal/audit"
	"buenvivir/education/internal/events/producers"
	"buenvivir/education/internal/models"
)

const generalAttendanceCollection = "asistencia_general"

// GeneralAttendanceRepository encapsula el acceso a la colección
// Firestore "asistencia_general" (HU-19bis/RF-64).
type GeneralAttendanceRepository interface {
	ListByDate(ctx context.Context, fecha string) ([]models.GeneralAttendance, error)
	// RegisterBatch crea o actualiza un documento por input con id {fecha}_{participanteId}.
	RegisterBatch(ctx context.Context, fecha string, inputs []models.GeneralAttendanceInput) error
}

type generalAttendanceRepository struct {
	client *firestore.Client
	events *producers.Educacion
}

// NewGeneralAttendanceRepository recibe opcionalmente el productor de
// eventos (asistencia_general.registrada / .actualizada).
func NewGeneralAttendanceRepository(client *firestore.Client, events ...*producers.Educacion) GeneralAttendanceRepository {
	return &generalAttendanceRepository{client: client, events: firstProducer(events)}
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
	refs := make([]*firestore.DocumentRef, len(inputs))
	docs := make([]models.GeneralAttendance, len(inputs))
	for i, in := range inputs {
		refs[i] = r.client.Collection(generalAttendanceCollection).Doc(fecha + "_" + in.ParticipanteID)
		docs[i] = models.GeneralAttendance{
			ParticipanteID: in.ParticipanteID,
			Fecha:          fecha,
			Presente:       in.Presente,
			Observacion:    in.Observacion,
			UpdatedAt:      now,
		}
	}

	// Transacción: lee los registros previos para saber si es carga nueva o edición, y
	// escribe junto con el registro la auditoría y el evento. Guardar lo mismo no deja rastro.
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		previous, err := tx.GetAll(refs)
		if err != nil {
			return err
		}
		for i, ref := range refs {
			if err := tx.Set(ref, docs[i]); err != nil {
				return err
			}
			nueva := !previous[i].Exists()
			if !nueva && !changed(previous[i], docs[i]) {
				continue
			}
			accion := audit.AccionActualizar
			if nueva {
				accion = audit.AccionRegistrar
			}
			if err := audit.RecordTx(ctx, r.client, tx, accion, audit.RecursoAsistenciaGeneral, ref.ID); err != nil {
				return err
			}
			if err := r.events.AsistenciaGeneral(ctx, tx, ref.ID, docs[i], nueva); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error registrando asistencia general: %w", err)
	}
	r.events.Notify()
	return nil
}

func changed(prev *firestore.DocumentSnapshot, next models.GeneralAttendance) bool {
	var old models.GeneralAttendance
	if err := prev.DataTo(&old); err != nil {
		return true
	}
	return old.Presente != next.Presente || old.Observacion != next.Observacion
}
