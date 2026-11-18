package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"

	"buenvivir/campaigns/internal/controllers"
	"buenvivir/campaigns/internal/middlewares"
	"buenvivir/campaigns/internal/models"
	"buenvivir/campaigns/internal/repositories"
	"buenvivir/campaigns/internal/services"
)

const secreto = "secreto-de-prueba-de-al-menos-32-caracteres"

// memRepo: repositorio en memoria para probar la API sin Firestore (las
// pruebas contra el emulador están en el paquete repositories).
type memRepo struct {
	datos map[string]models.Campaign
	sig   int
}

func (m *memRepo) Create(_ context.Context, c models.Campaign) (models.Campaign, error) {
	m.sig++
	c.ID = fmt.Sprintf("cmp%d", m.sig)
	m.datos[c.ID] = c
	return c, nil
}

func (m *memRepo) Get(_ context.Context, id string) (models.Campaign, error) {
	c, ok := m.datos[id]
	if !ok {
		return models.Campaign{}, repositories.ErrNotFound
	}
	return c, nil
}

func (m *memRepo) Update(ctx context.Context, id string, in models.CampaignInput) (models.Campaign, error) {
	c, err := m.Get(ctx, id)
	if err != nil {
		return c, err
	}
	c.Titulo, c.Texto, c.ImagenURL, c.Fecha = in.Titulo, in.Texto, in.ImagenURL, in.Fecha
	m.datos[id] = c
	return c, nil
}

func (m *memRepo) SetEstado(ctx context.Context, id, estado string) (models.Campaign, error) {
	c, err := m.Get(ctx, id)
	if err != nil {
		return c, err
	}
	c.Estado = estado
	m.datos[id] = c
	return c, nil
}

