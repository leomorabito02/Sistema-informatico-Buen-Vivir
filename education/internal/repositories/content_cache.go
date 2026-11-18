package repositories

import (
	"sync"
	"time"
)

// contentCache guarda las respuestas de la Content Delivery API de
// Contentful (HU-17, riesgo R2):
//
//   - Mientras una respuesta está fresca (TTL) se sirve sin llamar a
//     Contentful.
//   - Si Contentful falla con un 5xx o no responde (timeout, red), se
//     sirve la última respuesta válida aunque esté vencida (stale
//     content), en vez de devolver un error.
//   - Los eventos contenido.publicado / contenido.despublicado (webhook
//     de Contentful → Pub/Sub) y cada escritura propia por la CMA la
//     invalidan: la marcan como vencida pero CONSERVAN el cuerpo, que
//     sigue sirviendo de respaldo si Contentful se cae justo después.
//
// Es por instancia (en memoria): con varias instancias de Cloud Run, el
// push de Pub/Sub invalida solo la que lo recibe; las demás se ponen al
// día al vencer el TTL. Ver docs/adr/0006-cache-de-contenido.md.
type contentCache struct {
	mu         sync.Mutex
	ttl        time.Duration
	maxEntries int
	entries    map[string]cacheEntry
	now        func() time.Time
}

type cacheEntry struct {
	body      []byte
	expiresAt time.Time
}

func newContentCache(ttl time.Duration) *contentCache {
	if ttl <= 0 {
		return nil
	}
	return &contentCache{ttl: ttl, maxEntries: 500, entries: map[string]cacheEntry{}, now: time.Now}
}

// get devuelve la respuesta guardada para key y si todavía está fresca.
func (c *contentCache) get(key string) (body []byte, fresh, ok bool) {
	if c == nil {
		return nil, false, false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return nil, false, false
	}
	return e.body, c.now().Before(e.expiresAt), true
}

func (c *contentCache) put(key string, body []byte) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.entries[key]; !exists && len(c.entries) >= c.maxEntries {
		// Lleno: se descarta la que vence primero.
		var oldest string
		var oldestAt time.Time
		for k, e := range c.entries {
			if oldest == "" || e.expiresAt.Before(oldestAt) {
				oldest, oldestAt = k, e.expiresAt
			}
		}
		delete(c.entries, oldest)
	}
	c.entries[key] = cacheEntry{body: body, expiresAt: c.now().Add(c.ttl)}
}

// invalidate vence todas las respuestas pero conserva sus cuerpos como
// respaldo (ver comentario del tipo). Se invalida todo y no una entry
// puntual porque los listados incluyen las entries y sus referencias.
func (c *contentCache) invalidate() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	past := c.now().Add(-time.Second)
	for k, e := range c.entries {
		e.expiresAt = past
		c.entries[k] = e
	}
}

// isTransient dice si un error de Contentful justifica servir contenido
// vencido: sin respuesta (status 0: timeout/red) o un 5xx.
func isTransient(status int) bool {
	return status == 0 || status >= 500
}
