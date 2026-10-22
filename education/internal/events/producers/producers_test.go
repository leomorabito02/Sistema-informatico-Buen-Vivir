package producers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/events"
	"buenvivir/education/internal/models"
)

func TestProductorNilNoPublica(t *testing.T) {
	var p *Educacion
	ctx := context.Background()
	assert.False(t, p.Enabled())
	p.Notify()
	course := &models.Course{ID: "c1"}
	assert.NoError(t, p.CursoCreado(ctx, course))
	assert.NoError(t, p.CursoActualizado(ctx, course))
	assert.NoError(t, p.CursoEliminado(ctx, "c1"))
	assert.NoError(t, p.ContenidoDocenteSubido(ctx, course))
	assert.NoError(t, p.ContenidoDocenteModerado(ctx, "c1", events.ModeracionAprobado))
	assert.NoError(t, p.MaterialDocenteSubido(ctx, nil, models.ContenidoDocente{ID: "m1"}))
	assert.NoError(t, p.MaterialDocenteModerado(ctx, nil, models.ContenidoDocente{ID: "m1"}))
	assert.NoError(t, p.InscripcionCreada(ctx, nil, models.Enrollment{}))
	assert.NoError(t, p.AsistenciaRegistrada(ctx, nil, "a", "c", models.Attendance{}))
	assert.NoError(t, p.CursoCompletado(ctx, nil, "c", "p", time.Now()))
	assert.NoError(t, p.HabilidadOtorgada(ctx, nil, models.ParticipantSkill{}))
	assert.NoError(t, p.AsistenciaGeneral(ctx, nil, "a", models.GeneralAttendance{}, true))
	assert.NoError(t, p.Contenido(ctx, "id", events.ContenidoPublicado, events.ContenidoPayload{}))

	// Con writer nil también está apagado.
	assert.False(t, NewEducacion(nil, "t", "c").Enabled())
}

// Los payloads son el contrato con otros servicios y no pueden llevar
// datos personales ni texto libre (Ley 25.326, RNF-06/08).
func TestPayloadsSinDatosPersonales(t *testing.T) {
	p := NewEducacion(nil, "educacion-eventos", "contenido-eventos")
	course := &models.Course{
		ID: "c1", Title: "Huerta comunitaria", ResponsibleTeacher: "doc-1", Mode: "Presencial",
		Description:  json.RawMessage(`{"texto":"descripción larga"}`),
		TrainingArea: &models.TrainingArea{ID: "area-1", Name: "Ambiente"},
	}
	material := models.ContenidoDocente{
		ID: "m1", DocenteID: "doc-1", Titulo: "Huerta en casa", Descripcion: "descripción",
		MotivoRechazo: "texto del motivo", ArchivoNombre: "huerta.pdf", EstadoModeracion: "rechazado",
	}
	assert.Equal(t, events.ContenidoDocenteModeradoPayload{ContenidoID: "m1", DocenteID: "doc-1", Tipo: "material", Resultado: "rechazado"}, materialModeradoPayload(material))
	payload := cursoPayload(course)
	assert.Equal(t, events.CursoPayload{CursoID: "c1", AreaFormativaID: "area-1", DocenteID: "doc-1", Modalidad: "Presencial"}, payload)
	assert.Equal(t, "", cursoPayload(&models.Course{ID: "c2"}).AreaFormativaID)

	ev := p.event(events.AsistenciaGeneralRegistrada, events.AsistenciaGeneralPayload{AsistenciaID: "x", ParticipanteID: "p1"}, "p1")
	assert.Equal(t, "educacion-eventos", ev.Topic)
	assert.Equal(t, "p1", ev.OrderingKey)

	for name, v := range map[string]any{
		"curso":              payload,
		"asistencia general": events.AsistenciaGeneralPayload{AsistenciaID: "x", ParticipanteID: "p1", Fecha: "2026-10-16", Presente: true},
		"inscripción":        events.InscripcionCreadaPayload{InscripcionID: "i", CursoID: "c", ParticipanteID: "p"},
		"material subido":    materialSubidoPayload(material),
		"material moderado":  materialModeradoPayload(material),
	} {
		raw, err := json.Marshal(v)
		require.NoError(t, err)
		for _, prohibido := range []string{"Huerta", "descripción", "observacion", "nombre", "email", "dni", "telefono", "texto"} {
			assert.NotContains(t, strings.ToLower(string(raw)), strings.ToLower(prohibido), "%s no debe llevar %q", name, prohibido)
		}
	}
}

func TestParseContentfulWebhook(t *testing.T) {
	publish := `{"sys":{"id":"entry1","type":"Entry","revision":4,"contentType":{"sys":{"id":"course"}},"environment":{"sys":{"id":"testing"}}}}`
	hook, ok, err := ParseContentfulWebhook("ContentManagement.Entry.publish", []byte(publish))
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, events.ContenidoPublicado, hook.EventType)
	assert.Equal(t, events.ContenidoPayload{EntryID: "entry1", ContentType: "course", Environment: "testing", Accion: "publish", Revision: 4}, hook.Payload)
	assert.True(t, events.IsUUID(hook.EventID))

	// Mismo webhook reintentado → mismo event_id; otra revisión → otro.
	again, _, _ := ParseContentfulWebhook("ContentManagement.Entry.publish", []byte(publish))
	assert.Equal(t, hook.EventID, again.EventID)
	other, _, _ := ParseContentfulWebhook("ContentManagement.Entry.publish", []byte(strings.Replace(publish, `"revision":4`, `"revision":5`, 1)))
	assert.NotEqual(t, hook.EventID, other.EventID)

	deleted := `{"sys":{"id":"entry1","type":"DeletedEntry","version":7,"deletedAt":"2026-10-16T10:00:00Z"}}`
	for _, topic := range []string{"ContentManagement.Entry.unpublish", "ContentManagement.Entry.delete"} {
		h, ok, err := ParseContentfulWebhook(topic, []byte(deleted))
		require.NoError(t, err)
		require.True(t, ok, topic)
		assert.Equal(t, events.ContenidoDespublicado, h.EventType)
		assert.Equal(t, 0, h.Payload.Revision)
	}
	unpub, _, _ := ParseContentfulWebhook("ContentManagement.Entry.unpublish", []byte(deleted))
	del, _, _ := ParseContentfulWebhook("ContentManagement.Entry.delete", []byte(deleted))
	assert.NotEqual(t, unpub.EventID, del.EventID, "acciones distintas son hechos distintos")

	for _, topic := range []string{"ContentManagement.Entry.save", "ContentManagement.Asset.publish", "otra.cosa", ""} {
		_, ok, err := ParseContentfulWebhook(topic, []byte(publish))
		assert.NoError(t, err)
		assert.False(t, ok, "%q se ignora", topic)
	}

	_, _, err = ParseContentfulWebhook("ContentManagement.Entry.publish", []byte(`{`))
	assert.ErrorIs(t, err, ErrWebhookInvalido)
	_, _, err = ParseContentfulWebhook("ContentManagement.Entry.publish", []byte(`{"sys":{}}`))
	assert.ErrorIs(t, err, ErrWebhookInvalido)
}