func (m *memRepo) List(_ context.Context, estado string) ([]models.Campaign, error) {
	out := []models.Campaign{}
	for _, c := range m.datos {
		if estado == "" || c.Estado == estado {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (m *memRepo) Delete(_ context.Context, id, _ string) error {
	if _, ok := m.datos[id]; !ok {
		return repositories.ErrNotFound
	}
	delete(m.datos, id)
	return nil
}

func nuevoServidor(t *testing.T) *httptest.Server {
	t.Helper()
	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r, "")
	svc := services.NewCampaignService(&memRepo{datos: map[string]models.Campaign{}})
	RegisterRoutes(r, services.NewJWTServiceWithSecret(secreto), controllers.NewCampaignController(svc))
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

func token(t *testing.T, rol, usuario string) string {
	t.Helper()
	s, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": usuario, "role": rol, "type": "access", "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(secreto))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// pedir hace el pedido y decodifica el JSON de la respuesta en out (si no es nil).
func pedir(t *testing.T, srv *httptest.Server, metodo, ruta, tok, cuerpo string, out any) int {
	t.Helper()
	req, err := http.NewRequest(metodo, srv.URL+ruta, strings.NewReader(cuerpo))
	if err != nil {
		t.Fatal(err)
	}
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	if cuerpo != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			t.Fatalf("%s %s: respuesta no es JSON: %v", metodo, ruta, err)
		}
	}
	return res.StatusCode
}

const cuerpoValido = `{"titulo":"Jornada de reciclaje","texto":"Traé tus botellas.","imagen_url":"https://example.org/a.jpg","fecha":"2026-11-05"}`

type respuesta struct {
	Mensaje string          `json:"mensaje"`
	Campana models.Campaign `json:"campana"`
	Error   string          `json:"error"`
}

func TestFlujoCompleto_CrearEditarPublicarYVerEnPublico(t *testing.T) {
	srv := nuevoServidor(t)
	vol := token(t, "role_volunteer", "voluntario-1")

	// HU-47: crear.
	var creada respuesta
	if code := pedir(t, srv, "POST", "/api/v1/campaigns", vol, cuerpoValido, &creada); code != http.StatusCreated {
		t.Fatalf("crear: status %d, %+v", code, creada)
	}
	if creada.Mensaje != "Campaña creada correctamente" || creada.Campana.Estado != "borrador" || creada.Campana.CreadoPor != "voluntario-1" {
		t.Fatalf("crear: respuesta inesperada %+v", creada)
	}
	id := creada.Campana.ID

	// HU-50: un borrador no se ve en la sección pública.
	var publicas []models.Campaign
	if code := pedir(t, srv, "GET", "/api/v1/campaigns/publicas", "", "", &publicas); code != http.StatusOK || len(publicas) != 0 {
		t.Fatalf("públicas antes de publicar: %d, %+v", code, publicas)
	}

	// HU-48: editar.
	var editada respuesta
	cuerpo := strings.Replace(cuerpoValido, "Jornada de reciclaje", "Gran jornada de reciclaje", 1)
	if code := pedir(t, srv, "PUT", "/api/v1/campaigns/"+id, vol, cuerpo, &editada); code != http.StatusOK || editada.Mensaje != "Campaña editada correctamente" {
		t.Fatalf("editar: %d, %+v", code, editada)
	}

	// HU-49: publicar.
	var publicada respuesta
	if code := pedir(t, srv, "POST", "/api/v1/campaigns/"+id+"/publicar", vol, "", &publicada); code != http.StatusOK || publicada.Campana.Estado != "publicada" {
		t.Fatalf("publicar: %d, %+v", code, publicada)
	}

	// HU-50 y HU-51: sin token, aparece y se encuentra por parte del título.
	if code := pedir(t, srv, "GET", "/api/v1/campaigns/publicas?q=gran%20jornada", "", "", &publicas); code != http.StatusOK || len(publicas) != 1 || publicas[0].Titulo != "Gran jornada de reciclaje" {
		t.Fatalf("búsqueda pública: %d, %+v", code, publicas)
	}
	if publicas[0].CreadoPor != "" {
		t.Errorf("la respuesta pública expone al autor: %q", publicas[0].CreadoPor)
	}

	// Gestión: listado y detalle.
	var todas []models.Campaign
	if code := pedir(t, srv, "GET", "/api/v1/campaigns?estado=publicada", vol, "", &todas); code != http.StatusOK || len(todas) != 1 {
		t.Fatalf("listado de gestión: %d, %+v", code, todas)
	}
	var detalle models.Campaign
	if code := pedir(t, srv, "GET", "/api/v1/campaigns/"+id, vol, "", &detalle); code != http.StatusOK || detalle.ID != id {
		t.Fatalf("detalle: %d, %+v", code, detalle)
	}
}

func TestPermisos(t *testing.T) {
	srv := nuevoServidor(t)
	casos := []struct {
		nombre string
		tok    string
		status int
	}{
		{"sin token", "", http.StatusUnauthorized},
		{"token inválido", "basura", http.StatusUnauthorized},
		{"participante", token(t, "role_participant", "p-1"), http.StatusForbidden},
		{"empresa", token(t, "role_company", "e-1"), http.StatusForbidden},
		{"docente", token(t, "role_teacher", "d-1"), http.StatusForbidden},
		{"voluntario", token(t, "role_volunteer", "v-1"), http.StatusCreated},
		{"administrador", token(t, "role_admin", "a-1"), http.StatusCreated},
	}
	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			if code := pedir(t, srv, "POST", "/api/v1/campaigns", tc.tok, cuerpoValido, nil); code != tc.status {
				t.Fatalf("POST: esperaba %d, obtuve %d", tc.status, code)
			}
		})
	}
	// Las rutas de gestión piden token aunque sean de lectura.
	for _, ruta := range []string{"/api/v1/campaigns", "/api/v1/campaigns/cmp1"} {
		if code := pedir(t, srv, "GET", ruta, "", "", nil); code != http.StatusUnauthorized {
			t.Errorf("GET %s sin token: %d", ruta, code)
		}
	}
	if code := pedir(t, srv, "POST", "/api/v1/campaigns/cmp1/publicar", token(t, "role_participant", "p-1"), "", nil); code != http.StatusForbidden {
		t.Errorf("publicar como participante: %d", code)
	}
}

func TestErrores(t *testing.T) {
	srv := nuevoServidor(t)
	vol := token(t, "role_volunteer", "v-1")
	casos := []struct {
		nombre, metodo, ruta, cuerpo string
		status                       int
		error                        string
	}{
		{"campo desconocido", "POST", "/api/v1/campaigns", `{"titulo":"a","estado":"publicada"}`, 400, "El cuerpo de la solicitud no es válido"},
		{"JSON roto", "POST", "/api/v1/campaigns", `{`, 400, "El cuerpo de la solicitud no es válido"},
		{"faltan campos", "POST", "/api/v1/campaigns", `{}`, 400, "Completar los campos obligatorios: título, texto, imagen, fecha"},
		{"editar inexistente", "PUT", "/api/v1/campaigns/noexiste", cuerpoValido, 404, "La campaña no existe"},
		{"publicar inexistente", "POST", "/api/v1/campaigns/noexiste/publicar", "", 404, "La campaña no existe"},
		{"detalle inexistente", "GET", "/api/v1/campaigns/noexiste", "", 404, "La campaña no existe"},
		{"estado inválido", "GET", "/api/v1/campaigns?estado=archivada", "", 400, `El estado debe ser "borrador" o "publicada"`},
	}
	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			var r respuesta
			if code := pedir(t, srv, tc.metodo, tc.ruta, vol, tc.cuerpo, &r); code != tc.status || r.Error != tc.error {
				t.Fatalf("esperaba %d %q, obtuve %d %q", tc.status, tc.error, code, r.Error)
			}
		})
	}
}

