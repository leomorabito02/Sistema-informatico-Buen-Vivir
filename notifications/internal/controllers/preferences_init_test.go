package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// preferences/init: un usuario no puede tocar las preferencias de otro;
// solo un servicio interno (role_dev) puede indicar otro user_id.
func TestInitPreferences_SoloServicioPuedeIndicarOtroUsuario(t *testing.T) {
	mux, _, _ := setupTestApp()

	casos := []struct {
		nombre string
		role   string
		body   string
		want   int
	}{
		{"participante pide las suyas", "role_participant", `{}`, http.StatusCreated},
		{"participante manda su propio id", "role_participant", `{"user_id":"usr-a"}`, http.StatusCreated},
		{"participante intenta resetear las de otro", "role_participant", `{"user_id":"usr-b","reset":true}`, http.StatusForbidden},
		{"servicio interno inicializa las de otro", "role_dev", `{"user_id":"usr-b"}`, http.StatusCreated},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/preferences/init", strings.NewReader(c.body))
			req.Header.Set("X-User-ID", "usr-a")
			req.Header.Set("X-User-Role", c.role)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != c.want {
				t.Fatalf("esperaba %d, fue %d: %s", c.want, rec.Code, rec.Body.String())
			}
		})
	}
}
