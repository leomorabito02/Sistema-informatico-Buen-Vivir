package repositories

import (
	"context"

	"buenvivir/education/internal/models"
)

const courseContentType = "course"

// courseInclude es la profundidad de "includes" que pedimos a la CDA:
// 1 nivel resuelve trainingArea/coverImage/material (entries/assets
// linkeados directo desde course), 2 niveles además resuelve el
// "file" (Media) de cada material.
const courseInclude = 2

// CourseRepository encapsula el acceso HTTP a Contentful para el
// Content Type "course". A diferencia de trainingArea y material, no
// publica automáticamente: crear/editar dejan la entry en borrador y
// requieren Publish explícito (moderación editorial del contenido).
type CourseRepository interface {
	List(ctx context.Context) ([]models.Course, error)
	Get(ctx context.Context, id string) (*models.Course, error)
	Create(ctx context.Context, input models.CourseInput) (*models.Course, error)
	Update(ctx context.Context, id string, input models.CourseInput) (*models.Course, error)
	Delete(ctx context.Context, id string) error
	Publish(ctx context.Context, id string) error
	Unpublish(ctx context.Context, id string) error
}

type courseRepository struct {
	client *ContentfulClient
}

func NewCourseRepository(client *ContentfulClient) CourseRepository {
	return &courseRepository{client: client}
}

func (r *courseRepository) List(ctx context.Context) ([]models.Course, error) {
	collection, err := r.client.ListEntries(ctx, courseContentType, courseInclude)
	if err != nil {
		return nil, err
	}

	idx := models.NewIncludesIndex(collection.Includes.Entry, collection.Includes.Asset)
	courses := make([]models.Course, 0, len(collection.Items))
	for _, item := range collection.Items {
		courses = append(courses, courseFromCDAEntry(item, idx))
	}
	return courses, nil
}

func (r *courseRepository) Get(ctx context.Context, id string) (*models.Course, error) {
	entry, err := r.client.GetPublishedEntry(ctx, id, courseInclude)
	if err != nil {
		return nil, err
	}

	idx := models.NewIncludesIndex(entry.Includes.Entry, entry.Includes.Asset)
	course := courseFromCDAEntry(entry.ContentfulEntry, idx)
	return &course, nil
}

func (r *courseRepository) Create(ctx context.Context, input models.CourseInput) (*models.Course, error) {
	created, err := r.client.CreateEntry(ctx, courseContentType, courseInputToFields(r.client.Locale(), input))
	if err != nil {
		return nil, err
	}

	course := courseFromInput(created.Sys, input)
	return &course, nil
}

// Update reemplaza por completo los campos del curso (semántica PUT).
// No vuelve a publicar automáticamente: si el curso ya estaba
// publicado, la versión pública queda como estaba hasta que se llame
// explícitamente a Publish, que es el comportamiento esperado para
// contenido editorial moderado.
func (r *courseRepository) Update(ctx context.Context, id string, input models.CourseInput) (*models.Course, error) {
	current, err := r.client.GetManagementEntry(ctx, id)
	if err != nil {
		return nil, err
	}

	updated, err := r.client.UpdateEntry(ctx, id, current.Sys.Version, courseInputToFields(r.client.Locale(), input))
	if err != nil {
		return nil, err
	}

	course := courseFromInput(updated.Sys, input)
	return &course, nil
}

func (r *courseRepository) Delete(ctx context.Context, id string) error {
	return r.client.DeleteEntryCascade(ctx, id)
}

func (r *courseRepository) Publish(ctx context.Context, id string) error {
	current, err := r.client.GetManagementEntry(ctx, id)
	if err != nil {
		return err
	}
	_, err = r.client.PublishEntry(ctx, id, current.Sys.Version)
	return err
}

func (r *courseRepository) Unpublish(ctx context.Context, id string) error {
	return r.client.UnpublishEntry(ctx, id)
}

func courseInputToFields(locale string, input models.CourseInput) map[string]any {
	links := make([]models.ContentfulLink, len(input.MaterialIDs))
	for i, materialID := range input.MaterialIDs {
		links[i] = models.NewEntryLink(materialID)
	}

	fields := map[string]any{
		"title":              wrapLocale(locale, input.Title),
		"description":        wrapLocale(locale, input.Description),
		"responsibleTeacher": wrapLocale(locale, input.ResponsibleTeacher),
		"mode":               wrapLocale(locale, input.Mode),
		"duration":           wrapLocale(locale, input.Duration),
		"trainingArea":       wrapLocale(locale, models.NewEntryLink(input.TrainingAreaID)),
		"material":           wrapLocale(locale, links),
	}
	if input.CoverImageAssetID != "" {
		fields["coverImage"] = wrapLocale(locale, models.NewAssetLink(input.CoverImageAssetID))
	}
	return fields
}

// courseFromCDAEntry mapea una entry publicada (Content Delivery API,
// campos ya "planos") al modelo de dominio, resolviendo sus
// referencias contra el índice de "includes".
func courseFromCDAEntry(entry models.ContentfulEntry, idx *models.IncludesIndex) models.Course {
	c := models.Course{
		ID:                 entry.Sys.ID,
		Title:              rawString(entry.Fields, "title"),
		Description:        entry.Fields["description"],
		ResponsibleTeacher: rawString(entry.Fields, "responsibleTeacher"),
		Mode:               rawString(entry.Fields, "mode"),
		Duration:           rawString(entry.Fields, "duration"),
		CreatedAt:          entry.Sys.CreatedAt,
		UpdatedAt:          entry.Sys.UpdatedAt,
	}

	if assetID, ok := rawLink(entry.Fields, "coverImage"); ok {
		if asset, found := idx.Asset(assetID); found {
			c.CoverImage = &asset
		}
	}

	if trainingAreaID, ok := rawLink(entry.Fields, "trainingArea"); ok {
		if taEntry, found := idx.Entry(trainingAreaID); found {
			ta := trainingAreaFromCDAEntry(taEntry)
			c.TrainingArea = &ta
		}
	}

	for _, materialID := range rawLinks(entry.Fields, "material") {
		if mEntry, found := idx.Entry(materialID); found {
			c.Material = append(c.Material, materialFromCDAEntry(mEntry, idx))
		}
	}

	return c
}

// courseFromInput arma la respuesta de los endpoints de escritura a
// partir de lo que el propio cliente envió más los metadatos que
// devuelve Contentful (id, timestamps). No vuelve a resolver
// trainingArea/material/coverImage contra Contentful: esos quedan solo
// con su id. La vista totalmente resuelta se obtiene con un GET
// posterior (que lee de la CDA una vez que el curso esté publicado).
func courseFromInput(sys models.ContentfulSys, input models.CourseInput) models.Course {
	c := models.Course{
		ID:                 sys.ID,
		Title:              input.Title,
		Description:        input.Description,
		ResponsibleTeacher: input.ResponsibleTeacher,
		Mode:               input.Mode,
		Duration:           input.Duration,
		CreatedAt:          sys.CreatedAt,
		UpdatedAt:          sys.UpdatedAt,
	}

	if input.CoverImageAssetID != "" {
		c.CoverImage = &models.AssetRef{ID: input.CoverImageAssetID}
	}
	if input.TrainingAreaID != "" {
		c.TrainingArea = &models.TrainingArea{ID: input.TrainingAreaID}
	}
	for _, materialID := range input.MaterialIDs {
		c.Material = append(c.Material, models.Material{ID: materialID})
	}

	return c
}
