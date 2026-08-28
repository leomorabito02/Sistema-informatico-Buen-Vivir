package repositories

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"buenvivir/education/internal/models"
)

const activitiesCollection = "actividades"

// ActivityRepository encapsula el acceso a la colección Firestore
// "actividades".
type ActivityRepository interface {
	ListByCourse(ctx context.Context, cursoID string) ([]models.Activity, error)
	Get(ctx context.Context, id string) (*models.Activity, error)
	Create(ctx context.Context, cursoID string, input models.ActivityInput) (*models.Activity, error)
}

type activityRepository struct {
	client *firestore.Client
}

func NewActivityRepository(client *firestore.Client) ActivityRepository {
	return &activityRepository{client: client}
}

func (r *activityRepository) ListByCourse(ctx context.Context, cursoID string) ([]models.Activity, error) {
	iter := r.client.Collection(activitiesCollection).Where("curso_id", "==", cursoID).Documents(ctx)
	defer iter.Stop()

	activities := make([]models.Activity, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listando actividades: %w", err)
		}
		var a models.Activity
		if err := doc.DataTo(&a); err != nil {
			return nil, fmt.Errorf("error leyendo actividad: %w", err)
		}
		a.ID = doc.Ref.ID
		activities = append(activities, a)
	}
	return activities, nil
}

func (r *activityRepository) Get(ctx context.Context, id string) (*models.Activity, error) {
	doc, err := r.client.Collection(activitiesCollection).Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error obteniendo actividad: %w", err)
	}
	var a models.Activity
	if err := doc.DataTo(&a); err != nil {
		return nil, fmt.Errorf("error leyendo actividad: %w", err)
	}
	a.ID = doc.Ref.ID
	return &a, nil
}

func (r *activityRepository) Create(ctx context.Context, cursoID string, input models.ActivityInput) (*models.Activity, error) {
	a := models.Activity{
		CursoID:   cursoID,
		Titulo:    input.Titulo,
		Fecha:     input.Fecha,
		CreatedAt: time.Now().UTC(),
	}

	ref, _, err := r.client.Collection(activitiesCollection).Add(ctx, a)
	if err != nil {
		return nil, fmt.Errorf("error creando actividad: %w", err)
	}
	a.ID = ref.ID
	return &a, nil
}
