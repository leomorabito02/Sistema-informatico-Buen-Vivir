package repositories

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"buenvivir/education/internal/audit"
	"buenvivir/education/internal/models"
)

const cursosBajaCollection = "cursos_baja"

// ErrCursoYaDadoDeBaja indica que el curso ya tenía su baja registrada.
var ErrCursoYaDadoDeBaja = errors.New("el curso ya fue dado de baja")

// CursoBajaRepository guarda la baja lógica de los cursos (HU-16) en la
// colección "cursos_baja". El curso en sí sigue en Contentful: acá solo
// queda la marca de que ya no se ofrece.
type CursoBajaRepository interface {
	// DarDeBaja guarda la baja y su registro de auditoría en una misma
	// transacción. ErrCursoYaDadoDeBaja si ya estaba.
	DarDeBaja(ctx context.Context, baja models.CursoBaja) error
	// IDs devuelve el conjunto de ids de cursos dados de baja.
	IDs(ctx context.Context) (map[string]bool, error)
	// EstaDadoDeBaja dice si un curso tiene la baja registrada.
	EstaDadoDeBaja(ctx context.Context, cursoID string) (bool, error)
}

type cursoBajaRepository struct {
	client *firestore.Client
}

func NewCursoBajaRepository(client *firestore.Client) CursoBajaRepository {
	return &cursoBajaRepository{client: client}
}

func (r *cursoBajaRepository) DarDeBaja(ctx context.Context, baja models.CursoBaja) error {
	ref := r.client.Collection(cursosBajaCollection).Doc(baja.CursoID)
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		if err := tx.Create(ref, baja); err != nil {
			return err
		}
		return audit.RecordTx(ctx, r.client, tx, audit.AccionBaja, audit.RecursoCurso, baja.CursoID)
	})
	if status.Code(err) == codes.AlreadyExists {
		return ErrCursoYaDadoDeBaja
	}
	if err != nil {
		return fmt.Errorf("error registrando la baja del curso: %w", err)
	}
	return nil
}

func (r *cursoBajaRepository) IDs(ctx context.Context) (map[string]bool, error) {
	refs, err := r.client.Collection(cursosBajaCollection).DocumentRefs(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("error leyendo los cursos dados de baja: %w", err)
	}
	out := make(map[string]bool, len(refs))
	for _, ref := range refs {
		out[ref.ID] = true
	}
	return out, nil
}

func (r *cursoBajaRepository) EstaDadoDeBaja(ctx context.Context, cursoID string) (bool, error) {
	_, err := r.client.Collection(cursosBajaCollection).Doc(cursoID).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error leyendo la baja del curso: %w", err)
	}
	return true, nil
}