// repoCaido simula que Firestore no responde.
type repoCaido struct{ memRepo }

func (repoCaido) List(context.Context, string) ([]models.Campaign, error) {
	return nil, fmt.Errorf("rpc error: code = Unavailable desc = firestore.googleapis.com no responde")
}

func TestErrorInternoNoExponeDetalles(t *testing.T) {
	r := chi.NewRouter()
	svc := services.NewCampaignService(&repoCaido{memRepo{datos: map[string]models.Campaign{}}})
	RegisterRoutes(r, services.NewJWTServiceWithSecret(secreto), controllers.NewCampaignController(svc))
	srv := httptest.NewServer(r)
	defer srv.Close()

	var res respuesta
	if code := pedir(t, srv, "GET", "/api/v1/campaigns/publicas", "", "", &res); code != http.StatusInternalServerError {
		t.Fatalf("esperaba 500, obtuve %d", code)
	}
	if res.Error != "error interno del servidor" || strings.Contains(res.Error, "firestore") {
		t.Fatalf("el error filtra detalles internos: %q", res.Error)
	}
}

func TestHealthYHeadersDeSeguridad(t *testing.T) {
	srv := nuevoServidor(t)
	res, err := srv.Client().Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK || res.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("health: %d, headers %v", res.StatusCode, res.Header)
	}
}

// Problema 5 de la auditoría (2026-11-11): admin y voluntario dan de baja;
// la campaña sale de la gestión y de la landing; participante, empresa,
// docente y sin sesión no pueden.
func TestEliminarCampana_PermisosYEfecto(t *testing.T) {
	srv := nuevoServidor(t)
	vol := token(t, "role_volunteer", "voluntario-1")
	admin := token(t, "role_admin", "admin-1")

	crear := func() string {
		var r respuesta
		if code := pedir(t, srv, http.MethodPost, "/api/v1/campaigns", vol, cuerpoValido, &r); code != http.StatusCreated {
			t.Fatalf("crear: %d", code)
		}
		if code := pedir(t, srv, http.MethodPost, "/api/v1/campaigns/"+r.Campana.ID+"/publicar", vol, "", nil); code != http.StatusOK {
			t.Fatalf("publicar: %d", code)
		}
		return r.Campana.ID
	}

	id := crear()
	for _, rol := range []string{"role_participant", "role_company", "role_teacher"} {
		if code := pedir(t, srv, http.MethodDelete, "/api/v1/campaigns/"+id, token(t, rol, "u"), "", nil); code != http.StatusForbidden {
			t.Errorf("%s: esperaba 403, dio %d", rol, code)
		}
	}
	if code := pedir(t, srv, http.MethodDelete, "/api/v1/campaigns/"+id, "", "", nil); code != http.StatusUnauthorized {
		t.Errorf("sin sesión: esperaba 401, dio %d", code)
	}

	var r respuesta
	if code := pedir(t, srv, http.MethodDelete, "/api/v1/campaigns/"+id, vol, "", &r); code != http.StatusOK {
		t.Fatalf("voluntario elimina: esperaba 200, dio %d", code)
	}
	if r.Mensaje != services.MsgCampanaEliminada {
		t.Errorf("mensaje: %q", r.Mensaje)
	}
	var publicas []models.Campaign
	pedir(t, srv, http.MethodGet, "/api/v1/campaigns/publicas", "", "", &publicas)
	if len(publicas) != 0 {
		t.Errorf("la landing sigue mostrando la campaña eliminada: %v", publicas)
	}
	if code := pedir(t, srv, http.MethodDelete, "/api/v1/campaigns/"+id, vol, "", nil); code != http.StatusNotFound {
		t.Errorf("segunda baja: esperaba 404, dio %d", code)
	}

	id2 := crear()
	if code := pedir(t, srv, http.MethodDelete, "/api/v1/campaigns/"+id2, admin, "", nil); code != http.StatusOK {
		t.Errorf("admin elimina: esperaba 200, dio %d", code)
	}
}
