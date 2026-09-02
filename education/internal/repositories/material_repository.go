package repositories

import (
	"context"

	"buenvivir/education/internal/models"
)

const materialContentType = "material"

// MaterialRepository encapsula el acceso HTTP a Contentful para el
// Content Type "material". Igual que trainingArea, es catálogo sin
// moderación editorial: crear y editar publican automáticamente.
type MaterialRepository interface {
	List(ctx context.Context) ([]models.Material, error)
	Create(ctx context.Context, input models.MaterialInput) (*models.Material, error)
	Update(ctx context.Context, id string, input models.MaterialInput) (*models.Material, error)
	Delete(ctx context.Context, id string) error
	// UploadAndCreate sube el binario a Contentful como Asset y crea un
	// material que lo referencia (material.file). Devuelve el material
	// ya con file.url resuelta.
	UploadAndCreate(ctx context.Context, title, materialType, fileName, contentType string, data []byte) (*models.Material, error)
}

type materialRepository struct {
	client *ContentfulClient
}

func NewMaterialRepository(client *ContentfulClient) MaterialRepository {
	return &materialRepository{client: client}
}

func (r *materialRepository) List(ctx context.Context) ([]models.Material, error) {
	collection, err := r.client.ListEntries(ctx, materialContentType, 1)
	if err != nil {
		return nil, err
	}

	idx := models.NewIncludesIndex(collection.Includes.Entry, collection.Includes.Asset)
	materials := make([]models.Material, 0, len(collection.Items))
	for _, item := range collection.Items {
		materials = append(materials, materialFromCDAEntry(item, idx))
	}
	return materials, nil
}

func (r *materialRepository) Create(ctx context.Context, input models.MaterialInput) (*models.Material, error) {
	locale := r.client.Locale()
	fields := map[string]any{
		"title": wrapLocale(locale, input.Title),
	}
	if input.Type != "" {
		fields["type"] = wrapLocale(locale, input.Type)
	}
	if input.FileAssetID != "" {
		fields["file"] = wrapLocale(locale, models.NewAssetLink(input.FileAssetID))
	}
	if input.URL != "" {
		fields["url"] = wrapLocale(locale, input.URL)
	}

	created, err := r.client.CreateEntry(ctx, materialContentType, fields)
	if err != nil {
		return nil, err
	}

	if _, err := r.client.PublishEntry(ctx, created.Sys.ID, created.Sys.Version); err != nil {
		return nil, err
	}

	material := materialFromInput(created.Sys.ID, input)
	return &material, nil
}

func (r *materialRepository) Update(ctx context.Context, id string, input models.MaterialInput) (*models.Material, error) {
	current, err := r.client.GetManagementEntry(ctx, id)
	if err != nil {
		return nil, err
	}

	locale := r.client.Locale()
	fields := map[string]any{
		"title": wrapLocale(locale, input.Title),
		"type":  wrapLocale(locale, input.Type),
	}
	if input.FileAssetID != "" {
		fields["file"] = wrapLocale(locale, models.NewAssetLink(input.FileAssetID))
	}
	if input.URL != "" {
		fields["url"] = wrapLocale(locale, input.URL)
	}

	updated, err := r.client.UpdateEntry(ctx, id, current.Sys.Version, fields)
	if err != nil {
		return nil, err
	}

	if _, err := r.client.PublishEntry(ctx, id, updated.Sys.Version); err != nil {
		return nil, err
	}

	material := materialFromInput(id, input)
	return &material, nil
}

func (r *materialRepository) Delete(ctx context.Context, id string) error {
	return r.client.DeleteEntryCascade(ctx, id)
}

func (r *materialRepository) UploadAndCreate(ctx context.Context, title, materialType, fileName, contentType string, data []byte) (*models.Material, error) {
	assetID, err := r.client.UploadAsset(ctx, title, fileName, contentType, data)
	if err != nil {
		return nil, err
	}

	locale := r.client.Locale()
	fields := map[string]any{
		"title": wrapLocale(locale, title),
		"file":  wrapLocale(locale, models.NewAssetLink(assetID)),
	}
	if materialType != "" {
		fields["type"] = wrapLocale(locale, materialType)
	}

	created, err := r.client.CreateEntry(ctx, materialContentType, fields)
	if err != nil {
		return nil, err
	}
	if _, err := r.client.PublishEntry(ctx, created.Sys.ID, created.Sys.Version); err != nil {
		return nil, err
	}

	m := models.Material{
		ID:    created.Sys.ID,
		Title: title,
		Type:  materialType,
		File:  &models.AssetRef{ID: assetID},
	}
	return &m, nil
}

func materialFromCDAEntry(entry models.ContentfulEntry, idx *models.IncludesIndex) models.Material {
	m := models.Material{
		ID:    entry.Sys.ID,
		Title: rawString(entry.Fields, "title"),
		Type:  rawString(entry.Fields, "type"),
		URL:   rawString(entry.Fields, "url"),
	}
	if assetID, ok := rawLink(entry.Fields, "file"); ok {
		if asset, found := idx.Asset(assetID); found {
			m.File = &asset
		}
	}
	return m
}

func materialFromInput(id string, input models.MaterialInput) models.Material {
	m := models.Material{
		ID:    id,
		Title: input.Title,
		Type:  input.Type,
		URL:   input.URL,
	}
	if input.FileAssetID != "" {
		m.File = &models.AssetRef{ID: input.FileAssetID}
	}
	return m
}
