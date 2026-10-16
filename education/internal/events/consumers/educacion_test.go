package consumers

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/events"
	"buenvivir/education/internal/models"
)

type fakeEnrollments struct {
	calls []string
	err   error
}

func (f *fakeEnrollments) DeactivateByParticipant(_ context.Context, participanteID, motivo string) (int, error) {
	f.calls = append(f.calls, participanteID+"|"+motivo)
	return 1, f.err
}

type fakeCache struct{ invalidations int }

func (f *fakeCache) InvalidateCache() { f.invalidations++ }

func quiet() *slog.Logger { return slog.New(slog.NewJSONHandler(io.Discard, nil)) }

func push(t *testing.T, h http.Handler, producer, eventType string, payload any) int {
	t.Helper()
	env, err := events.New(context.Background(), producer, eventType, payload)
	require.NoError(t, err)
	data, err := env.Marshal()
	require.NoError(t, err)
	body, err := events.BuildPushBody(data, env.Attributes(), "m", "s", 1)
	require.NoError(t, err)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/eventos", bytes.NewReader(body)))
	return rec.Code
}

func TestUsuariosParticipanteEliminadoDaDeBajaInscripciones(t *testing.T) {
	repo := &fakeEnrollments{}
	c := NewUsuarios(events.NoVerifier{}, events.NewMemoryStore(10), repo, quiet())

	code := push(t, c, events.ProducerAutenticacion, events.ParticipanteEliminado, events.ParticipantePayload{ParticipanteID: "part-1", UserID: "u1"})
	assert.Equal(t, http.StatusNoContent, code)
	assert.Equal(t, []string{"part-1|" + models.MotivoBajaParticipanteEliminado}, repo.calls)

	// Sin participante_id es un evento mal armado: se reintenta → DLQ.
	code = push(t, c, events.ProducerAutenticacion, events.ParticipanteEliminado, events.ParticipantePayload{UserID: "u1"})
	assert.Equal(t, http.StatusInternalServerError, code)

	repo.err = errors.New("firestore caído")
	code = push(t, c, events.ProducerAutenticacion, events.ParticipanteEliminado, events.ParticipantePayload{ParticipanteID: "part-2"})
	assert.Equal(t, http.StatusInternalServerError, code)
}

func TestUsuariosUsuarioActualizado(t *testing.T) {
	repo := &fakeEnrollments{}
	c := NewUsuarios(events.NoVerifier{}, events.NewMemoryStore(10), repo, quiet())

	// Cambios que no dan de baja: cuenta activa, sin participante o sin estado.
	for _, p := range []events.UsuarioActualizadoPayload{
		{UserID: "u1", ParticipanteID: "part-1", Status: "active"},
		{UserID: "u2", Status: "blocked"},
		{UserID: "u3", ParticipanteID: "part-3"},
	} {
		assert.Equal(t, http.StatusNoContent, push(t, c, events.ProducerAutenticacion, events.UsuarioActualizado, p))
	}
	assert.Empty(t, repo.calls)

	assert.Equal(t, http.StatusNoContent, push(t, c, events.ProducerAutenticacion, events.UsuarioActualizado,
		events.UsuarioActualizadoPayload{UserID: "u4", ParticipanteID: "part-4", Status: "blocked"}))
	assert.Equal(t, []string{"part-4|" + models.MotivoBajaParticipanteEliminado}, repo.calls)
}

func TestUsuariosPayloadInvalido(t *testing.T) {
	c := NewUsuarios(events.NoVerifier{}, events.NewMemoryStore(10), &fakeEnrollments{}, quiet())
	assert.Equal(t, http.StatusInternalServerError, push(t, c, events.ProducerAutenticacion, events.ParticipanteEliminado, "texto"))
	assert.Equal(t, http.StatusInternalServerError, push(t, c, events.ProducerAutenticacion, events.UsuarioActualizado, []int{1}))
}

func TestContenidoInvalidaCache(t *testing.T) {
	cache := &fakeCache{}
	c := NewContenido(events.NoVerifier{}, events.NewMemoryStore(10), cache, quiet())
	payload := events.ContenidoPayload{EntryID: "e1", Accion: "publish"}
	assert.Equal(t, http.StatusNoContent, push(t, c, events.ProducerContentful, events.ContenidoPublicado, payload))
	assert.Equal(t, http.StatusNoContent, push(t, c, events.ProducerContentful, events.ContenidoDespublicado, payload))
	assert.Equal(t, 2, cache.invalidations)
	assert.ElementsMatch(t, []string{events.ContenidoPublicado, events.ContenidoDespublicado}, c.EventTypes())
}

// El filtro de cada suscripción en infra/pubsub/setup.sh tiene que
// coincidir con los tipos que el consumidor procesa.
func TestFiltrosDeInfraCoincidenConLosConsumidores(t *testing.T) {
	filters, err := events.InfraFilters("../../../../infra/pubsub/setup.sh")
	require.NoError(t, err)
	assert.ElementsMatch(t, NewUsuarios(events.NoVerifier{}, nil, nil, nil).EventTypes(), filters["educacion-usuarios"])
	assert.ElementsMatch(t, NewContenido(events.NoVerifier{}, nil, nil, nil).EventTypes(), filters["educacion-contenido"])
}
