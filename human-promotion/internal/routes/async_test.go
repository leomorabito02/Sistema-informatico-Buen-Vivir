package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"buenvivir/human-promotion/internal/events"
)

func TestWithAsync(t *testing.T) {
	status := func(code int) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(code) })
	}
	h := WithAsync(status(http.StatusTeapot), AsyncEndpoints{
		Educacion: status(http.StatusNoContent),
		Usuarios:  http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }),
	})
	cases := map[string]int{
		"/eventos/educacion":             http.StatusNoContent,
		"/eventos/usuarios":              http.StatusInternalServerError, // Recoverer
		"/internal/outbox/flush":         http.StatusTeapot,              // no configurado: cae en la API
		"/api/v1/human-promotion/fichas": http.StatusTeapot,
	}
	for path, want := range cases {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path, nil))
		if rec.Code != want {
			t.Errorf("%s: %d, se esperaba %d", path, rec.Code, want)
		}
		if rec.Header().Get(events.HeaderCorrelationID) == "" {
			t.Errorf("%s: falta el correlation id", path)
		}
	}
}
