package services

import (
	"context"
	"encoding/base64"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"buenvivir/campaigns/internal/events/producers"
	"buenvivir/campaigns/internal/models"
	"buenvivir/campaigns/internal/repositories"
)

// Largos máximos, en caracteres.
const (
	maxTitulo    = 150
	maxTexto     = 5000
	maxImagenURL = 2048
	// Foto subida desde la computadora: el navegador la achica y viaja como
	// data:image/...;base64 (límite de Firestore: 1 MB por documento).
	maxImagenDatos = 700000
	maxBusqueda    = 100
	minAnioFecha   = 1900
	maxAnioFecha   = 2100
)

var (
	htmlTagPattern = regexp.MustCompile(`</?[a-zA-Z][^>]*>`)
	idPattern      = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)
)

// CampaignService aplica las reglas de HU-47 a HU-51 antes de delegar en
// el repository.
type CampaignService interface {
	// Create crea la campaña en borrador, a nombre de quien la carga (HU-47).
	Create(ctx context.Context, in models.CampaignInput, creadoPor string) (models.Campaign, error)
	// Update edita una campaña, esté en borrador o publicada (HU-48).
	Update(ctx context.Context, id string, in models.CampaignInput) (models.Campaign, error)
	// Publish la pasa a "publicada" (HU-49). Publicar una ya publicada no
	// cambia nada.
	Publish(ctx context.Context, id string) (models.Campaign, error)
	Get(ctx context.Context, id string) (models.Campaign, error)
	// Delete da de baja (baja lógica) una campaña, esté en borrador o
	// publicada: deja de verse en la gestión y en la sección pública.
	Delete(ctx context.Context, id, usuarioID string) error
	// List devuelve todas las campañas para la gestión, opcionalmente
	// filtradas por estado y por título.
	List(ctx context.Context, busqueda, estado string) ([]models.Campaign, error)
	// ListPublished devuelve solo las publicadas, para la sección pública
	// (HU-50), opcionalmente filtradas por título (HU-51).
	ListPublished(ctx context.Context, busqueda string) ([]models.Campaign, error)
}

type CampaignServiceOption func(*campaignService)

// WithCampanasProducer configures the outbox producer for campaign domain events.
func WithCampanasProducer(p producers.CampanasProducer) CampaignServiceOption {
	return func(s *campaignService) {
		s.producer = p
	}
}

type campaignService struct {
	repo     repositories.CampaignRepository
	producer producers.CampanasProducer
}

func NewCampaignService(repo repositories.CampaignRepository, opts ...CampaignServiceOption) CampaignService {
	s := &campaignService{repo: repo}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *campaignService) Create(ctx context.Context, in models.CampaignInput, creadoPor string) (models.Campaign, error) {
	if strings.TrimSpace(creadoPor) == "" {
		return models.Campaign{}, NewValidationError("No se pudo identificar al usuario que crea la campaña")
	}
	in, err := normalizarInput(in)
	if err != nil {
		return models.Campaign{}, err
	}
	created, err := s.repo.Create(ctx, models.Campaign{
		Titulo:    in.Titulo,
		Texto:     in.Texto,
		ImagenURL: in.ImagenURL,
		Fecha:     in.Fecha,
		Estado:    models.EstadoBorrador,
		CreadoPor: creadoPor,
	})
	if err != nil {
		return models.Campaign{}, err
	}
	// campana.creada no se emite acá: la campaña nace en borrador. El evento (que dispara
	// los avisos) sale una sola vez, al publicarla.
	return created, nil
}

func (s *campaignService) Update(ctx context.Context, id string, in models.CampaignInput) (models.Campaign, error) {
	if !validID(id) {
		return models.Campaign{}, repositories.ErrNotFound
	}
	in, err := normalizarInput(in)
	if err != nil {
		return models.Campaign{}, err
	}
	// Una campaña dada de baja no se edita (Get da ErrNotFound).
	if _, err := s.repo.Get(ctx, id); err != nil {
		return models.Campaign{}, err
	}
	return s.repo.Update(ctx, id, in)
}

func (s *campaignService) Delete(ctx context.Context, id, usuarioID string) error {
	if !validID(id) {
		return repositories.ErrNotFound
	}
	if strings.TrimSpace(usuarioID) == "" {
		return NewValidationError("No se pudo identificar al usuario que elimina la campaña")
	}
	return s.repo.Delete(ctx, id, usuarioID)
}

func (s *campaignService) Publish(ctx context.Context, id string) (models.Campaign, error) {
	c, err := s.Get(ctx, id)
	if err != nil {
		return models.Campaign{}, err
	}
	if c.Estado == models.EstadoPublicada {
		return c, nil
	}
	published, err := s.repo.SetEstado(ctx, id, models.EstadoPublicada)
	if err != nil {
		return models.Campaign{}, err
	}
	if s.producer != nil {
		if err := s.producer.CampanaCreada(ctx, published.ID, published.CreadoPor); err != nil {
			slog.Warn("failed to emit campana.creada on publish to outbox", slog.String("error", err.Error()), slog.String("campaign_id", published.ID))
		} else {
			s.producer.Notify()
		}
	}
	return published, nil
}

func (s *campaignService) Get(ctx context.Context, id string) (models.Campaign, error) {
	if !validID(id) {
		return models.Campaign{}, repositories.ErrNotFound
	}
	return s.repo.Get(ctx, id)
}

