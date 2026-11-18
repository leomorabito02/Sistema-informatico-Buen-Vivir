package consumers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"buenvivir/human-promotion/internal/events"
	"buenvivir/human-promotion/internal/repositories"
)

type fakeLinea struct {
	entries map[string]repositories.EntradaLineaTiempo
	err     error
}

func (f *fakeLinea) Add(_ context.Context, eventID string, e repositories.EntradaLineaTiempo) error {
	if f.err != nil {
		return f.err
	}
	if f.entries == nil {
		f.entries = map[string]repositories.EntradaLineaTiempo{}
	}
	f.entries[eventID] = e
	return nil
}

type fakeFichas struct {
	calls []string
	err   error
}

func (f *fakeFichas) DeactivateByParticipante(_ context.Context, participanteID string) (bool, error) {
	f.calls = append(f.calls, participanteID)
	return true, f.err
}

func quiet() *slog.Logger { return slog.New(slog.NewJSONHandler(io.Discard, nil)) }

func push(t *testing.T, h http.Handler, eventType string, payload any) int {
	t.Helper()
	env, err := events.New(context.Background(), events.ProducerEducacion, eventType, payload)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := env.Marshal()
	body, _ := events.BuildPushBody(data, env.Attributes(), "m", "s", 1)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/eventos", bytes.NewReader(body)))
	return rec.Code
}

func TestEducacionAgregaEntradasSinNumeros(t *testing.T) {
	linea := &fakeLinea{}
	c := NewEducacion(events.NoVerifier{}, events.NewMemoryStore(10), linea, quiet())
	at := time.Date(2026, 10, 16, 15, 0, 0, 0, time.UTC)

	cases := []struct {
		eventType string
		payload   any
		tipo      string
		fecha     string
	}{
		{events.InscripcionCreada, events.InscripcionCreadaPayload{InscripcionID: "i1", CursoID: "c1", ParticipanteID: "p1", FechaInscripcion: at}, repositories.LineaTiempoInscripcion, "2026-10-16"},
		{events.CursoCompletadoParticipante, events.CursoCompletadoPayload{CursoID: "c1", ParticipanteID: "p1", CompletadoEn: at}, repositories.LineaTiempoCursoCompletado, "2026-10-16"},
		{events.AsistenciaGeneralRegistrada, events.AsistenciaGeneralPayload{AsistenciaID: "a", ParticipanteID: "p1", Fecha: "2026-10-15", Presente: true}, repositories.LineaTiempoAsistenciaGeneral, "2026-10-15"},
	}
	for _, tc := range cases {
		if code := push(t, c, tc.eventType, tc.payload); code != http.StatusNoContent {
			t.Fatalf("%s: status %d", tc.eventType, code)
		}
	}
	if len(linea.entries) != 3 {
		t.Fatalf("entradas: %d", len(linea.entries))
	}
	byType := map[string]repositories.EntradaLineaTiempo{}
	for _, e := range linea.entries {
		byType[e.Tipo] = e
	}
	for _, tc := range cases {
		e := byType[tc.tipo]
		if e.ParticipanteID != "p1" || e.Fecha != tc.fecha || e.EventType != tc.eventType {
			t.Errorf("%s: entrada %+v", tc.tipo, e)
		}
		if strings.ContainsAny(e.Descripcion, "0123456789%") {
			t.Errorf("la línea de tiempo no lleva números ni porcentajes: %q", e.Descripcion)
		}
	}
}

func TestEducacionAusenciaNoSeAnota(t *testing.T) {
	linea := &fakeLinea{}
	c := NewEducacion(events.NoVerifier{}, events.NewMemoryStore(10), linea, quiet())
	code := push(t, c, events.AsistenciaGeneralRegistrada, events.AsistenciaGeneralPayload{AsistenciaID: "a", ParticipanteID: "p1", Fecha: "2026-10-15"})
	if code != http.StatusNoContent || len(linea.entries) != 0 {
		t.Fatalf("una ausencia no va a la línea de tiempo: %d %v", code, linea.entries)
	}
}

func TestEducacionErrores(t *testing.T) {
	c := NewEducacion(events.NoVerifier{}, events.NewMemoryStore(10), &fakeLinea{}, quiet())
	if code := push(t, c, events.InscripcionCreada, events.InscripcionCreadaPayload{CursoID: "c1"}); code != http.StatusInternalServerError {
		t.Errorf("sin participante: %d", code)
	}
	for _, et := range []string{events.InscripcionCreada, events.CursoCompletadoParticipante, events.AsistenciaGeneralRegistrada} {
		if code := push(t, c, et, "texto"); code != http.StatusInternalServerError {
			t.Errorf("%s payload inválido: %d", et, code)
		}
	}
	broken := NewEducacion(events.NoVerifier{}, events.NewMemoryStore(10), &fakeLinea{err: errors.New("caído")}, quiet())
	if code := push(t, broken, events.InscripcionCreada, events.InscripcionCreadaPayload{ParticipanteID: "p"}); code != http.StatusInternalServerError {
		t.Errorf("error al guardar: %d", code)
	}
}

func TestUsuariosBajaDeFicha(t *testing.T) {
	fichas := &fakeFichas{}
	c := NewUsuarios(events.NoVerifier{}, events.NewMemoryStore(10), fichas, quiet())
	if code := push(t, c, events.ParticipanteEliminado, events.ParticipantePayload{ParticipanteID: "p1"}); code != http.StatusNoContent {
		t.Fatalf("status %d", code)
	}
	if len(fichas.calls) != 1 || fichas.calls[0] != "p1" {
		t.Fatalf("baja: %v", fichas.calls)
	}
	if code := push(t, c, events.ParticipanteEliminado, events.ParticipantePayload{}); code != http.StatusInternalServerError {
		t.Errorf("sin participante: %d", code)
	}
	if code := push(t, c, events.ParticipanteEliminado, "x"); code != http.StatusInternalServerError {
		t.Errorf("payload inválido: %d", code)
	}
	fichas.err = errors.New("caído")
	if code := push(t, c, events.ParticipanteEliminado, events.ParticipantePayload{ParticipanteID: "p2"}); code != http.StatusInternalServerError {
		t.Errorf("error del repo: %d", code)
	}
}

// El filtro de cada suscripción en infra/pubsub/setup.sh tiene que
// coincidir con los tipos que el consumidor procesa.
func TestFiltrosDeInfraCoincidenConLosConsumidores(t *testing.T) {
	filters, err := events.InfraFilters("../../../../infra/pubsub/setup.sh")
	if err != nil {
		t.Fatal(err)
	}
	if got := NewEducacion(events.NoVerifier{}, nil, nil, nil).EventTypes(); !events.SameTypes(got, filters["promocion-humana-educacion"]) {
		t.Errorf("promocion-humana-educacion: consumidor %v, filtro %v", got, filters["promocion-humana-educacion"])
	}
	if got := NewUsuarios(events.NoVerifier{}, nil, nil, nil).EventTypes(); !events.SameTypes(got, filters["promocion-humana-usuarios"]) {
		t.Errorf("promocion-humana-usuarios: consumidor %v, filtro %v", got, filters["promocion-humana-usuarios"])
	}
}
