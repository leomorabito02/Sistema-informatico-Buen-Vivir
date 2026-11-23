package repositories

import (
	"context"
	"fmt"
	"strings"

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
	ListByResponsibleTeachers(ctx context.Context, teacherIDs []string) ([]models.Course, error)
	List(ctx context.Context) ([]models.Course, error)
	Get(ctx context.Context, id string) (*models.Course, error)
	Create(ctx context.Context, input models.CourseInput) (*models.Course, error)
	Update(ctx context.Context, id string, input models.CourseInput) (*models.Course, error)
	Delete(ctx context.Context, id string) error
	Publish(ctx context.Context, id string) error
	Unpublish(ctx context.Context, id string) error
	// ListByResponsibleTeacher devuelve los cursos publicados (CDA) de ese docente.
	ListByResponsibleTeacher(ctx context.Context, teacherID string) ([]models.Course, error)
	// ListDraftsByResponsibleTeacher devuelve los borradores (CMA) de ese docente,
	// incluidos los despublicados. Las referencias vienen solo con su id.
	ListDraftsByResponsibleTeacher(ctx context.Context, teacherID string) ([]models.Course, error)
	// ListAllDrafts devuelve todos los cursos en borrador (CMA), sin filtrar por docente.
	ListAllDrafts(ctx context.Context) ([]models.Course, error)
	// ExistsInManagement confirma si la entry existe en la CMA, publicada o no. Sirve para
	// distinguir "curso inexistente" de "curso despublicado".
	ExistsInManagement(ctx context.Context, id string) (bool, error)
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

// Get usa el endpoint de colección de la CDA filtrado por sys.id porque
// GET /entries/{id} no devuelve "includes".
func (r *courseRepository) Get(ctx context.Context, id string) (*models.Course, error) {
	collection, err := r.client.ListEntriesWhere(ctx, courseContentType, courseInclude, map[string]string{
		"sys.id": id,
	})
	if err != nil {
		return nil, err
	}
	if len(collection.Items) == 0 {
		return nil, ErrNotFound
	}

	idx := models.NewIncludesIndex(collection.Includes.Entry, collection.Includes.Asset)
	course := courseFromCDAEntry(collection.Items[0], idx)
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
func (r *courseRepository) Update(ctx context.Context, id string, input models.CourseInput) (*models.Course, error) {
	current, err := r.client.GetManagementEntry(ctx, id)
	if err != nil {
		return nil, err
	}

	updated, err := r.client.UpdateEntry(ctx, id, current.Sys.Version, courseInputToFields(r.client.Locale(), input))
	if err != nil {
		return nil, err
	}

	// Un curso publicado se vuelve a publicar para que el cambio se vea en el catálogo;
	// un borrador queda como borrador.
	if current.Sys.PublishedVersion != 0 {
		published, err := r.client.PublishEntry(ctx, id, updated.Sys.Version)
		if err != nil {
			return nil, fmt.Errorf("el curso se guardó pero no se pudo volver a publicar: %w", err)
		}
		updated = published
	}

	course := courseFromInput(updated.Sys, input)
	return &course, nil
}

func (r *courseRepository) Delete(ctx context.Context, id string) error {
	return r.client.DeleteEntryCascade(ctx, id)
}

func (r *courseRepository) ExistsInManagement(ctx context.Context, id string) (bool, error) {
	_, err := r.client.GetManagementEntry(ctx, id)
	if err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
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

func (r *courseRepository) ListByResponsibleTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	collection, err := r.client.ListEntriesWhere(ctx, courseContentType, courseInclude, map[string]string{
		"fields.responsibleTeacher": teacherID,
	})
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

func (r *courseRepository) ListByResponsibleTeachers(ctx context.Context, teacherIDs []string) ([]models.Course, error) {
	if len(teacherIDs) == 0 {
		return []models.Course{}, nil
	}
	courses := make([]models.Course, 0)
	for i := 0; i < len(teacherIDs); i += 100 {
		end := i + 100
		if end > len(teacherIDs) {
			end = len(teacherIDs)
		}
		chunk := teacherIDs[i:end]
		joinedIDs := strings.Join(chunk, ",")
		collection, err := r.client.ListEntriesWhere(ctx, courseContentType, courseInclude, map[string]string{
			"fields.responsibleTeacher[in]": joinedIDs,
		})
		if err != nil {
			return nil, err
		}
		idx := models.NewIncludesIndex(collection.Includes.Entry, collection.Includes.Asset)
		for _, item := range collection.Items {
			courses = append(courses, courseFromCDAEntry(item, idx))
		}
	}
	return courses, nil
}

func (r *courseRepository) ListDraftsByResponsibleTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	collection, err := r.client.ListManagementEntriesWhere(ctx, courseContentType, map[string]string{
		"fields.responsibleTeacher":    teacherID,
		"sys.publishedVersion[exists]": "false",
	})
	if err != nil {
		return nil, err
	}

	courses := make([]models.Course, 0, len(collection.Items))
	for _, item := range collection.Items {
		courses = append(courses, courseFromCMAEntry(r.client.Locale(), item))
	}
	return courses, nil
}

func (r *courseRepository) ListAllDrafts(ctx context.Context) ([]models.Course, error) {
	collection, err := r.client.ListManagementEntriesWhere(ctx, courseContentType, map[string]string{
		"sys.publishedVersion[exists]": "false",
	})
	if err != nil {
		return nil, err
	}

	courses := make([]models.Course, 0, len(collection.Items))
	for _, item := range collection.Items {
		courses = append(courses, courseFromCMAEntry(r.client.Locale(), item))
	}
	return courses, nil
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
	// capacity solo se manda si vino: si el campo no existe en el content type, mandarlo
	// rompería el guardado de cualquier curso.
	if input.Capacity != nil {
		fields["capacity"] = wrapLocale(locale, *input.Capacity)
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
		Capacity:           rawInt(entry.Fields, "capacity"),
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

// courseFromCMAEntry mapea una entry de la CMA (campos por locale, referencias sin
// resolver) al modelo. Las referencias quedan solo con su id.
func courseFromCMAEntry(locale string, entry models.ContentfulEntry) models.Course {
	flat := flattenCMAFields(entry.Fields, locale)

	c := models.Course{
		ID:                 entry.Sys.ID,
		Title:              rawString(flat, "title"),
		Description:        flat["description"],
		ResponsibleTeacher: rawString(flat, "responsibleTeacher"),
		Mode:               rawString(flat, "mode"),
		Duration:           rawString(flat, "duration"),
		Capacity:           rawInt(flat, "capacity"),
		CreatedAt:          entry.Sys.CreatedAt,
		UpdatedAt:          entry.Sys.UpdatedAt,
	}

	if assetID, ok := rawLink(flat, "coverImage"); ok {
		c.CoverImage = &models.AssetRef{ID: assetID}
	}
	if trainingAreaID, ok := rawLink(flat, "trainingArea"); ok {
		c.TrainingArea = &models.TrainingArea{ID: trainingAreaID}
	}
	for _, materialID := range rawLinks(flat, "material") {
		c.Material = append(c.Material, models.Material{ID: materialID})
	}

	return c
}

// courseFromInput arma la respuesta de las escrituras con lo enviado más los
// metadatos de Contentful. Las referencias quedan solo con su id.
func courseFromInput(sys models.ContentfulSys, input models.CourseInput) models.Course {
	c := models.Course{
		ID:                 sys.ID,
		Title:              input.Title,
		Description:        input.Description,
		ResponsibleTeacher: input.ResponsibleTeacher,
		Mode:               input.Mode,
		Duration:           input.Duration,
		Capacity:           input.Capacity,
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
