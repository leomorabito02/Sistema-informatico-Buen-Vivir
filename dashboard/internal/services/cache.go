package services

import (
	"context"
	"sync"
	"time"

	"buenvivir/dashboard/internal/clients"
	"buenvivir/dashboard/internal/models"
)

// Fuente identifica cada una de las consultas del dashboard que se
// cachean (una por indicador de HU-52).
type Fuente string

const (
	FuenteParticipantes Fuente = "participantes"
	FuenteCursos        Fuente = "cursos"
	FuenteHabilidades   Fuente = "habilidades"
	FuentePostulaciones Fuente = "postulaciones"
	FuenteAreas         Fuente = "areas"
	FuentePromocion     Fuente = "promocion"
)

// CachedDashboardService guarda en memoria lo que devuelve cada fuente,
// para no consultar a los demás servicios en cada carga del dashboard.
// Los eventos de Pub/Sub (participante.creado, curso.creado, ...) la
// invalidan apenas cambia algo; el TTL es el respaldo para las
// instancias que no recibieron el push y para las áreas formativas (que
// no tienen evento). Dashboard sigue sin base de datos: la evolución
// temporal se reconstruye con las fechas que guardan los servicios de
// origen (ver docs/adr/0003-dashboard-sin-snapshots.md).
//
// Los datos son los mismos para todos los usuarios que pueden ver el
// dashboard (solo administración), por eso la caché no depende de quién
// consulta. Los errores no se cachean.
type CachedDashboardService struct {
	inner DashboardService
	ttl   time.Duration
	now   func() time.Time

	mu      sync.Mutex
	entries map[Fuente]cacheEntry
}

type cacheEntry struct {
	data      any
	expiresAt time.Time
}

func NewCachedDashboardService(inner DashboardService, ttl time.Duration) *CachedDashboardService {
	return &CachedDashboardService{inner: inner, ttl: ttl, now: time.Now, entries: map[Fuente]cacheEntry{}}
}

// Invalidate descarta las fuentes indicadas (todas si no se indica
// ninguna).
func (c *CachedDashboardService) Invalidate(fuentes ...Fuente) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(fuentes) == 0 {
		c.entries = map[Fuente]cacheEntry{}
		return
	}
	for _, f := range fuentes {
		delete(c.entries, f)
	}
}

func cached[T any](c *CachedDashboardService, f Fuente, ctx context.Context, cred clients.Credentials, load func(context.Context, clients.Credentials) (T, error)) (T, error) {
	if c.ttl > 0 {
		c.mu.Lock()
		e, ok := c.entries[f]
		c.mu.Unlock()
		if ok && c.now().Before(e.expiresAt) {
			return e.data.(T), nil
		}
	}
	data, err := load(ctx, cred)
	if err != nil || c.ttl <= 0 {
		return data, err
	}
	c.mu.Lock()
	c.entries[f] = cacheEntry{data: data, expiresAt: c.now().Add(c.ttl)}
	c.mu.Unlock()
	return data, nil
}

func (c *CachedDashboardService) Participantes(ctx context.Context, cred clients.Credentials) ([]models.Participante, error) {
	return cached(c, FuenteParticipantes, ctx, cred, c.inner.Participantes)
}

func (c *CachedDashboardService) Cursos(ctx context.Context, cred clients.Credentials) ([]models.Curso, error) {
	return cached(c, FuenteCursos, ctx, cred, c.inner.Cursos)
}

func (c *CachedDashboardService) Habilidades(ctx context.Context, cred clients.Credentials) ([]models.ParticipanteHabilidad, error) {
	return cached(c, FuenteHabilidades, ctx, cred, c.inner.Habilidades)
}

func (c *CachedDashboardService) Postulaciones(ctx context.Context, cred clients.Credentials) ([]models.Postulacion, error) {
	return cached(c, FuentePostulaciones, ctx, cred, c.inner.Postulaciones)
}

func (c *CachedDashboardService) Areas(ctx context.Context, cred clients.Credentials) ([]models.Area, error) {
	return cached(c, FuenteAreas, ctx, cred, c.inner.Areas)
}

func (c *CachedDashboardService) Promocion(ctx context.Context, cred clients.Credentials) (models.Promocion, error) {
	return cached(c, FuentePromocion, ctx, cred, c.inner.Promocion)
}
