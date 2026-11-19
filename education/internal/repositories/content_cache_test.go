package repositories

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentCache_FrescaVencidaEInvalidada(t *testing.T) {
	now := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	c := newContentCache(time.Minute)
	c.now = func() time.Time { return now }

	_, _, ok := c.get("k")
	assert.False(t, ok)

	c.put("k", []byte("v1"))
	body, fresh, ok := c.get("k")
	assert.True(t, ok)
	assert.True(t, fresh)
	assert.Equal(t, "v1", string(body))

	now = now.Add(2 * time.Minute)
	body, fresh, ok = c.get("k")
	assert.True(t, ok)
	assert.False(t, fresh, "vencida pero conservada como respaldo")
	assert.Equal(t, "v1", string(body))

	c.put("k", []byte("v2"))
	c.invalidate()
	body, fresh, ok = c.get("k")
	assert.True(t, ok)
	assert.False(t, fresh, "invalidar vence sin borrar")
	assert.Equal(t, "v2", string(body))
}

func TestContentCache_NoGuardaComoFrescoJustoDespuesDeInvalidar(t *testing.T) {
	now := time.Date(2026, 11, 18, 12, 0, 0, 0, time.UTC)
	c := newContentCache(5 * time.Minute)
	c.now = func() time.Time { return now }

	c.invalidate()
	// La CDA todavía puede devolver la versión vieja: se guarda como
	// respaldo, pero el próximo pedido vuelve a consultar.
	c.put("k", []byte("vieja"))
	body, fresh, ok := c.get("k")
	assert.True(t, ok)
	assert.False(t, fresh)
	assert.Equal(t, "vieja", string(body))

	now = now.Add(settleWindow + time.Second)
	c.put("k", []byte("nueva"))
	body, fresh, _ = c.get("k")
	assert.True(t, fresh, "pasada la ventana se vuelve a cachear normal")
	assert.Equal(t, "nueva", string(body))
}

func TestContentCache_DesactivadaYLlena(t *testing.T) {
	off := newContentCache(0)
	assert.Nil(t, off)
	off.put("k", []byte("v"))
	off.invalidate()
	_, _, ok := off.get("k")
	assert.False(t, ok)

	c := newContentCache(time.Minute)
	c.maxEntries = 2
	c.put("a", []byte("1"))
	time.Sleep(time.Millisecond)
	c.put("b", []byte("2"))
	c.put("c", []byte("3"))
	_, _, okA := c.get("a")
	_, _, okC := c.get("c")
	assert.False(t, okA, "se descarta la que vence primero")
	assert.True(t, okC)
	assert.Len(t, c.entries, 2)
}

func TestIsTransient(t *testing.T) {
	assert.True(t, isTransient(0))
	assert.True(t, isTransient(503))
	assert.False(t, isTransient(404))
	assert.False(t, isTransient(200))
}

// HU-17 / R2: si Contentful se cae, se sirve el último contenido válido.
func TestContentfulClient_SirveContenidoStaleSiContentfulFalla(t *testing.T) {
	var calls atomic.Int32
	var status atomic.Int32
	status.Store(http.StatusOK)
	cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(int(status.Load()))
		_, _ = fmt.Fprint(w, `{"items":[{"sys":{"id":"c1"}}],"total":1}`)
	}))
	defer cda.Close()

	client := newTestClient(cda.URL, "")
	client.cache = newContentCache(time.Minute)
	ctx := t.Context()

	first, err := client.ListEntries(ctx, "course", 1)
	require.NoError(t, err)
	require.Len(t, first.Items, 1)
	_, err = client.ListEntries(ctx, "course", 1)
	require.NoError(t, err)
	assert.Equal(t, int32(1), calls.Load(), "la segunda lectura sale de la caché")

	// contenido.publicado → se invalida; Contentful responde 503.
	client.InvalidateCache()
	status.Store(http.StatusServiceUnavailable)
	stale, err := client.ListEntries(ctx, "course", 1)
	require.NoError(t, err, "con Contentful caído se sirve lo último válido")
	assert.Len(t, stale.Items, 1)
	assert.Equal(t, int32(2), calls.Load(), "invalidada: intentó ir a Contentful")

	// Sin nada cacheado para esa consulta, el error se propaga.
	_, err = client.ListEntriesWhere(ctx, "course", 1, map[string]string{"fields.title": "otro"})
	assert.Error(t, err)

	// Un 4xx no es transitorio: no se tapa con contenido viejo.
	client.InvalidateCache()
	status.Store(http.StatusNotFound)
	_, err = client.ListEntries(ctx, "course", 1)
	assert.Error(t, err)
}

func TestContentfulClient_EscrituraPropiaInvalidaCache(t *testing.T) {
	var cdaCalls atomic.Int32
	cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		cdaCalls.Add(1)
		_, _ = fmt.Fprint(w, `{"items":[],"total":0}`)
	}))
	defer cda.Close()
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_, _ = fmt.Fprint(w, `{"sys":{"id":"c1","version":3}}`)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer cma.Close()

	client := newTestClient(cda.URL, cma.URL)
	client.cache = newContentCache(time.Minute)
	ctx := t.Context()

	_, err := client.ListEntries(ctx, "course", 1)
	require.NoError(t, err)
	_, err = client.GetManagementEntry(ctx, "c1") // lectura por CMA: no invalida
	require.NoError(t, err)
	_, err = client.ListEntries(ctx, "course", 1)
	require.NoError(t, err)
	assert.Equal(t, int32(1), cdaCalls.Load())

	require.NoError(t, client.UnpublishEntry(ctx, "c1"))
	_, err = client.ListEntries(ctx, "course", 1)
	require.NoError(t, err)
	assert.Equal(t, int32(2), cdaCalls.Load(), "después de escribir se vuelve a leer de Contentful")
}
