package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"buenvivir/human-promotion/internal/audit"
	"buenvivir/human-promotion/internal/events/producers"
	"buenvivir/human-promotion/internal/models"
)

const fichasCollection = "fichas_proceso_personal"

// FichaRepository encapsula el acceso a la colección de fichas. Trabaja
// siempre con los campos confidenciales ya cifrados: cifrar y descifrar
// es responsabilidad de la capa de servicio.
type FichaRepository interface {
	// Create guarda la ficha y devuelve su id. Devuelve ErrAlreadyExists
	// si el participante ya tiene una.
	Create(ctx context.Context, doc models.FichaDocument) (string, error)
	GetByID(ctx context.Context, id string) (*models.FichaDocument, error)
	// GetByParticipante devuelve la ficha del participante y su id, o
	// ErrNotFound si todavía no tiene.
	GetByParticipante(ctx context.Context, participanteID string) (string, *models.FichaDocument, error)
	UpdateContenido(ctx context.Context, id, situacionActual, avances string, updatedAt time.Time) error
	// FindIDsByParticipantes devuelve participante_id -> id de ficha para
	// los participantes que ya tienen ficha (lo usa la búsqueda, HU-36).
	FindIDsByParticipantes(ctx context.Context, participanteIDs []string) (map[string]string, error)
	// DeactivateByParticipante da de baja (lógica) la ficha de un
	// participante que se dio de baja en auth-service. Devuelve false si
	// no tenía ficha o ya estaba de baja (idempotente).
	DeactivateByParticipante(ctx context.Context, participanteID string) (bool, error)
}

// firestoreInLimit es el máximo de valores que admite un filtro "in".
const firestoreInLimit = 30

type fichaRepository struct {
	client *firestore.Client
	events *producers.PromocionHumana
}

// NewFichaRepository recibe opcionalmente el productor de eventos (sin
// él no se publica nada; la auditoría se escribe siempre).
func NewFichaRepository(client *firestore.Client, events ...*producers.PromocionHumana) FichaRepository {
	return &fichaRepository{client: client, events: firstProducer(events)}
}

func firstProducer(ps []*producers.PromocionHumana) *producers.PromocionHumana {
	if len(ps) > 0 {
		return ps[0]
	}
	return nil
}

// Create corre en una transacción: la consulta por participante_id y el
// alta se hacen atómicamente, así dos altas simultáneas para el mismo
// participante no terminan en dos fichas.
func (r *fichaRepository) Create(ctx context.Context, doc models.FichaDocument) (string, error) {
	col := r.client.Collection(fichasCollection)
	ref := col.NewDoc()

	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		existing, err := tx.Documents(col.Where("participante_id", "==", doc.ParticipanteID).Limit(1)).GetAll()
		if err != nil {
			return err
		}
		if len(existing) > 0 {
			return ErrAlreadyExists
		}
		if err := tx.Create(ref, doc); err != nil {
			return err
		}
		if err := r.events.FichaCreada(ctx, tx, ref.ID, doc.ParticipanteID); err != nil {
			return err
		}
		return audit.RecordTx(ctx, r.client, tx, audit.AccionCrear, audit.RecursoFicha, ref.ID)
	})
	if err == ErrAlreadyExists {
		return "", err
	}
	if err != nil {
		return "", fmt.Errorf("error creando la ficha: %w", err)
	}
	r.events.Notify()
	return ref.ID, nil
}

func (r *fichaRepository) GetByID(ctx context.Context, id string) (*models.FichaDocument, error) {
	snap, err := r.client.Collection(fichasCollection).Doc(id).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("error leyendo la ficha: %w", err)
	}
	var doc models.FichaDocument
	if err := snap.DataTo(&doc); err != nil {
		return nil, fmt.Errorf("error leyendo la ficha: %w", err)
	}
	return &doc, nil
}

func (r *fichaRepository) GetByParticipante(ctx context.Context, participanteID string) (string, *models.FichaDocument, error) {
	iter := r.client.Collection(fichasCollection).Where("participante_id", "==", participanteID).Limit(1).Documents(ctx)
	defer iter.Stop()

	snap, err := iter.Next()
	if err == iterator.Done {
		return "", nil, ErrNotFound
	}
	if err != nil {
		return "", nil, fmt.Errorf("error buscando la ficha del participante: %w", err)
	}
	var doc models.FichaDocument
	if err := snap.DataTo(&doc); err != nil {
		return "", nil, fmt.Errorf("error leyendo la ficha: %w", err)
	}
	return snap.Ref.ID, &doc, nil
}

func (r *fichaRepository) UpdateContenido(ctx context.Context, id, situacionActual, avances string, updatedAt time.Time) error {
	ref := r.client.Collection(fichasCollection).Doc(id)
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		snap, err := tx.Get(ref)
		if err != nil {
			return err
		}
		participanteID, _ := snap.Data()["participante_id"].(string)
		if err := tx.Update(ref, []firestore.Update{
			{Path: "situacion_actual", Value: situacionActual},
			{Path: "avances", Value: avances},
			{Path: "updated_at", Value: updatedAt},
		}); err != nil {
			return err
		}
		if err := r.events.FichaActualizada(ctx, tx, id, participanteID); err != nil {
			return err
		}
		return audit.RecordTx(ctx, r.client, tx, audit.AccionActualizar, audit.RecursoFicha, id)
	})
	if status.Code(err) == codes.NotFound {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("error actualizando la ficha: %w", err)
	}
	r.events.Notify()
	return nil
}

func (r *fichaRepository) FindIDsByParticipantes(ctx context.Context, participanteIDs []string) (map[string]string, error) {
	out := make(map[string]string, len(participanteIDs))
	for start := 0; start < len(participanteIDs); start += firestoreInLimit {
		end := min(start+firestoreInLimit, len(participanteIDs))
		snaps, err := r.client.Collection(fichasCollection).
			Where("participante_id", "in", participanteIDs[start:end]).
			Select("participante_id").
			Documents(ctx).GetAll()
		if err != nil {
			return nil, fmt.Errorf("error buscando fichas de participantes: %w", err)
		}
		for _, snap := range snaps {
			if pid, ok := snap.Data()["participante_id"].(string); ok {
				out[pid] = snap.Ref.ID
			}
		}
	}
	return out, nil
}

func (r *fichaRepository) DeactivateByParticipante(ctx context.Context, participanteID string) (bool, error) {
	col := r.client.Collection(fichasCollection)
	changed := false
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		changed = false
		docs, err := tx.Documents(col.Where("participante_id", "==", participanteID).Limit(1)).GetAll()
		if err != nil {
			return err
		}
		if len(docs) == 0 {
			return nil
		}
		if estado, _ := docs[0].Data()["estado"].(string); estado == models.FichaEstadoBaja {
			return nil
		}
		changed = true
		if err := tx.Update(docs[0].Ref, []firestore.Update{
			{Path: "estado", Value: models.FichaEstadoBaja},
			{Path: "fecha_baja", Value: time.Now().UTC()},
		}); err != nil {
			return err
		}
		return audit.RecordTx(ctx, r.client, tx, audit.AccionBaja, audit.RecursoFicha, docs[0].Ref.ID)
	})
	if err != nil {
		return false, fmt.Errorf("error dando de baja la ficha: %w", err)
	}
	return changed, nil
}
