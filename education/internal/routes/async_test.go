package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"buenvivir/education/internal/events"
)

func TestWithAsync(t *testing.T) {
	ok := func(code int) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(code) })
	}
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if events.CorrelationIDFrom(r.Context()) == "" {
			t.Error("la API también recibe correlation id")
		}
		w.WriteHeader(http.StatusTeapot)
	})
	panics := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") })

	h := WithAsync(api, AsyncEndpoints{ContentfulWebhook: ok(http.StatusAccepted), Usuarios: ok(http.StatusNoContent), Contenido: panics})

	cases := []struct {
		method, path string
		want         int
	}{
		{http.MethodPost, "/webhooks/contentful", http.StatusAccepted},
		{http.MethodPost, "/eventos/usuarios", http.StatusNoContent},
		{http.MethodPost, "/eventos/contenido", http.StatusInternalServerError}, // Recoverer
		{http.MethodPost, "/internal/outbox/flush", http.StatusTeapot},          // no configurado: cae en la API
		{http.MethodGet, "/api/cursos", http.StatusTeapot},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(c.method, c.path, nil))
		assert.Equal(t, c.want, rec.Code, "%s %s", c.method, c.path)
		assert.NotEmpty(t, rec.Header().Get(events.HeaderCorrelationID))
	}
}
