package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"buenvivir/education/internal/config"
	"buenvivir/education/internal/models"
)

const (
	defaultCDABaseURL = "https://cdn.contentful.com"
	defaultCMABaseURL = "https://api.contentful.com"
)

// ContentfulAPIError envuelve un error de Contentful conservando el status HTTP para
// mapearlo a la respuesta (404, 409, etc.).
type ContentfulAPIError struct {
	StatusCode int
	Body       string
}

func (e *ContentfulAPIError) Error() string {
	return fmt.Sprintf("contentful respondió %d: %s", e.StatusCode, e.Body)
}

func IsConflict(err error) bool {
	apiErr, ok := err.(*ContentfulAPIError)
	return ok && apiErr.StatusCode == http.StatusConflict
}

// ContentfulClient es el cliente HTTP de Contentful compartido por los repositories:
// lee de la CDA y escribe en la CMA con versionado optimista.
type ContentfulClient struct {
	spaceID     string
	environment string
	cdaToken    string
	cmaToken    string
	locale      string
	httpClient  *http.Client

	// Los hosts son campos para que los tests los apunten a un httptest.Server.
	cdaBaseURL    string
	cmaBaseURL    string
	uploadBaseURL string

	// cache de las lecturas de la CDA (nil = sin caché). Ver
	// content_cache.go.
	cache *contentCache
}

func NewContentfulClient(cfg *config.Config) *ContentfulClient {
	return &ContentfulClient{
		spaceID:       cfg.ContentfulSpaceID,
		environment:   string(cfg.ContentfulEnvironment),
		cdaToken:      cfg.ContentfulCDAToken,
		cmaToken:      cfg.ContentfulCMAToken,
		locale:        cfg.ContentfulLocale,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
		cdaBaseURL:    defaultCDABaseURL,
		cmaBaseURL:    defaultCMABaseURL,
		uploadBaseURL: defaultUploadBaseURL,
		cache:         newContentCache(cfg.Events.ContentCacheTTL),
	}
}

// InvalidateCache vence las lecturas cacheadas de la CDA (las conserva
// como respaldo ante una caída de Contentful). La llama el consumidor de
// contenido-eventos.
func (c *ContentfulClient) InvalidateCache() {
	c.cache.invalidate()
}

func (c *ContentfulClient) Locale() string {
	return c.locale
}

// -- Content Delivery API (lectura pública, solo contenido publicado) --

// ListEntries devuelve las entries publicadas de un content type,
// junto con sus referencias resueltas (includes) hasta la profundidad
// indicada.
func (c *ContentfulClient) ListEntries(ctx context.Context, contentType string, include int) (*models.ContentfulCollection, error) {
	return c.ListEntriesWhere(ctx, contentType, include, nil)
}

// ListEntriesWhere es como ListEntries con filtros de query de la CDA
// (ej. {"fields.responsibleTeacher": "<userId>"}).
func (c *ContentfulClient) ListEntriesWhere(ctx context.Context, contentType string, include int, filters map[string]string) (*models.ContentfulCollection, error) {
	query := url.Values{}
	query.Set("content_type", contentType)
	query.Set("include", fmt.Sprintf("%d", include))
	query.Set("order", "-sys.createdAt")
	for k, v := range filters {
		query.Set(k, v)
	}

	body, _, err := c.doCDA(ctx, http.MethodGet, "/entries", query)
	if err != nil {
		return nil, err
	}

	var collection models.ContentfulCollection
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, fmt.Errorf("no se pudo parsear la respuesta de Contentful: %w", err)
	}
	return &collection, nil
}

