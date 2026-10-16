package repositories

import (
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/events"
	"buenvivir/education/internal/events/producers"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/outbox"
)

// Tests de los eventos que escriben los repositorios en el outbox, en
// la misma transacción que el cambio (contra el emulador de Firestore).

func newEventsTestEnv(t *testing.T) (*firestore.Client, *producers.Educacion) {
	t.Helper()
	client := newTestFirestore(t)
	for _, c := range []string{outbox.Collection, activitiesCollection, attendancesCollection, submissionsCollection,
		enrollmentsCollection, generalAttendanceCollection, skillsCollection, activitySkillsCollection, participantSkillsCollection} {
		wipeCollection(t, client, c)
	}
	writer := outbox.NewWriter(client, events.ProducerEducacion, nil)
	return client, producers.NewEducacion(writer, "educacion-eventos", "contenido-eventos")
}

// outboxEvents devuelve los envelopes pendientes, por tipo.
func outboxEvents(t *testing.T, client *firestore.Client) map[string][]outbox.Record {
	t.Helper()
	docs, err := client.Collection(outbox.Collection).Documents(t.Context()).GetAll()
	require.NoError(t, err)
	out := map[string][]outbox.Record{}
	for _, d := range docs {
		var r outbox.Record
		require.NoError(t, d.DataTo(&r))
		out[r.EventType] = append(out[r.EventType], r)
	}
	return out
}

func decode[T any](t *testing.T, r outbox.Record) T {
	t.Helper()
	env, err := events.Decode([]byte(r.Data))
	require.NoError(t, err)
	p, err := events.DecodePayload[T](env)
	require.NoError(t, err)
	return p
}

func TestEnrollmentCreate_PublicaInscripcionCreada(t *testing.T) {
	client, p := newEventsTestEnv(t)
	repo := NewEnrollmentRepository(client, p)

	e, err := repo.Create(t.Context(), models.Enrollment{ParticipanteID: "part-1", CursoID: "c1"})
	require.NoError(t, err)

	evs := outboxEvents(t, client)[events.InscripcionCreada]
	require.Len(t, evs, 1)
	assert.Equal(t, "educacion-eventos", evs[0].Topic)
	assert.Equal(t, "part-1", evs[0].OrderingKey, "ordering key = participante")
	payload := decode[events.InscripcionCreadaPayload](t, evs[0])
	assert.Equal(t, e.ID, payload.InscripcionID)
	assert.Equal(t, "c1", payload.CursoID)
}

func TestEnrollmentDeactivateByParticipant(t *testing.T) {
	client, _ := newEventsTestEnv(t)
	repo := NewEnrollmentRepository(client)
	ctx := t.Context()
	for _, e := range []models.Enrollment{
		{ParticipanteID: "part-1", CursoID: "c1"},
		{ParticipanteID: "part-1", CursoID: "c2", Estado: models.EnrollmentEstadoFinalizada},
		{ParticipanteID: "part-2", CursoID: "c1"},
	} {
		_, err := repo.Create(ctx, e)
		require.NoError(t, err)
	}
	assert.Empty(t, outboxEvents(t, client), "sin productor no se publica nada")

	n, err := repo.DeactivateByParticipant(ctx, "part-1", models.MotivoBajaParticipanteEliminado)
	require.NoError(t, err)
	assert.Equal(t, 1, n, "solo las activas")
	n, err = repo.DeactivateByParticipant(ctx, "part-1", models.MotivoBajaParticipanteEliminado)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "repetirlo no cambia nada")

	list, err := repo.ListByParticipant(ctx, "part-1")
	require.NoError(t, err)
	estados := map[string]models.Enrollment{}
	for _, e := range list {
		estados[e.CursoID] = e
	}
	assert.Equal(t, models.EnrollmentEstadoBaja, estados["c1"].Estado)
	assert.Equal(t, models.MotivoBajaParticipanteEliminado, estados["c1"].MotivoBaja)
	assert.NotNil(t, estados["c1"].FechaBaja)
	assert.Equal(t, models.EnrollmentEstadoFinalizada, estados["c2"].Estado, "el historial se conserva")

	other, err := repo.ListByParticipant(ctx, "part-2")
	require.NoError(t, err)
	assert.Equal(t, models.EnrollmentEstadoActiva, other[0].Estado)
}

