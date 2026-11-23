package repositories

import (
	"context"

	"buenvivir/education/internal/models"
)

const trainingAreaContentType = "trainingArea"

// TrainingAreaRepository encapsula el Content Type "trainingArea" de Contentful. Es
// un catálogo: crear y editar publican automáticamente.
type TrainingAreaRepository interface {
	List(ctx context.Context) ([]models.TrainingArea, error)
	Create(ctx context.Context, input models.TrainingAreaInput) (*models.TrainingArea, error)
	Update(ctx context.Context, id string, input models.TrainingAreaInput) (*models.TrainingArea, error)
	Delete(ctx context.Context, id string) error
}

type trainingAreaRepository struct {
	client *ContentfulClient
}

func NewTrainingAreaRepository(client *ContentfulClient) TrainingAreaRepository {
	return &trainingAreaRepository{client: client}
}

func (r *trainingAreaRepository) List(ctx context.Context) ([]models.TrainingArea, error) {
	collection, err := r.client.ListEntries(ctx, trainingAreaContentType, 0)
	if err != nil {
		return nil, err
	}

	areas := make([]models.TrainingArea, 0, len(collection.Items))
	for _, item := range collection.Items {
		areas = append(areas, trainingAreaFromCDAEntry(item))
	}
	return areas, nil
}

func (r *trainingAreaRepository) Create(ctx context.Context, input models.TrainingAreaInput) (*models.TrainingArea, error) {
	locale := r.client.Locale()
	fields := map[string]any{
		"name": wrapLocale(locale, input.Name),
	}
	if input.Description != "" {
		fields["description"] = wrapLocale(locale, input.Description)
	}

	created, err := r.client.CreateEntry(ctx, trainingAreaContentType, fields)
	if err != nil {
		return nil, err
	}

	if _, err := r.client.PublishEntry(ctx, created.Sys.ID, created.Sys.Version); err != nil {
		return nil, err
	}

	area := trainingAreaFromInput(created.Sys.ID, input)
	return &area, nil
}

func (r *trainingAreaRepository) Update(ctx context.Context, id string, input models.TrainingAreaInput) (*models.TrainingArea, error) {
	current, err := r.client.GetManagementEntry(ctx, id)
	if err != nil {
		return nil, err
	}

	locale := r.client.Locale()
	fields := map[string]any{
		"name":        wrapLocale(locale, input.Name),
		"description": wrapLocale(locale, input.Description),
	}

	updated, err := r.client.UpdateEntry(ctx, id, current.Sys.Version, fields)
	if err != nil {
		return nil, err
	}

	if _, err := r.client.PublishEntry(ctx, id, updated.Sys.Version); err != nil {
		return nil, err
	}

	area := trainingAreaFromInput(id, input)
	return &area, nil
}

func (r *trainingAreaRepository) Delete(ctx context.Context, id string) error {
	return r.client.DeleteEntryCascade(ctx, id)
}

func trainingAreaFromCDAEntry(entry models.ContentfulEntry) models.TrainingArea {
	return models.TrainingArea{
		ID:          entry.Sys.ID,
		Name:        rawString(entry.Fields, "name"),
		Description: rawString(entry.Fields, "description"),
	}
}

func trainingAreaFromInput(id string, input models.TrainingAreaInput) models.TrainingArea {
	return models.TrainingArea{
		ID:          id,
		Name:        input.Name,
		Description: input.Description,
	}
}