// ListManagementEntriesWhere lista entries por la CMA, que también devuelve borradores.
// La CMA no resuelve includes y cada campo viene envuelto por locale (ver flattenCMAFields).
func (c *ContentfulClient) ListManagementEntriesWhere(ctx context.Context, contentType string, filters map[string]string) (*models.ContentfulCollection, error) {
	query := url.Values{}
	query.Set("content_type", contentType)
	query.Set("order", "-sys.createdAt")
	for k, v := range filters {
		query.Set(k, v)
	}

	body, _, err := c.doCMA(ctx, http.MethodGet, "/entries?"+query.Encode(), nil, nil)
	if err != nil {
		return nil, err
	}

	var collection models.ContentfulCollection
	if err := json.Unmarshal(body, &collection); err != nil {
		return nil, fmt.Errorf("no se pudo parsear la respuesta de Contentful: %w", err)
	}
	return &collection, nil
}

func (c *ContentfulClient) doCDA(ctx context.Context, method, path string, query url.Values) ([]byte, int, error) {
	if query == nil {
		query = url.Values{}
	}
	// La clave de la caché no incluye el token (Encode ordena las claves,
	// así que la misma consulta da siempre la misma clave).
	cacheKey := method + " " + path + "?" + query.Encode()
	cacheable := method == http.MethodGet && c.cache != nil
	if cacheable {
		if body, fresh, ok := c.cache.get(cacheKey); ok && fresh {
			return body, http.StatusOK, nil
		}
	}

	query.Set("access_token", c.cdaToken)
	reqURL := fmt.Sprintf("%s/spaces/%s/environments/%s%s?%s", c.cdaBaseURL, c.spaceID, c.environment, path, query.Encode())
	req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
	if err != nil {
		return nil, 0, err
	}

	body, status, err := c.do(req)
	if !cacheable {
		return body, status, err
	}
	if err == nil {
		c.cache.put(cacheKey, body)
		return body, status, nil
	}
	if stale, _, ok := c.cache.get(cacheKey); ok && isTransient(status) {
		slog.WarnContext(ctx, "Contentful no disponible: se sirve contenido cacheado (stale)",
			slog.String("path", path), slog.Int("status", status), slog.String("error", err.Error()))
		return stale, http.StatusOK, nil
	}
	return body, status, err
}

// -- Content Management API (escritura) --

// GetManagementEntry trae una entry (publicada o no) con su sys.version
// actual, necesario antes de editar, publicar o borrar (versionado
// optimista de Contentful).
func (c *ContentfulClient) GetManagementEntry(ctx context.Context, id string) (*models.ContentfulEntry, error) {
	body, _, err := c.doCMA(ctx, http.MethodGet, "/entries/"+id, nil, nil)
	if err != nil {
		return nil, err
	}

	var entry models.ContentfulEntry
	if err := json.Unmarshal(body, &entry); err != nil {
		return nil, fmt.Errorf("no se pudo parsear la respuesta de Contentful: %w", err)
	}
	return &entry, nil
}

// CreateEntry crea una entry en estado borrador (sin publicar).
func (c *ContentfulClient) CreateEntry(ctx context.Context, contentType string, fields map[string]any) (*models.ContentfulEntry, error) {
	payload, err := json.Marshal(map[string]any{"fields": fields})
	if err != nil {
		return nil, err
	}

	headers := map[string]string{"X-Contentful-Content-Type": contentType}
	body, _, err := c.doCMA(ctx, http.MethodPost, "/entries", headers, payload)
	if err != nil {
		return nil, err
	}

	var entry models.ContentfulEntry
	if err := json.Unmarshal(body, &entry); err != nil {
		return nil, fmt.Errorf("no se pudo parsear la respuesta de Contentful: %w", err)
	}
	return &entry, nil
}

// UpdateEntry reemplaza por completo los fields de una entry existente
// (semántica PUT). version debe ser el sys.version actual, obtenido
// con GetManagementEntry.
func (c *ContentfulClient) UpdateEntry(ctx context.Context, id string, version int, fields map[string]any) (*models.ContentfulEntry, error) {
	payload, err := json.Marshal(map[string]any{"fields": fields})
	if err != nil {
		return nil, err
	}

	headers := map[string]string{"X-Contentful-Version": fmt.Sprintf("%d", version)}
	body, _, err := c.doCMA(ctx, http.MethodPut, "/entries/"+id, headers, payload)
	if err != nil {
		return nil, err
	}

	var entry models.ContentfulEntry
	if err := json.Unmarshal(body, &entry); err != nil {
		return nil, fmt.Errorf("no se pudo parsear la respuesta de Contentful: %w", err)
	}
	return &entry, nil
}

