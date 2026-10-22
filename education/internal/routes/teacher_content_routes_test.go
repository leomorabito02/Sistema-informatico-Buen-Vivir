package routes

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
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

// jwtRoles acepta "<rol>:<usuario>" como token, con is_external para
// "externo".
type jwtRoles struct{}

func (jwtRoles) ValidateToken(tok string) (jwt.MapClaims, error) {
	rol, user, ok := strings.Cut(tok, ":")
	if !ok {
		return nil, errors.New("token inválido")
	}
	claims := jwt.MapClaims{"sub": user, "role": rol}
	if rol == "externo" {
		claims["role"], claims["is_external"] = "role_teacher", true
	}
	return claims, nil
}

type memContentRepo struct {
	items     map[string]models.ContenidoDocente
	descargas []models.DescargaBiblioteca
	n         int
}

func (m *memContentRepo) NewID() string { m.n++; return "c" + string(rune('0'+m.n)) }
func (m *memContentRepo) Create(_ context.Context, c *models.ContenidoDocente) error {
	m.items[c.ID] = *c
	return nil
}
func (m *memContentRepo) Get(_ context.Context, id string) (*models.ContenidoDocente, error) {
	c, ok := m.items[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return &c, nil
}
func (m *memContentRepo) list(f func(models.ContenidoDocente) bool) []models.ContenidoDocente {
	out := []models.ContenidoDocente{}
	for _, c := range m.items {
		if f(c) {
			out = append(out, c)
		}
	}
	return out
}
func (m *memContentRepo) ListByDocente(_ context.Context, d string) ([]models.ContenidoDocente, error) {
	return m.list(func(c models.ContenidoDocente) bool { return c.DocenteID == d }), nil
}
func (m *memContentRepo) ListByEstado(_ context.Context, e string) ([]models.ContenidoDocente, error) {
	return m.list(func(c models.ContenidoDocente) bool { return e == "" || c.EstadoModeracion == e }), nil
}
func (m *memContentRepo) Update(_ context.Context, c *models.ContenidoDocente) error {
	m.items[c.ID] = *c
	return nil
}
func (m *memContentRepo) Moderate(_ context.Context, id, estado, motivo, admin string) (*models.ContenidoDocente, error) {
	c, ok := m.items[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	c.EstadoModeracion, c.MotivoRechazo, c.ModeradoPor = estado, motivo, admin
	m.items[id] = c
	return &c, nil
}
func (m *memContentRepo) Delete(_ context.Context, id string) error { delete(m.items, id); return nil }
func (m *memContentRepo) CreateDescarga(_ context.Context, d *models.DescargaBiblioteca) error {
	m.descargas = append(m.descargas, *d)
	return nil
}
func (m *memContentRepo) ListDescargas(context.Context) ([]models.DescargaBiblioteca, error) {
	return m.descargas, nil
}

type memFiles struct{ files map[string][]byte }

func (s *memFiles) Upload(_ context.Context, p, _ string, d []byte) error { s.files[p] = d; return nil }
func (s *memFiles) Open(_ context.Context, p string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s.files[p])), nil
}
func (s *memFiles) DeletePrefix(context.Context, string) error { return nil }

