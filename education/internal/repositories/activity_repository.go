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
	ListByDateRange(ctx context.Context, start, end time.Time) ([]models.Activity, error)
	Get(ctx context.Context, id string) (*models.Activity, error)
	Create(ctx context.Context, cursoID string, input models.ActivityInput) (*models.Activity, error)
	// Update reemplaza titulo/consigna/fecha de la actividad y, en la
	// misma transacción, reescribe la fecha de sus asistencias: la
	// asistencia generada por una entrega lleva siempre la fecha de la
	// actividad, así que si la actividad se reprograma, sus asistencias
	// se mueven con ella.
	Update(ctx context.Context, id string, input models.ActivityInput) (*models.Activity, error)
	// Delete borra la actividad junto con sus entregas, asistencias y
	// habilidades asociadas en una sola transacción, para que el progreso (entregadas / total) no
	// quede contando registros de una actividad que ya no existe. Los
	// adjuntos del bucket los limpia el service (ver
	// ActivityService.DeleteActivity).
	Delete(ctx context.Context, id string) error
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

func (r *activityRepository) ListByDateRange(ctx context.Context, start, end time.Time) ([]models.Activity, error) {
	iter := r.client.Collection(activitiesCollection).
		Where("fecha", ">=", start).
		Where("fecha", "<=", end).
		Documents(ctx)
	defer iter.Stop()

	activities := make([]models.Activity, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error listando actividades por rango de fecha: %w", err)
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
		Consigna:  input.Consigna,
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

func (r *activityRepository) Update(ctx context.Context, id string, input models.ActivityInput) (*models.Activity, error) {
	ref := r.client.Collection(activitiesCollection).Doc(id)
	var updated models.Activity

	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return ErrNotFound
			}
			return err
		}
		if err := doc.DataTo(&updated); err != nil {
			return err
		}
		attendances, err := tx.Documents(r.client.Collection(attendancesCollection).Where("actividad_id", "==", id)).GetAll()
		if err != nil {
			return err
		}

		updated.Titulo = input.Titulo
		updated.Consigna = input.Consigna
		updated.Fecha = input.Fecha
		if err := tx.Set(ref, updated); err != nil {
			return err
		}
		for _, a := range attendances {
			if err := tx.Update(a.Ref, []firestore.Update{{Path: "fecha", Value: input.Fecha}}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if err == ErrNotFound {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error actualizando actividad: %w", err)
	}
	updated.ID = id
	return &updated, nil
}

func (r *activityRepository) Delete(ctx context.Context, id string) error {
	ref := r.client.Collection(activitiesCollection).Doc(id)

	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		if _, err := tx.Get(ref); err != nil {
			if status.Code(err) == codes.NotFound {
				return ErrNotFound
			}
			return err
		}
		// También se borran sus vínculos de Actividad_Habilidad: si no,
		// el curso seguiría mostrando habilidades de una actividad que ya
		// no existe. Las habilidades que los participantes ya ganaron
		// (participante_habilidades) se conservan: las aprendieron igual.
		var related []*firestore.DocumentSnapshot
		for _, col := range []string{submissionsCollection, attendancesCollection, activitySkillsCollection} {
			docs, err := tx.Documents(r.client.Collection(col).Where("actividad_id", "==", id)).GetAll()
			if err != nil {
				return err
			}
			related = append(related, docs...)
		}

		for _, d := range related {
			if err := tx.Delete(d.Ref); err != nil {
				return err
			}
		}
		return tx.Delete(ref)
	})
	if err != nil {
		if err == ErrNotFound {
			return ErrNotFound
		}
		return fmt.Errorf("error eliminando actividad: %w", err)
	}
	return nil
}
