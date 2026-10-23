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
	"buenvivir/education/internal/services"
)

// jwtPorRol acepta como token el nombre del rol ("role_admin", ...).
type jwtPorRol struct{}

func (jwtPorRol) ValidateToken(tok string) (jwt.MapClaims, error) {
	if tok == "" {
		return nil, errors.New("token inválido")
	}
	return jwt.MapClaims{"sub": "usr-" + tok, "role": tok}, nil
}

type participantesFalsos struct{}

func (participantesFalsos) Buscar(_ context.Context, q, cursoID, _ string) ([]models.ParticipanteResumen, error) {
	if q == "x" {
		return nil, services.NewValidationError("Escribí al menos 2 letras del nombre o del DNI")
	}
	return []models.ParticipanteResumen{{ParticipanteID: "p1", Nombre: q, Apellido: cursoID}}, nil
}

func (participantesFalsos) HistorialAsistencia(_ context.Context, id string) ([]models.HistorialCurso, error) {
	return []models.HistorialCurso{{CursoID: "c-" + id}}, nil
}

type auditoriaFalsa struct{}

func (auditoriaFalsa) List(_ context.Context, f services.FiltroAuditoria, _ string) ([]models.RegistroAuditoria, error) {
	return []models.RegistroAuditoria{{ID: "1", Recurso: f.Recurso}}, nil
}

func TestParticipantesYAuditoria_RutasYPermisos(t *testing.T) {
	r := chi.NewRouter()
	RegisterRoutes(r, jwtPorRol{}, Controllers{
		Participante: controllers.NewParticipanteController(participantesFalsos{}),
		Auditoria:    controllers.NewAuditoriaController(auditoriaFalsa{}),
	})
	do := func(path, rol string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		if rol != "" {
			req.Header.Set("Authorization", "Bearer "+rol)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w
	}

	casos := []struct {
		path, rol string
		want      int
	}{
		{"/api/participantes?q=ana&cursoId=c1", "role_volunteer", http.StatusOK},
		{"/api/participantes?q=ana", "role_admin", http.StatusOK},
		{"/api/participantes?q=x", "role_admin", http.StatusBadRequest},
		{"/api/participantes?q=ana", "role_participant", http.StatusForbidden},
		{"/api/participantes?q=ana", "role_teacher", http.StatusForbidden},
		{"/api/participantes?q=ana", "", http.StatusUnauthorized},
		{"/api/participants/p1/asistencia", "role_volunteer", http.StatusOK},
		{"/api/participants/p1/asistencia", "role_participant", http.StatusForbidden},
		{"/api/auditoria?recurso=curso", "role_admin", http.StatusOK},
		{"/api/auditoria", "role_volunteer", http.StatusForbidden},
		{"/api/auditoria", "role_teacher", http.StatusForbidden},
	}
	for _, c := range casos {
		if w := do(c.path, c.rol); w.Code != c.want {
			t.Errorf("%s como %q: %d, se esperaba %d (%s)", c.path, c.rol, w.Code, c.want, w.Body)
		}
	}

	if body := do("/api/participantes?q=ana&cursoId=c1", "role_admin").Body.String(); body != `[{"participanteId":"p1","nombre":"ana","apellido":"c1"}]`+"\n" {
		t.Errorf("buscador: %s", body)
	}
	if body := do("/api/participants/p7/asistencia", "role_admin").Body.String(); !strings.Contains(body, `"cursoId":"c-p7"`) {
		t.Errorf("historial: %s", body)
	}
	if body := do("/api/auditoria?recurso=curso", "role_admin").Body.String(); !strings.Contains(body, `"recurso":"curso"`) {
		t.Errorf("auditoría: %s", body)
	}
}