func (s *campaignService) List(ctx context.Context, busqueda, estado string) ([]models.Campaign, error) {
	estado = strings.TrimSpace(estado)
	if estado != "" && estado != models.EstadoBorrador && estado != models.EstadoPublicada {
		return nil, NewValidationError(`El estado debe ser "borrador" o "publicada"`)
	}
	return s.listar(ctx, busqueda, estado)
}

func (s *campaignService) ListPublished(ctx context.Context, busqueda string) ([]models.Campaign, error) {
	return s.listar(ctx, busqueda, models.EstadoPublicada)
}

// listar filtra por título en memoria: Firestore no busca por parte de un
// texto, y la cantidad de campañas de un centro barrial es chica.
func (s *campaignService) listar(ctx context.Context, busqueda, estado string) ([]models.Campaign, error) {
	busqueda = sanitizeText(busqueda)
	if tooLong(busqueda, maxBusqueda) {
		return nil, NewValidationError("La búsqueda no puede superar los 100 caracteres")
	}
	campaigns, err := s.repo.List(ctx, estado)
	if err != nil {
		return nil, err
	}
	if busqueda == "" {
		return campaigns, nil
	}
	clave := normalizarBusqueda(busqueda)
	filtradas := make([]models.Campaign, 0, len(campaigns))
	for _, c := range campaigns {
		if strings.Contains(normalizarBusqueda(c.Titulo), clave) {
			filtradas = append(filtradas, c)
		}
	}
	return filtradas, nil
}

// normalizarInput limpia los textos y valida los cuatro datos de HU-47:
// título, texto, imagen y fecha.
func normalizarInput(in models.CampaignInput) (models.CampaignInput, error) {
	in.Titulo = sanitizeText(in.Titulo)
	in.Texto = sanitizeText(in.Texto)
	in.ImagenURL = strings.TrimSpace(in.ImagenURL)
	in.Fecha = strings.TrimSpace(in.Fecha)

	var faltan []string
	for _, campo := range []struct{ nombre, valor string }{
		{"título", in.Titulo}, {"texto", in.Texto}, {"imagen", in.ImagenURL}, {"fecha", in.Fecha},
	} {
		if campo.valor == "" {
			faltan = append(faltan, campo.nombre)
		}
	}
	if len(faltan) > 0 {
		return in, NewValidationError("Completar los campos obligatorios: " + strings.Join(faltan, ", "))
	}
	if tooLong(in.Titulo, maxTitulo) {
		return in, NewValidationError("El título no puede superar los 150 caracteres")
	}
	if tooLong(in.Texto, maxTexto) {
		return in, NewValidationError("El texto no puede superar los 5000 caracteres")
	}
	if !validImageURL(in.ImagenURL) {
		return in, NewValidationError("La imagen debe ser una dirección web válida (http o https)")
	}
	fecha, ok := parseFecha(in.Fecha)
	if !ok {
		return in, NewValidationError("La fecha debe tener el formato AAAA-MM-DD")
	}
	in.Fecha = fecha
	return in, nil
}

// sanitizeText descarta UTF-8 inválido, etiquetas HTML y caracteres de
// control (salvo salto de línea y tabulación), y recorta espacios. Los
// textos se guardan como texto plano: el frontend los muestra con
// interpolación, nunca como HTML (mismo criterio que human-promotion).
func sanitizeText(s string) string {
	s = strings.ToValidUTF8(s, "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = htmlTagPattern.ReplaceAllString(s, "")
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}

func tooLong(s string, max int) bool {
	return utf8.RuneCountInString(s) > max
}

// validImageURL acepta solo direcciones http(s) con host: descarta, por
// ejemplo, "javascript:" o "data:", que el navegador podría ejecutar.
func validImageURL(s string) bool {
	if validImageData(s) {
		return true
	}
	if tooLong(s, maxImagenURL) {
		return false
	}
	u, err := url.Parse(s)
	if err != nil || u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

// validImageData acepta una foto subida (jpeg, png o webp en base64). Solo
// esos tipos: nada de svg ni html, que podrían ejecutar código.
func validImageData(s string) bool {
	if len(s) > maxImagenDatos {
		return false
	}
	for _, p := range []string{"data:image/jpeg;base64,", "data:image/png;base64,", "data:image/webp;base64,"} {
		if strings.HasPrefix(s, p) {
			_, err := base64.StdEncoding.DecodeString(s[len(p):])
			return err == nil && len(s) > len(p)
		}
	}
	return false
}

// parseFecha valida una fecha calendario AAAA-MM-DD en un rango razonable.
func parseFecha(fecha string) (string, bool) {
	t, err := time.Parse(models.FechaLayout, fecha)
	if err != nil || t.Year() < minAnioFecha || t.Year() > maxAnioFecha {
		return "", false
	}
	return t.Format(models.FechaLayout), true
}

// validID acepta los ids que genera Firestore (alfanuméricos) y rechaza
// cualquier otra cosa antes de llegar a la base.
func validID(id string) bool {
	return idPattern.MatchString(id)
}

// normalizarBusqueda pasa a minúsculas y quita tildes, para que
// "campana" encuentre "Campaña" y "ARBOL" encuentre "árbol".
func normalizarBusqueda(s string) string {
	reemplazos := strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")
	return reemplazos.Replace(strings.ToLower(s))
}
