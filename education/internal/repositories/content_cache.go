package repositories

import (
	"sync"
	"time"
)

// contentCache guarda las respuestas de la CDA de Contentful (HU-17): sirve lo fresco
// sin llamar a Contentful, sirve lo vencido si Contentful falla, y los webhooks o las
// escrituras propias la invalidan conservando el cuerpo como respaldo. Es en memoria,
// por instancia (ver docs/adr/0006-cache-de-contenido.md).
type contentCache struct {
	mu         sync.Mutex
	ttl        time.Duration
	maxEntries int
	entries    map[string]cacheEntry
	now        func() time.Time
	// settleUntil: hasta este momento, lo que se guarda queda vencido
	// desde el principio (sirve de respaldo, pero no se sirve como
	// fresco). Ver settleWindow.
	settleUntil time.Time
}

// settleWindow es cuánto tarda la CDA en reflejar lo publicado: durante ese tiempo
// después de invalidar, nada se guarda como fresco para no cachear la versión vieja.
const settleWindow = 30 * time.Second

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
	now := c.now()
	expiresAt := now.Add(c.ttl)
	if now.Before(c.settleUntil) {
		expiresAt = now
	}
	c.entries[key] = cacheEntry{body: body, expiresAt: expiresAt}
}

// invalidate vence todas las respuestas (conserva los cuerpos como respaldo) y abre
// la ventana de settleWindow. Se invalida todo porque los listados incluyen referencias.
func (c *contentCache) invalidate() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	c.settleUntil = now.Add(settleWindow)
	past := now.Add(-time.Second)
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