func TestSubmission_PublicaAsistenciaHabilidadYCursoCompletado(t *testing.T) {
	client, p := newEventsTestEnv(t)
	ctx := t.Context()
	activities := NewActivityRepository(client)
	skills := NewSkillRepository(client, p)
	subs := NewSubmissionRepository(client, p)

	a1, err := activities.Create(ctx, "curso-1", models.ActivityInput{Titulo: "A1", Consigna: "x", Fecha: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)})
	require.NoError(t, err)
	a2, err := activities.Create(ctx, "curso-1", models.ActivityInput{Titulo: "A2", Consigna: "y", Fecha: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)})
	require.NoError(t, err)
	_, err = skills.SetActivitySkills(ctx, a1, []models.Skill{{ID: "compost", Nombre: "Compost", Categoria: models.SkillCategoryNaturaleza}})
	require.NoError(t, err)

	_, err = subs.CreateWithAttendance(ctx, a1, models.Submission{ParticipanteID: "p1", Texto: "texto privado", FechaEntrega: time.Now().UTC()})
	require.NoError(t, err)
	evs := outboxEvents(t, client)
	require.Len(t, evs[events.AsistenciaRegistrada], 1)
	require.Len(t, evs[events.HabilidadOtorgada], 1)
	assert.Empty(t, evs[events.CursoCompletadoParticipante], "falta una actividad")
	assert.NotContains(t, evs[events.AsistenciaRegistrada][0].Data, "texto privado", "el texto de la entrega no viaja")
	asis := decode[events.AsistenciaRegistradaPayload](t, evs[events.AsistenciaRegistrada][0])
	assert.Equal(t, events.AsistenciaRegistradaPayload{
		AsistenciaID: a1.ID + "_p1", ActividadID: a1.ID, CursoID: "curso-1", ParticipanteID: "p1", Fecha: a1.Fecha, Presente: true,
	}, asis)
	hab := decode[events.HabilidadOtorgadaPayload](t, evs[events.HabilidadOtorgada][0])
	assert.Equal(t, "compost", hab.HabilidadID)

	_, err = subs.CreateWithAttendance(ctx, a2, models.Submission{ParticipanteID: "p1", FechaEntrega: time.Now().UTC()})
	require.NoError(t, err)
	done := outboxEvents(t, client)[events.CursoCompletadoParticipante]
	require.Len(t, done, 1, "con la última entrega completa el curso")
	assert.Equal(t, "p1", decode[events.CursoCompletadoPayload](t, done[0]).ParticipanteID)

	// Un curso de una sola actividad se completa con esa entrega.
	solo, err := activities.Create(ctx, "curso-2", models.ActivityInput{Titulo: "Única", Consigna: "z", Fecha: time.Now().UTC()})
	require.NoError(t, err)
	_, err = subs.CreateWithAttendance(ctx, solo, models.Submission{ParticipanteID: "p2", FechaEntrega: time.Now().UTC()})
	require.NoError(t, err)
	assert.Len(t, outboxEvents(t, client)[events.CursoCompletadoParticipante], 2)
}

func TestGrantToParticipants_PublicaSoloLasNuevas(t *testing.T) {
	client, p := newEventsTestEnv(t)
	skills := NewSkillRepository(client, p)
	grant := models.SkillGrant{Origen: models.SkillOrigenActividad, Fecha: time.Now().UTC(), ActividadID: "a1", CursoID: "c1"}

	require.NoError(t, skills.GrantToParticipants(t.Context(), []string{"p1", "p2"}, []string{"h1"}, grant))
	require.NoError(t, skills.GrantToParticipants(t.Context(), []string{"p1", "p3"}, []string{"h1"}, grant))

	evs := outboxEvents(t, client)[events.HabilidadOtorgada]
	assert.Len(t, evs, 3, "p1 ya la tenía: no se vuelve a otorgar")
}

func TestGeneralAttendance_RegistradaYActualizada(t *testing.T) {
	client, p := newEventsTestEnv(t)
	repo := NewGeneralAttendanceRepository(client, p)
	ctx := t.Context()

	require.NoError(t, repo.RegisterBatch(ctx, "2026-10-16", []models.GeneralAttendanceInput{
		{ParticipanteID: "p1", Presente: true, Observacion: "llegó tarde por el colectivo"},
		{ParticipanteID: "p2", Presente: false},
	}))
	evs := outboxEvents(t, client)
	require.Len(t, evs[events.AsistenciaGeneralRegistrada], 2)
	for _, r := range evs[events.AsistenciaGeneralRegistrada] {
		assert.NotContains(t, r.Data, "colectivo", "la observación no viaja")
	}

	// Mismo dato otra vez: no es un cambio. p2 pasa a presente: actualizada.
	require.NoError(t, repo.RegisterBatch(ctx, "2026-10-16", []models.GeneralAttendanceInput{
		{ParticipanteID: "p1", Presente: true, Observacion: "llegó tarde por el colectivo"},
		{ParticipanteID: "p2", Presente: true},
	}))
	evs = outboxEvents(t, client)
	assert.Len(t, evs[events.AsistenciaGeneralRegistrada], 2)
	require.Len(t, evs[events.AsistenciaGeneralActualizada], 1)
	upd := decode[events.AsistenciaGeneralPayload](t, evs[events.AsistenciaGeneralActualizada][0])
	assert.Equal(t, events.AsistenciaGeneralPayload{AsistenciaID: "2026-10-16_p2", ParticipanteID: "p2", Fecha: "2026-10-16", Presente: true}, upd)

	list, err := repo.ListByDate(ctx, "2026-10-16")
	require.NoError(t, err)
	assert.Len(t, list, 2)
}