func TestContenidosDocentes_FlujoYPermisos(t *testing.T) {
	repo := &memContentRepo{items: map[string]models.ContenidoDocente{}}
	r := chi.NewRouter()
	RegisterRoutes(r, jwtRoles{}, Controllers{
		TeacherContent: controllers.NewTeacherContentController(
			services.NewTeacherContentService(repo, &memFiles{files: map[string][]byte{}})),
	})
	do := func(method, path, token, contentType string, body io.Reader) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, body)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}
	material := func() (string, io.Reader) {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		_ = mw.WriteField("titulo", "Guía de compost")
		_ = mw.WriteField("descripcion", "Para la huerta")
		fw, _ := mw.CreateFormFile("archivo", "compost.pdf")
		_, _ = fw.Write([]byte("%PDF-1.4"))
		_ = mw.Close()
		return mw.FormDataContentType(), &buf
	}

	// El docente externo sube (las rutas /api/docente le dan 403, estas no).
	ct, body := material()
	w := do("POST", "/api/contenidos-docentes", "externo:usr-ext", ct, body)
	if w.Code != http.StatusCreated {
		t.Fatalf("subir como externo: %d %s", w.Code, w.Body)
	}
	if !strings.Contains(w.Body.String(), `"estadoModeracion":"pendiente"`) || strings.Contains(w.Body.String(), "archivo_ruta") {
		t.Errorf("respuesta inesperada: %s", w.Body)
	}
	id := "c1"

	ct, body = material()
	if w = do("POST", "/api/contenidos-docentes", "role_participant:usr-p", ct, body); w.Code != http.StatusForbidden {
		t.Errorf("participante sube: %d, quiero 403", w.Code)
	}
	if w = do("GET", "/api/contenidos-docentes/mios", "externo:usr-ext", "", nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "Guía de compost") {
		t.Errorf("mis contenidos: %d %s", w.Code, w.Body)
	}

	// Pendiente: no aparece en la biblioteca ni se descarga.
	if w = do("GET", "/api/biblioteca", "", "", nil); w.Body.String() != "[]\n" {
		t.Errorf("biblioteca con pendiente: %s", w.Body)
	}
	datos := `{"nombre":"Ana","apellido":"Gómez","email":"ana@gmail.com","telefono":"3547 123456"}`
	if w = do("POST", "/api/biblioteca/"+id+"/descargas", "", "application/json", strings.NewReader(datos)); w.Code != http.StatusNotFound {
		t.Errorf("descargar pendiente: %d, quiero 404", w.Code)
	}
	if w = do("GET", "/api/contenidos-docentes/"+id+"/archivo", "role_teacher:otro", "", nil); w.Code != http.StatusForbidden {
		t.Errorf("otro docente baja pendiente: %d, quiero 403", w.Code)
	}

	// Moderar es solo del administrador.
	aprobar := `{"estado":"aprobado"}`
	if w = do("POST", "/api/contenidos-docentes/"+id+"/moderacion", "externo:usr-ext", "application/json", strings.NewReader(aprobar)); w.Code != http.StatusForbidden {
		t.Errorf("docente modera: %d, quiero 403", w.Code)
	}
	if w = do("GET", "/api/contenidos-docentes?estado=pendiente", "role_volunteer:v", "", nil); w.Code != http.StatusForbidden {
		t.Errorf("voluntario lista para moderar: %d, quiero 403", w.Code)
	}
	if w = do("GET", "/api/contenidos-docentes?estado=pendiente", "role_admin:a", "", nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), id) {
		t.Errorf("admin lista pendientes: %d %s", w.Code, w.Body)
	}
	if w = do("POST", "/api/contenidos-docentes/"+id+"/moderacion", "role_admin:a", "application/json", strings.NewReader(aprobar)); w.Code != http.StatusOK {
		t.Fatalf("admin aprueba: %d %s", w.Code, w.Body)
	}

	// Aprobado: ya no se edita y aparece en la biblioteca sin el docente.
	ct, body = material()
	if w = do("PUT", "/api/contenidos-docentes/"+id, "externo:usr-ext", ct, body); w.Code != http.StatusBadRequest {
		t.Errorf("editar aprobado: %d, quiero 400", w.Code)
	}
	w = do("GET", "/api/biblioteca", "", "", nil)
	if !strings.Contains(w.Body.String(), "Guía de compost") || strings.Contains(w.Body.String(), "usr-ext") {
		t.Errorf("biblioteca pública: %s", w.Body)
	}

	// Descarga pública: pide los datos y devuelve el archivo.
	if w = do("POST", "/api/biblioteca/"+id+"/descargas", "", "application/json", strings.NewReader(`{"nombre":"Ana"}`)); w.Code != http.StatusBadRequest {
		t.Errorf("descarga sin datos: %d, quiero 400", w.Code)
	}
	w = do("POST", "/api/biblioteca/"+id+"/descargas", "", "application/json", strings.NewReader(datos))
	if w.Code != http.StatusOK || w.Body.String() != "%PDF-1.4" {
		t.Fatalf("descarga pública: %d %q", w.Code, w.Body)
	}
	if cd := w.Header().Get("Content-Disposition"); !strings.Contains(cd, "compost.pdf") {
		t.Errorf("Content-Disposition: %q", cd)
	}

	// Quién descargó: solo el administrador.
	if w = do("GET", "/api/biblioteca/descargas", "", "", nil); w.Code != http.StatusUnauthorized {
		t.Errorf("descargas sin login: %d, quiero 401", w.Code)
	}
	if w = do("GET", "/api/biblioteca/descargas", "role_admin:a", "", nil); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "3547123456") {
		t.Errorf("descargas admin: %d %s", w.Code, w.Body)
	}

	// Borrar: el dueño ya no puede (aprobado); el admin sí.
	if w = do("DELETE", "/api/contenidos-docentes/"+id, "externo:usr-ext", "", nil); w.Code != http.StatusBadRequest {
		t.Errorf("dueño borra aprobado: %d, quiero 400", w.Code)
	}
	if w = do("DELETE", "/api/contenidos-docentes/"+id, "role_admin:a", "", nil); w.Code != http.StatusNoContent {
		t.Errorf("admin borra: %d, quiero 204", w.Code)
	}
}
