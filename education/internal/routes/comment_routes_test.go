package routes

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
	"buenvivir/education/internal/services"
)

// fakeJWT acepta "admin" y "participante" como tokens.
type fakeJWT struct{}

func (fakeJWT) ValidateToken(tok string) (jwt.MapClaims, error) {
	switch tok {
	case "admin":
		return jwt.MapClaims{"sub": "usr-admin", "role": "role_admin"}, nil
	case "participante":
		return jwt.MapClaims{"sub": "usr-part", "role": "role_participant"}, nil
	}
	return nil, errors.New("token inválido")
}

type memComments struct{ items []models.Comentario }

func (m *memComments) ListByEntidad(_ context.Context, tipo, id string) ([]models.Comentario, error) {
	out := []models.Comentario{}
	for _, c := range m.items {
		if c.EntidadTipo == tipo && c.EntidadID == id {
			out = append(out, c)
		}
	}
	return out, nil
}

func (m *memComments) Create(_ context.Context, c *models.Comentario) error {
	c.ID = "com-1"
	m.items = append(m.items, *c)
	return nil
}

func (m *memComments) Delete(_ context.Context, id string) error {
	for i, c := range m.items {
		if c.ID == id {
			m.items = append(m.items[:i], m.items[i+1:]...)
			return nil
		}
	}
	return repositories.ErrNotFound
}

func TestComentarios_RutasYPermisos(t *testing.T) {
	r := chi.NewRouter()
	RegisterRoutes(r, fakeJWT{}, Controllers{
		Comment: controllers.NewCommentController(services.NewCommentService(&memComments{})),
	})
	do := func(method, path, body, token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	// Sin login se comenta y se lee (HU-24: contenido público).
	w := do("POST", "/api/comentarios", `{"entidadTipo":"campana","entidadId":"camp-1","nombre":"Ana","apellido":"Gómez","texto":"<b>Genial</b>"}`, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("POST sin login: %d %s", w.Code, w.Body)
	}
	if strings.Contains(w.Body.String(), "<b>") {
		t.Errorf("el texto tiene que volver sanitizado: %s", w.Body)
	}
	if strings.Contains(w.Body.String(), "usuario") {
		t.Errorf("el usuario_id no se expone: %s", w.Body)
	}

	w = do("GET", "/api/comentarios?entidadTipo=campana&entidadId=camp-1", "", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Genial") {
		t.Fatalf("GET sin login: %d %s", w.Code, w.Body)
	}

	w = do("POST", "/api/comentarios", `{"entidadTipo":"campana","entidadId":"camp-1","texto":"sin nombre"}`, "")
	if w.Code != http.StatusBadRequest {
		t.Errorf("sin nombre: %d, quiero 400", w.Code)
	}

	// Moderar es solo del administrador.
	if w = do("DELETE", "/api/comentarios/com-1", "", ""); w.Code != http.StatusUnauthorized {
		t.Errorf("DELETE sin login: %d, quiero 401", w.Code)
	}
	if w = do("DELETE", "/api/comentarios/com-1", "", "participante"); w.Code != http.StatusForbidden {
		t.Errorf("DELETE participante: %d, quiero 403", w.Code)
	}
	if w = do("DELETE", "/api/comentarios/com-1", "", "admin"); w.Code != http.StatusNoContent {
		t.Errorf("DELETE admin: %d, quiero 204", w.Code)
	}
	if w = do("DELETE", "/api/comentarios/com-1", "", "admin"); w.Code != http.StatusNotFound {
		t.Errorf("DELETE de nuevo: %d, quiero 404", w.Code)
	}
}

func TestComentarios_LimiteDeComentariosPorIP(t *testing.T) {
	r := chi.NewRouter()
	RegisterRoutes(r, fakeJWT{}, Controllers{
		Comment: controllers.NewCommentController(services.NewCommentService(&memComments{})),
	})
	body := `{"entidadTipo":"curso","entidadId":"c-1","nombre":"Ana","apellido":"Gómez","texto":"hola"}`
	var ultimo int
	for i := 0; i < 11; i++ {
		req := httptest.NewRequest("POST", "/api/comentarios", strings.NewReader(body))
		req.RemoteAddr = "10.0.0.9:1234"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		ultimo = w.Code
	}
	if ultimo != http.StatusTooManyRequests {
		t.Errorf("el comentario 11 en un minuto: %d, quiero 429", ultimo)
	}
}
