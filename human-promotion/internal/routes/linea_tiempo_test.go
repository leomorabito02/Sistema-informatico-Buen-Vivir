package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"buenvivir/human-promotion/internal/controllers"
	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/repositories"
	"buenvivir/human-promotion/internal/services"
)

type actividadFalsa struct{}

func (actividadFalsa) DeParticipante(context.Context, string) ([]models.EntradaActividad, error) {
	return nil, nil
}

func (actividadFalsa) DeFicha(_ context.Context, fichaID string) ([]models.EntradaActividad, error) {
	if fichaID != "f1" {
		return nil, repositories.ErrNotFound
	}
	return []models.EntradaActividad{{ID: "i1", Tipo: "inscripcion", Fecha: "2026-09-02", Descripcion: "Se inscribió en un curso", CursoID: "c1"}}, nil
}

type participanteFalso struct{ services.ParticipanteService }

func (participanteFalso) LineaDeTiempoPropia(context.Context, string) (*models.LineaDeTiempoPropia, error) {
	return &models.LineaDeTiempoPropia{
		Hitos:     []models.HitoParticipante{{ID: "h1", Fecha: "2026-09-01", Descripcion: "Primera entrevista"}},
		Actividad: []models.EntradaActividad{{ID: "asistencia-2026-09", Tipo: "asistencia_general", Fecha: "2026-09-05", Descripcion: "Asistió al centro en septiembre de 2026"}},
	}, nil
}

func servidorLineaTiempo(t *testing.T) *chi.Mux {
	t.Helper()
	t.Setenv("JWT_SECRET", secretoPrueba)
	r := chi.NewRouter()
	RegisterRoutes(r, services.NewJWTService(), Controllers{
		Actividad:    controllers.NewActividadController(actividadFalsa{}),
		Participante: controllers.NewParticipanteController(participanteFalso{}),
	})
	return r
}

func get(r http.Handler, path, rol string, t *testing.T) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if rol != "" {
		req.Header.Set("Authorization", "Bearer "+token(t, rol))
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// La actividad de una ficha es parte de la vista profesional (HU-33):
// mismos permisos que los hitos.
func TestActividadDeFicha_Permisos(t *testing.T) {
	r := servidorLineaTiempo(t)
	for rol, esperado := range map[string]int{
		"role_admin":       http.StatusOK,
		"role_volunteer":   http.StatusOK,
		"role_dev":         http.StatusOK,
		"role_participant": http.StatusForbidden,
		"role_teacher":     http.StatusForbidden,
		"":                 http.StatusUnauthorized,
	} {
		if rec := get(r, "/api/v1/human-promotion/fichas/f1/actividad", rol, t); rec.Code != esperado {
			t.Errorf("rol %q: esperaba %d, obtuve %d", rol, esperado, rec.Code)
		}
	}
}

func TestActividadDeFicha_Respuesta(t *testing.T) {
	r := servidorLineaTiempo(t)
	rec := get(r, "/api/v1/human-promotion/fichas/f1/actividad", "role_volunteer", t)
	var got []models.EntradaActividad
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil || len(got) != 1 || got[0].CursoID != "c1" {
		t.Fatalf("respuesta inesperada: %v %+v", err, got)
	}

	if rec := get(r, "/api/v1/human-promotion/fichas/otra/actividad", "role_admin", t); rec.Code != http.StatusNotFound {
		t.Errorf("ficha inexistente: esperaba 404, obtuve %d", rec.Code)
	}
}

func TestMiLineaDeTiempo_HitosYActividad(t *testing.T) {
	r := servidorLineaTiempo(t)
	rec := get(r, "/api/v1/human-promotion/mi-linea-de-tiempo", "role_participant", t)
	if rec.Code != http.StatusOK {
		t.Fatalf("esperaba 200, obtuve %d", rec.Code)
	}
	var got struct {
		Hitos     []map[string]any `json:"hitos"`
		Actividad []map[string]any `json:"actividad"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got.Hitos) != 1 || len(got.Actividad) != 1 {
		t.Errorf("esperaba hitos y actividad: %+v", got)
	}

	if rec := get(r, "/api/v1/human-promotion/mi-linea-de-tiempo", "role_volunteer", t); rec.Code != http.StatusForbidden {
		t.Errorf("solo el participante ve su línea: esperaba 403, obtuve %d", rec.Code)
	}
}
