package repositories

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
}

func NewHitoRepository(client *firestore.Client) HitoRepository {
	return &hitoRepository{client: client}
}

func (r *hitoRepository) Create(ctx context.Context, doc models.HitoDocument) (string, error) {
	ref := r.client.Collection(hitosCollection).NewDoc()
	if _, err := ref.Create(ctx, doc); err != nil {
		return "", fmt.Errorf("error creando el hito: %w", err)
	}
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
	_, err := r.client.Collection(hitosCollection).Doc(id).Update(ctx, []firestore.Update{
		{Path: "fecha", Value: doc.Fecha},
		{Path: "descripcion", Value: doc.Descripcion},
		{Path: "responsable_id", Value: doc.ResponsableID},
		{Path: "visible_participante", Value: doc.VisibleParticipante},
	})
	if status.Code(err) == codes.NotFound {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("error actualizando el hito: %w", err)
	}
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
