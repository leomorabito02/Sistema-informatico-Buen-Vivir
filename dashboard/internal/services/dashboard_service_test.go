package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"buenvivir/dashboard/internal/clients"
	"buenvivir/dashboard/internal/models"
	"buenvivir/dashboard/internal/services"
)

var cred = clients.Credentials{
	Cookie:        "access_token=jwt",
	Authorization: "Bearer jwt",
	UserID:        "usr-admin",
	UserRole:      "role_admin",
}

func fecha(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// fake arma un servicio de origen: path → respuesta JSON (o status).
type fake struct {
	routes map[string]any
	status map[string]int
	hits   atomic.Int32
	// Último pedido recibido, para revisar las credenciales reenviadas.
	last atomic.Pointer[http.Request]
}

func newFake(t *testing.T, routes map[string]any) (*fake, *clients.Upstream) {
	f := &fake{routes: routes, status: map[string]int{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.hits.Add(1)
		f.last.Store(r)
		if st, ok := f.status[r.URL.Path]; ok {
			w.WriteHeader(st)
			return
		}
		body, ok := f.routes[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if s, isRaw := body.(string); isRaw {
			_, _ = w.Write([]byte(s))
			return
		}
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	return f, clients.NewUpstream("fake", srv.URL, 5*time.Second)
}

func svcWith(auth, edu, labor *clients.Upstream) services.DashboardService {
	return services.NewDashboardService(auth, edu, labor, nil)
}

// ── Participantes ──────────────────────────────────────────────────────

func TestMapParticipante(t *testing.T) {
	alta := fecha("2026-01-10T00:00:00Z")
	aprob := fecha("2026-01-15T00:00:00Z")
	baja := fecha("2026-04-01T00:00:00Z")
	mod := fecha("2026-05-01T00:00:00Z")

	p, ok := services.MapParticipante(models.ParticipantHistory{UserID: "u1", Status: "active", ApprovalStatus: "approved", CreatedAt: alta, ApprovalDate: &aprob})
	if !ok || p.Estado != "activo" || !p.FechaAlta.Equal(aprob) || p.FechaBaja != nil {
		t.Fatalf("activo aprobado: %+v", p)
	}

	p, ok = services.MapParticipante(models.ParticipantHistory{UserID: "u2", Status: "active", ApprovalStatus: "accepted", CreatedAt: alta, DeletedAt: &baja})
	if !ok || p.Estado != "baja" || !p.FechaBaja.Equal(baja) || !p.FechaAlta.Equal(alta) {
		t.Fatalf("baja lógica: %+v", p)
	}

	p, ok = services.MapParticipante(models.ParticipantHistory{UserID: "u3", Status: "blocked", ApprovalStatus: "approved", CreatedAt: alta, UpdatedAt: mod})
	if !ok || p.Estado != "baja" || !p.FechaBaja.Equal(mod) {
		t.Fatalf("bloqueado: %+v", p)
	}

	if _, ok := services.MapParticipante(models.ParticipantHistory{UserID: "u4", Status: "active", ApprovalStatus: "pending", CreatedAt: alta}); ok {
		t.Fatal("un participante pendiente de aprobación no cuenta")
	}
}

func TestParticipantes(t *testing.T) {
	f, auth := newFake(t, map[string]any{
		"/api/v1/users/participants/history": map[string]any{
			"success": true,
			"data": []map[string]any{
				{"user_id": "u1", "status": "active", "approval_status": "approved", "created_at": "2026-01-10T00:00:00Z"},
				{"user_id": "u2", "status": "active", "approval_status": "pending", "created_at": "2026-01-10T00:00:00Z"},
			},
		},
	})
	got, err := svcWith(auth, nil, nil).Participantes(context.Background(), cred)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != "u1" || got[0].Rol != "participante" {
		t.Fatalf("participantes: %+v", got)
	}
	r := f.last.Load()
	if r.Header.Get("Cookie") != cred.Cookie || r.Header.Get("Authorization") != cred.Authorization {
		t.Fatal("reenvía las credenciales del admin a autenticación")
	}
}

func TestParticipantes_Error(t *testing.T) {
	f, auth := newFake(t, nil)
	f.status["/api/v1/users/participants/history"] = http.StatusForbidden
	_, err := svcWith(auth, nil, nil).Participantes(context.Background(), cred)
	var up *clients.UpstreamError
	if !errors.As(err, &up) || up.Status != http.StatusForbidden {
		t.Fatalf("esperaba UpstreamError 403, llegó %v", err)
	}
}

// ── Cursos y áreas ─────────────────────────────────────────────────────

func educationFake(t *testing.T) (*fake, *clients.Upstream) {
	return newFake(t, map[string]any{
		"/api/cursos": []map[string]any{
			{"id": "c1", "title": "Huerta", "trainingArea": map[string]any{"id": "a-huerta", "name": "Huerta"}, "createdAt": "2026-01-05T10:00:00Z"},
			{"id": "c2", "title": "Sin área", "createdAt": "fecha-rota"},
		},
		"/api/cursos/borradores": []map[string]any{
			{"id": "c3", "title": "Borrador", "trainingArea": map[string]any{"id": "a-oficios", "name": "Oficios"}, "createdAt": "2026-02-01T00:00:00Z"},
		},
		"/api/cursos/c1/actividades": []map[string]any{
			{"id": "act1", "fecha": "2026-02-01T00:00:00Z"},
			{"id": "act2", "fecha": "2026-03-01T00:00:00Z"},
		},
		"/api/cursos/c2/actividades": []map[string]any{},
		"/api/cursos/c3/actividades": []map[string]any{{"id": "act3", "fecha": "2026-04-01T00:00:00Z"}},
		"/api/areas-formativas": []map[string]any{
			{"id": "a-huerta", "name": "Huerta y agroecología"},
			{"id": "a-oficios", "name": "Oficios"},
		},
		"/api/habilidades/otorgadas": []map[string]any{
			{"participanteId": "u1", "habilidadId": "armado-de-huerta", "origen": "actividad", "fecha": "2026-02-02T00:00:00Z", "cursoId": "c1"},
			{"participanteId": "u2", "habilidadId": "uso-de-excel", "origen": "manual", "fecha": "2026-03-02T00:00:00Z"},
		},
	})
}

func TestCursos(t *testing.T) {
	_, edu := educationFake(t)
	got, err := svcWith(nil, edu, nil).Cursos(context.Background(), cred)
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(got, func(i, j int) bool { return got[i].ID < got[j].ID })
	if len(got) != 3 {
		t.Fatalf("publicados + borradores: %+v", got)
	}
	c1, c2, c3 := got[0], got[1], got[2]
	if c1.Estado != "activo" || c1.AreaID != "a-huerta" || len(c1.Actividades) != 2 || !c1.FechaCreacion.Equal(fecha("2026-01-05T10:00:00Z")) {
		t.Fatalf("c1: %+v", c1)
	}
	if c2.AreaID != "" || !c2.FechaCreacion.IsZero() || c2.Actividades == nil || len(c2.Actividades) != 0 {
		t.Fatalf("c2 sin área ni fecha válida, con lista vacía de actividades: %+v", c2)
	}
	if c3.Estado != "borrador" || c3.Actividades[0].ID != "act3" {
		t.Fatalf("c3: %+v", c3)
	}
}

func TestCursos_FallaUnaActividad(t *testing.T) {
	f, edu := educationFake(t)
	f.status["/api/cursos/c2/actividades"] = http.StatusBadGateway
	if _, err := svcWith(nil, edu, nil).Cursos(context.Background(), cred); err == nil {
		t.Fatal("si no se pueden leer las actividades de un curso, el indicador no es confiable")
	}
}

func TestCursos_FallanBorradores(t *testing.T) {
	f, edu := educationFake(t)
	f.status["/api/cursos/borradores"] = http.StatusForbidden
	if _, err := svcWith(nil, edu, nil).Cursos(context.Background(), cred); err == nil {
		t.Fatal("esperaba error")
	}
}

func TestAreas(t *testing.T) {
	_, edu := educationFake(t)
	got, err := svcWith(nil, edu, nil).Areas(context.Background(), cred)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Nombre != "Huerta y agroecología" {
		t.Fatalf("areas: %+v", got)
	}
}

func TestAreas_Error(t *testing.T) {
	f, edu := educationFake(t)
	f.status["/api/areas-formativas"] = http.StatusInternalServerError
	if _, err := svcWith(nil, edu, nil).Areas(context.Background(), cred); err == nil {
		t.Fatal("esperaba error")
	}
}

// ── Habilidades ────────────────────────────────────────────────────────

func TestHabilidades(t *testing.T) {
	_, edu := educationFake(t)
	got, err := svcWith(nil, edu, nil).Habilidades(context.Background(), cred)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("habilidades: %+v", got)
	}
	if got[0].AreaID != "a-huerta" || got[0].Estado != "aceptada" || got[0].ID != "u1_armado-de-huerta" || got[0].Origen != "actividad" {
		t.Fatalf("toma el área del curso: %+v", got[0])
	}
	if got[1].AreaID != "" || got[1].CursoID != "" {
		t.Fatalf("sin curso no hay área: %+v", got[1])
	}
}

func TestHabilidades_Errores(t *testing.T) {
	f, edu := educationFake(t)
	f.status["/api/habilidades/otorgadas"] = http.StatusBadGateway
	if _, err := svcWith(nil, edu, nil).Habilidades(context.Background(), cred); err == nil {
		t.Fatal("esperaba error")
	}
	delete(f.status, "/api/habilidades/otorgadas")
	f.status["/api/cursos"] = http.StatusBadGateway
	if _, err := svcWith(nil, edu, nil).Habilidades(context.Background(), cred); err == nil {
		t.Fatal("sin cursos no se puede asignar área")
	}
}

// ── Postulaciones ──────────────────────────────────────────────────────

func TestMapPostulacion(t *testing.T) {
	upd := fecha("2026-05-01T00:00:00Z")
	post := fecha("2026-04-01T00:00:00Z")
	cases := map[string]string{
		"Aceptado":     "seleccionado",
		"seleccionado": "seleccionado",
		"Rechazado":    "rechazado",
		"En revisión":  "en_revision",
		"Postulado":    "postulado",
		"Pendiente":    "postulado",
	}
	for in, want := range cases {
		got := services.MapPostulacion(models.JobApplication{ID: "x", Status: in, UpdatedAt: upd, ApplicationDate: post})
		if got.Estado != want {
			t.Errorf("%q → %q, esperaba %q", in, got.Estado, want)
		}
		if !got.FechaEstado.Equal(upd) {
			t.Errorf("la fecha es la del último cambio de estado")
		}
	}
	sinUpd := services.MapPostulacion(models.JobApplication{Status: "Postulado", ApplicationDate: post})
	if !sinUpd.FechaEstado.Equal(post) {
		t.Error("sin updated_at, usa la fecha de postulación")
	}
}

func TestPostulaciones(t *testing.T) {
	f, labor := newFake(t, map[string]any{
		"/api/v1/labor-integration/offers": []map[string]any{{"id": "o1"}, {"id": "o2"}},
		"/api/v1/labor-integration/offers/o1/applications": []map[string]any{
			{"id": "p1", "participante_id": "u1", "estado": "Aceptado", "updated_at": "2026-03-01T00:00:00Z"},
		},
		"/api/v1/labor-integration/offers/o2/applications": "null",
	})
	got, err := svcWith(nil, nil, labor).Postulaciones(context.Background(), cred)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Estado != "seleccionado" || got[0].ParticipanteID != "u1" {
		t.Fatalf("postulaciones: %+v", got)
	}
	r := f.last.Load()
	if r.Header.Get("X-User-Role") != "role_admin" || r.Header.Get("X-User-ID") != "usr-admin" {
		t.Fatal("inserción laboral recibe el rol y el usuario del admin")
	}
}

func TestPostulaciones_Errores(t *testing.T) {
	f, labor := newFake(t, map[string]any{
		"/api/v1/labor-integration/offers": []map[string]any{{"id": "o1"}},
	})
	if _, err := svcWith(nil, nil, labor).Postulaciones(context.Background(), cred); err == nil {
		t.Fatal("falla si no se pueden leer las postulaciones de una oferta")
	}
	f.status["/api/v1/labor-integration/offers"] = http.StatusBadGateway
	if _, err := svcWith(nil, nil, labor).Postulaciones(context.Background(), cred); err == nil {
		t.Fatal("falla si no se pueden leer las ofertas")
	}
}

func TestPostulaciones_SinOfertas(t *testing.T) {
	_, labor := newFake(t, map[string]any{"/api/v1/labor-integration/offers": []map[string]any{}})
	got, err := svcWith(nil, nil, labor).Postulaciones(context.Background(), cred)
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("sin ofertas: lista vacía, no null (%v, %v)", got, err)
	}
}

// ── Errores de transporte ──────────────────────────────────────────────

func TestUpstream_SinRespuestaYJSONInvalido(t *testing.T) {
	caido := clients.NewUpstream("caído", "http://127.0.0.1:1", time.Second)
	_, err := svcWith(caido, nil, nil).Participantes(context.Background(), cred)
	var up *clients.UpstreamError
	if !errors.As(err, &up) || up.Status != 0 || !strings.Contains(err.Error(), "no respondió") {
		t.Fatalf("sin respuesta: %v", err)
	}

	_, roto := newFake(t, map[string]any{"/api/areas-formativas": "{no es json"})
	_, err = svcWith(nil, roto, nil).Areas(context.Background(), cred)
	if !errors.As(err, &up) || up.Status != http.StatusOK || errors.Unwrap(err) == nil {
		t.Fatalf("JSON inválido: %v", err)
	}
	if !strings.Contains(err.Error(), "respondió 200") {
		t.Fatalf("mensaje: %v", err)
	}

	bad := &clients.Upstream{Name: "mal", BaseURL: "http://[::1]:namedport", HTTP: http.DefaultClient}
	if err := bad.GetJSON(context.Background(), "/x", cred, &struct{}{}); err == nil {
		t.Fatal("URL inválida")
	}
}

// ── Promoción Humana ───────────────────────────────────────────────────

func TestPromocion_TraduceFichasEHitos(t *testing.T) {
	_, ph := upstreamPromocion(t, `{"fichas":[{"id":"f1","participante_id":"p1","estado":"","creada_en":"2026-03-01T00:00:00Z","fecha_baja":null},{"id":"f2","participante_id":"p2","estado":"baja","creada_en":"2026-04-01T00:00:00Z","fecha_baja":"2026-09-01T00:00:00Z"}],"hitos":[{"participante_id":"p1","fecha":"2026-09-10"}]}`)
	got, err := services.NewDashboardService(nil, nil, nil, ph).Promocion(context.Background(), clients.Credentials{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Fichas) != 2 || got.Fichas[0].Estado != "activa" || got.Fichas[1].Estado != "baja" || got.Fichas[1].FechaBaja == nil {
		t.Fatalf("fichas: %+v", got.Fichas)
	}
	if !got.Fichas[0].FechaAlta.Equal(time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("fecha de alta: %v", got.Fichas[0].FechaAlta)
	}
	if len(got.Hitos) != 1 || got.Hitos[0].ParticipanteID != "p1" || got.Hitos[0].Fecha != "2026-09-10" {
		t.Fatalf("hitos: %+v", got.Hitos)
	}
}

func TestPromocion_SinConfigurarYConError(t *testing.T) {
	if _, err := services.NewDashboardService(nil, nil, nil, nil).Promocion(context.Background(), clients.Credentials{}); !errors.Is(err, clients.ErrNoConfigurado) {
		t.Fatalf("sin URL: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusForbidden) }))
	defer srv.Close()
	_, err := services.NewDashboardService(nil, nil, nil, clients.NewUpstream("ph", srv.URL, time.Second)).Promocion(context.Background(), clients.Credentials{})
	var up *clients.UpstreamError
	if !errors.As(err, &up) || up.Status != http.StatusForbidden {
		t.Fatalf("403: %v", err)
	}
}

func TestPromocion_SinDatosDevuelveListasVacias(t *testing.T) {
	_, ph := upstreamPromocion(t, `{"fichas":[],"hitos":[]}`)
	got, err := services.NewDashboardService(nil, nil, nil, ph).Promocion(context.Background(), clients.Credentials{})
	if err != nil || got.Fichas == nil || got.Hitos == nil {
		t.Fatalf("%+v %v", got, err)
	}
}

func upstreamPromocion(t *testing.T, body string) (*httptest.Server, *clients.Upstream) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/human-promotion/indicadores" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, clients.NewUpstream("ph", srv.URL, 2*time.Second)
}