// PublishEntry publica la versión actual de una entry.
func (c *ContentfulClient) PublishEntry(ctx context.Context, id string, version int) (*models.ContentfulEntry, error) {
	headers := map[string]string{"X-Contentful-Version": fmt.Sprintf("%d", version)}
	body, _, err := c.doCMA(ctx, http.MethodPut, "/entries/"+id+"/published", headers, nil)
	if err != nil {
		return nil, err
	}

	var entry models.ContentfulEntry
	if err := json.Unmarshal(body, &entry); err != nil {
		return nil, fmt.Errorf("no se pudo parsear la respuesta de Contentful: %w", err)
	}
	return &entry, nil
}

// UnpublishEntry despublica una entry. Si ya estaba despublicada,
// Contentful responde 404 y lo tratamos como éxito (idempotente).
func (c *ContentfulClient) UnpublishEntry(ctx context.Context, id string) error {
	_, status, err := c.doCMA(ctx, http.MethodDelete, "/entries/"+id+"/published", nil, nil)
	if err != nil && status != http.StatusNotFound {
		return err
	}
	return nil
}

// DeleteEntry borra definitivamente una entry. Contentful exige que la
// entry esté despublicada antes de poder borrarla, por eso los
// repositories llaman primero a UnpublishEntry.
func (c *ContentfulClient) DeleteEntry(ctx context.Context, id string) error {
	_, _, err := c.doCMA(ctx, http.MethodDelete, "/entries/"+id, nil, nil)
	return err
}

// DeleteEntryCascade despublica y luego borra una entry. Si es un borrador, la CMA
// responde 400 al despublicar: se confirma que no tiene versión publicada y se borra igual.
func (c *ContentfulClient) DeleteEntryCascade(ctx context.Context, id string) error {
	if err := c.UnpublishEntry(ctx, id); err != nil {
		apiErr, ok := err.(*ContentfulAPIError)
		if !ok || apiErr.StatusCode != http.StatusBadRequest {
			return err
		}
		entry, getErr := c.GetManagementEntry(ctx, id)
		if getErr != nil {
			return getErr
		}
		if entry.Sys.PublishedVersion != 0 {
			return err
		}
	}
	return c.DeleteEntry(ctx, id)
}

func (c *ContentfulClient) doCMA(ctx context.Context, method, path string, extraHeaders map[string]string, body []byte) ([]byte, int, error) {
	reqURL := fmt.Sprintf("%s/spaces/%s/environments/%s%s", c.cmaBaseURL, c.spaceID, c.environment, path)

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reader)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Authorization", "Bearer "+c.cmaToken)
	req.Header.Set("Content-Type", "application/vnd.contentful.management.v1+json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	respBody, status, err := c.do(req)
	if err == nil && method != http.MethodGet {
		// Una escritura propia cambia lo que devuelve la CDA: se vence la
		// caché para que este mismo usuario vea su cambio enseguida (el
		// webhook de Contentful avisa a las demás instancias).
		c.cache.invalidate()
	}
	return respBody, status, err
}

func (c *ContentfulClient) do(req *http.Request) ([]byte, int, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("error de red hablando con Contentful: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("no se pudo leer la respuesta de Contentful: %w", err)
	}

	if resp.StatusCode >= 400 {
		return respBody, resp.StatusCode, &ContentfulAPIError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	return respBody, resp.StatusCode, nil
}

// wrapLocale envuelve un valor de campo bajo el locale configurado,
// tal como lo espera el body de la Content Management API
// (fields.<campo>.<locale> = valor).
func wrapLocale(locale string, value any) map[string]any {
	return map[string]any{locale: value}
}
