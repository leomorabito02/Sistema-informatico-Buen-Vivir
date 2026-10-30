package repositories

import (
	"context"
	"fmt"
	"sort"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"buenvivir/education/internal/models"
)

const commentCollection = "comentarios"

// maxComentariosPorEntidad acota lo que se lee por curso o campaña.
const maxComentariosPorEntidad = 200

// CommentRepository encapsula la colección Firestore "comentarios"
// (HU-24, DER Comentario).
type CommentRepository interface {
	// ListByEntidad devuelve los comentarios de un curso o campaña, del
	// más nuevo al más viejo.
	ListByEntidad(ctx context.Context, entidadTipo, entidadID string) ([]models.Comentario, error)
	Create(ctx context.Context, c *models.Comentario) error
	// Delete devuelve ErrNotFound si el id no existe.
	Delete(ctx context.Context, id string) error
}

type commentRepository struct {
	client *firestore.Client
}

func NewCommentRepository(client *firestore.Client) CommentRepository {
	return &commentRepository{client: client}
}

func (r *commentRepository) ListByEntidad(ctx context.Context, entidadTipo, entidadID string) ([]models.Comentario, error) {
	// Solo filtros de igualdad: el orden se hace acá para no necesitar
	// un índice compuesto en Firestore.
	iter := r.client.Collection(commentCollection).
		Where("entidad_tipo", "==", entidadTipo).
		Where("entidad_id", "==", entidadID).
		Limit(maxComentariosPorEntidad).
		Documents(ctx)
	defer iter.Stop()

	out := make([]models.Comentario, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listando comentarios: %w", err)
		}
		var c models.Comentario
		if err := doc.DataTo(&c); err != nil {
			return nil, fmt.Errorf("error leyendo comentario: %w", err)
		}
		c.ID = doc.Ref.ID
		out = append(out, c)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Fecha.After(out[j].Fecha) })
	return out, nil
}

func (r *commentRepository) Create(ctx context.Context, c *models.Comentario) error {
	ref := r.client.Collection(commentCollection).NewDoc()
	if _, err := ref.Create(ctx, c); err != nil {
		return fmt.Errorf("error guardando comentario: %w", err)
	}
	c.ID = ref.ID
	return nil
}

func (r *commentRepository) Delete(ctx context.Context, id string) error {
	ref := r.client.Collection(commentCollection).Doc(id)
	if _, err := ref.Get(ctx); err != nil {
		if status.Code(err) == codes.NotFound {
			return ErrNotFound
		}
		return fmt.Errorf("error buscando comentario: %w", err)
	}
	if _, err := ref.Delete(ctx); err != nil {
		return fmt.Errorf("error borrando comentario: %w", err)
	}
	return nil
}
