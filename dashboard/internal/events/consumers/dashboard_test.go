package consumers

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"buenvivir/dashboard/internal/events"
	"buenvivir/dashboard/internal/services"
)

type fakeCache struct{ invalidated []services.Fuente }

func (f *fakeCache) Invalidate(fuentes ...services.Fuente) {
	f.invalidated = append(f.invalidated, fuentes...)
}

func push(t *testing.T, h http.Handler, env events.Envelope) int {
	t.Helper()
	data, _ := env.Marshal()
	body, _ := events.BuildPushBody(data, env.Attributes(), "m", "s", 1)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/eventos", bytes.NewReader(body)))
	return rec.Code
}

func TestDashboardInvalidaLaFuenteDeCadaEvento(t *testing.T) {
	cache := &fakeCache{}
	c := NewDashboard(events.NoVerifier{}, events.NewMemoryStore(100), cache, slog.New(slog.NewJSONHandler(io.Discard, nil)))

	cases := map[string]services.Fuente{
		events.ParticipanteCreado:          services.FuenteParticipantes,
		events.ParticipanteEliminado:       services.FuenteParticipantes,
		events.CursoCreado:                 services.FuenteCursos,
		events.CursoEliminado:              services.FuenteCursos,
		events.CursoCompletadoParticipante: services.FuenteCursos,
		events.HabilidadOtorgada:           services.FuenteHabilidades,
		events.PostulacionSeleccionada:     services.FuentePostulaciones,
	}
	for eventType, want := range cases {
		cache.invalidated = nil
		env, err := events.New(context.Background(), "test", eventType, map[string]string{"id": "x"})
		if err != nil {
			t.Fatal(err)
		}
		if code := push(t, c, env); code != http.StatusNoContent {
			t.Fatalf("%s: %d", eventType, code)
		}
		if len(cache.invalidated) != 1 || cache.invalidated[0] != want {
			t.Errorf("%s invalidó %v, se esperaba %s", eventType, cache.invalidated, want)
		}
		// Reentrega del mismo evento: se descarta sin volver a invalidar.
		cache.invalidated = nil
		push(t, c, env)
		if len(cache.invalidated) != 0 {
			t.Errorf("%s duplicado no debe reprocesarse", eventType)
		}
	}
	if len(c.EventTypes()) != len(cases) {
		t.Errorf("tipos registrados: %v", c.EventTypes())
	}
}

// Las tres suscripciones de Dashboard, juntas, cubren exactamente los
// tipos que procesa el consumidor.
func TestFiltrosDeInfraCoincidenConElConsumidor(t *testing.T) {
	filters, err := events.InfraFilters("../../../../infra/pubsub/setup.sh")
	if err != nil {
		t.Fatal(err)
	}
	var all []string
	for _, sub := range []string{"dashboard-usuarios", "dashboard-educacion", "dashboard-insercion-laboral"} {
		all = append(all, filters[sub]...)
	}
	if got := NewDashboard(events.NoVerifier{}, nil, nil, nil).EventTypes(); !events.SameTypes(got, all) {
		t.Errorf("consumidor %v, filtros %v", got, all)
	}
}
