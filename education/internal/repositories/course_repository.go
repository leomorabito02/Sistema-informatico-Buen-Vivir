package repositories

import (
	"context"
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
	// ListByResponsibleTeacher filtra por fields.responsibleTeacher en
	// la CDA, es decir, solo devuelve cursos YA PUBLICADOS de ese
	// docente. Un curso en borrador no aparece acá aunque el campo
	// responsibleTeacher ya lo tenga asignado.
	ListByResponsibleTeacher(ctx context.Context, teacherID string) ([]models.Course, error)
	// ListDraftsByResponsibleTeacher es el complemento por CMA: devuelve
	// los cursos de ese docente que TODAVÍA NO están publicados
	// (borradores). Las referencias vienen solo con su id (la CMA no
	// resuelve includes), igual que la respuesta de los endpoints de
	// escritura.
	ListDraftsByResponsibleTeacher(ctx context.Context, teacherID string) ([]models.Course, error)
	// ListAllDrafts es la versión de administración de
	// ListDraftsByResponsibleTeacher: TODOS los cursos en borrador
	// (sin filtrar por docente), leídos de la CMA. Referencias solo con
	// su id.
	ListAllDrafts(ctx context.Context) ([]models.Course, error)
	// ExistsInManagement confirma si la entry existe en la CMA (borrador
	// o publicada), sin importar su estado de publicación. Lo usa
	// EnrollmentService para distinguir "curso inexistente" de "curso
	// inactivo/despublicado" cuando Get (que solo ve la CDA, publicados)
	// devuelve not-found.
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

// Get busca el curso por sys.id vía el endpoint de colección de la CDA
// (no GET /entries/{id}): ese endpoint de entry suelta no devuelve
// "includes" en la respuesta real de Contentful (confirmado contra la
// API real), así que trainingArea/material/coverImage quedaban sin
// resolver. El endpoint de colección filtrado sí trae "includes".
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

// ListDraftsByResponsibleTeacher lee de la CMA (ve borradores) y se
// queda solo con las entries sin publicar (sys.publishedVersion
// ausente). Un curso publicado y luego despublicado también cae acá,
// que es lo correcto: no está en el catálogo.
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

// ListAllDrafts devuelve todas las entries de "course" sin publicar
// (sys.publishedVersion ausente), sin filtrar por responsibleTeacher.
// Un curso publicado y luego despublicado también cae acá.
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
	// capacity solo se manda si vino seteado: el campo es opcional y,
	// si todavía no existe en el content type de Contentful, mandarlo
	// siempre rompería el Create/Update de CUALQUIER curso (no solo los
	// que quieren fijar cupo). Ver comentario en models.CourseInput.
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

// courseFromCMAEntry mapea una entry cruda de la Content Management API
// (campos envueltos por locale, referencias sin resolver) al modelo de
// dominio. Se usa para listar borradores del docente: trainingArea,
// material y coverImage quedan solo con su id, igual que la respuesta
// de los endpoints de escritura.
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
