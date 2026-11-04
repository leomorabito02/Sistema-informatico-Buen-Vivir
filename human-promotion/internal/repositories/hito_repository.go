package repositories

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"buenvivir/human-promotion/internal/audit"
	"buenvivir/human-promotion/internal/events"
	"buenvivir/human-promotion/internal/events/producers"
	"buenvivir/human-promotion/internal/models"
)

const hitosCollection = "hitos"

// HitoRepository encapsula el acceso a la colección de hitos. Como en
// FichaRepository, la descripción llega y sale ya cifrada.
type HitoRepository interface {
	Create(ctx context.Context, doc models.HitoDocument) (string, error)
	GetByID(ctx context.Context, id string) (*models.HitoDocument, error)
	Update(ctx context.Context, id string, doc models.HitoDocument) error
	// ListByFicha devuelve los hitos de la ficha con su id, sin orden
	// garantizado: el orden cronológico lo arma el servicio, para no
	// depender de un índice compuesto (ficha_id, fecha) en Firestore.
	ListByFicha(ctx context.Context, fichaID string) (map[string]models.HitoDocument, error)
}

type hitoRepository struct {
	client *firestore.Client
	events *producers.PromocionHumana
}

// NewHitoRepository recibe opcionalmente el productor de eventos (sin
// él no se publica nada; la auditoría se escribe siempre).
func NewHitoRepository(client *firestore.Client, events ...*producers.PromocionHumana) HitoRepository {
	return &hitoRepository{client: client, events: firstProducer(events)}
}

// participanteDeFicha lee, dentro de la transacción, el participante de
// la ficha del hito (es la ordering key de los eventos de hitos).
func (r *hitoRepository) participanteDeFicha(tx *firestore.Transaction, fichaID string) (string, error) {
	snap, err := tx.Get(r.client.Collection(fichasCollection).Doc(fichaID))
	if status.Code(err) == codes.NotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	participanteID, _ := snap.Data()["participante_id"].(string)
	return participanteID, nil
}

func hitoPayload(id, participanteID string, doc models.HitoDocument) events.HitoPayload {
	return events.HitoPayload{
		HitoID: id, FichaID: doc.FichaID, ParticipanteID: participanteID,
		Fecha: doc.Fecha, ResponsableID: doc.ResponsableID, VisibleParticipante: doc.VisibleParticipante,
	}
}

func (r *hitoRepository) Create(ctx context.Context, doc models.HitoDocument) (string, error) {
	ref := r.client.Collection(hitosCollection).NewDoc()
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		participanteID, err := r.participanteDeFicha(tx, doc.FichaID)
		if err != nil {
			return err
		}
		if err := tx.Create(ref, doc); err != nil {
			return err
		}
		if err := r.events.Hito(ctx, tx, true, hitoPayload(ref.ID, participanteID, doc)); err != nil {
			return err
		}
		return audit.RecordTx(ctx, r.client, tx, audit.AccionCrear, audit.RecursoHito, ref.ID)
	})
	if err != nil {
		return "", fmt.Errorf("error creando el hito: %w", err)
	}
	r.events.Notify()
	return ref.ID, nil
}

func (r *hitoRepository) GetByID(ctx context.Context, id string) (*models.HitoDocument, error) {
	snap, err := r.client.Collection(hitosCollection).Doc(id).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error leyendo el hito: %w", err)
	}
	var doc models.HitoDocument
	if err := snap.DataTo(&doc); err != nil {
		return nil, fmt.Errorf("error leyendo el hito: %w", err)
	}
	return &doc, nil
}

func (r *hitoRepository) Update(ctx context.Context, id string, doc models.HitoDocument) error {
	ref := r.client.Collection(hitosCollection).Doc(id)
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		participanteID, err := r.participanteDeFicha(tx, doc.FichaID)
		if err != nil {
			return err
		}
		if err := tx.Update(ref, []firestore.Update{
			{Path: "fecha", Value: doc.Fecha},
			{Path: "descripcion", Value: doc.Descripcion},
			{Path: "responsable_id", Value: doc.ResponsableID},
			{Path: "visible_participante", Value: doc.VisibleParticipante},
		}); err != nil {
			return err
		}
		if err := r.events.Hito(ctx, tx, false, hitoPayload(id, participanteID, doc)); err != nil {
			return err
		}
		return audit.RecordTx(ctx, r.client, tx, audit.AccionActualizar, audit.RecursoHito, id)
	})
	if status.Code(err) == codes.NotFound {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("error actualizando el hito: %w", err)
	}
	r.events.Notify()
	return nil
}

func (r *hitoRepository) ListByFicha(ctx context.Context, fichaID string) (map[string]models.HitoDocument, error) {
	iter := r.client.Collection(hitosCollection).Where("ficha_id", "==", fichaID).Documents(ctx)
	defer iter.Stop()

	out := make(map[string]models.HitoDocument)
	for {
		snap, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listando hitos: %w", err)
		}
		var doc models.HitoDocument
		if err := snap.DataTo(&doc); err != nil {
			return nil, fmt.Errorf("error leyendo un hito: %w", err)
		}
		out[snap.Ref.ID] = doc
	}
	return out, nil
}
