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
	FuenteEquipo        Fuente = "equipo"
)

// CachedDashboardService guarda en memoria lo que devuelve cada fuente. Los eventos de
// Pub/Sub la invalidan y el TTL es el respaldo. Es la misma para todos (solo admin) y
// los errores no se cachean (ver docs/adr/0003-dashboard-sin-snapshots.md).
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

func (c *CachedDashboardService) Equipo(ctx context.Context, cred clients.Credentials) (models.Equipo, error) {
	return cached(c, FuenteEquipo, ctx, cred, c.inner.Equipo)
}

func (c *CachedDashboardService) Promocion(ctx context.Context, cred clients.Credentials) (models.Promocion, error) {
	return cached(c, FuentePromocion, ctx, cred, c.inner.Promocion)
}
