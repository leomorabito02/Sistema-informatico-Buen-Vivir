package repositories

import (
	"testing"
	"time"

	"buenvivir/education/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSubmissionTestEnv(t *testing.T) (SubmissionRepository, ActivityRepository, AttendanceRepository, *models.Activity) {
	t.Helper()
	client := newTestFirestore(t)
	for _, c := range []string{activitiesCollection, attendancesCollection, submissionsCollection, enrollmentsCollection} {
		wipeCollection(t, client, c)
	}
	activities := NewActivityRepository(client)
	activity, err := activities.Create(t.Context(), "curso-1", models.ActivityInput{
		Titulo:   "Práctica",
		Consigna: "Traer una receta",
		Fecha:    time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
	return NewSubmissionRepository(client), activities, NewAttendanceRepository(client), activity
}

func TestSubmissionRepository_CreateWithAttendance_WritesBoth(t *testing.T) {
	subs, _, attendances, activity := newSubmissionTestEnv(t)

	entregada := time.Date(2026, 10, 9, 18, 30, 0, 0, time.UTC)
	sub, err := subs.CreateWithAttendance(t.Context(), activity, models.Submission{
		ParticipanteID: "p1",
		Texto:          "mi respuesta",
		Archivos:       []models.SubmissionFile{{Nombre: "tp.pdf", Tipo: "application/pdf", Tamano: 3, Ruta: "entregas/x"}},
		FechaEntrega:   entregada,
	})
	require.NoError(t, err)
	assert.Equal(t, activity.ID+"_p1", sub.ID)

	got, err := subs.Get(t.Context(), sub.ID)
	require.NoError(t, err)
	assert.Equal(t, "mi respuesta", got.Texto)
	assert.Equal(t, activity.ID, got.ActividadID)
	require.Len(t, got.Archivos, 1)
	assert.Equal(t, "entregas/x", got.Archivos[0].Ruta)
	assert.True(t, got.FechaEntrega.Equal(entregada))

	att, err := attendances.ListByActivity(t.Context(), activity.ID)
	require.NoError(t, err)
	require.Len(t, att, 1)
	assert.Equal(t, sub.ID, att[0].ID, "Entrega y Asistencia comparten id")
	assert.True(t, att[0].Presente)
	assert.True(t, att[0].Fecha.Equal(activity.Fecha), "la asistencia lleva la fecha de la actividad, no la de la entrega")
}

func TestSubmissionRepository_CreateWithAttendance_IsUniqueAndImmutable(t *testing.T) {
	subs, _, _, activity := newSubmissionTestEnv(t)

	_, err := subs.CreateWithAttendance(t.Context(), activity, models.Submission{ParticipanteID: "p1", Texto: "primera"})
	require.NoError(t, err)

	_, err = subs.CreateWithAttendance(t.Context(), activity, models.Submission{ParticipanteID: "p1", Texto: "segunda"})
	assert.ErrorIs(t, err, ErrAlreadySubmitted)

	got, err := subs.Get(t.Context(), activity.ID+"_p1")
	require.NoError(t, err)
	assert.Equal(t, "primera", got.Texto, "la segunda entrega no pisa la primera")
}

func TestSubmissionRepository_GetNotFound(t *testing.T) {
	subs, _, _, _ := newSubmissionTestEnv(t)
	_, err := subs.Get(t.Context(), "no-existe")
	assert.True(t, IsNotFound(err))
}

func TestSubmissionRepository_ListByActivityAndGetForParticipant(t *testing.T) {
	subs, activities, _, activity := newSubmissionTestEnv(t)
	other, err := activities.Create(t.Context(), "curso-1", models.ActivityInput{Titulo: "Otra", Consigna: "c", Fecha: time.Now().UTC()})
	require.NoError(t, err)

	for _, p := range []string{"p1", "p2"} {
		_, err := subs.CreateWithAttendance(t.Context(), activity, models.Submission{ParticipanteID: p})
		require.NoError(t, err)
	}

	list, err := subs.ListByActivity(t.Context(), activity.ID)
	require.NoError(t, err)
	assert.Len(t, list, 2)
	for _, s := range list {
		assert.NotNil(t, s.Archivos, "nunca null en el JSON")
	}

	mine, err := subs.GetForParticipant(t.Context(), "p1", []string{activity.ID, other.ID})
	require.NoError(t, err)
	assert.Len(t, mine, 1)
	assert.NotNil(t, mine[activity.ID])
	assert.Nil(t, mine[other.ID])

	empty, err := subs.GetForParticipant(t.Context(), "p1", nil)
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestActivityRepository_Update_MovesAttendanceDate(t *testing.T) {
	subs, activities, attendances, activity := newSubmissionTestEnv(t)
	_, err := subs.CreateWithAttendance(t.Context(), activity, models.Submission{ParticipanteID: "p1"})
	require.NoError(t, err)

	nueva := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	updated, err := activities.Update(t.Context(), activity.ID, models.ActivityInput{Titulo: "Práctica 2", Consigna: "Nueva consigna", Fecha: nueva})
	require.NoError(t, err)
	assert.Equal(t, "Práctica 2", updated.Titulo)
	assert.Equal(t, "Nueva consigna", updated.Consigna)
	assert.Equal(t, "curso-1", updated.CursoID)
	assert.False(t, updated.CreatedAt.IsZero())

	att, err := attendances.ListByActivity(t.Context(), activity.ID)
	require.NoError(t, err)
	require.Len(t, att, 1)
	assert.True(t, att[0].Fecha.Equal(nueva), "la asistencia se mueve con la actividad")
}

func TestActivityRepository_Update_NotFound(t *testing.T) {
	_, activities, _, _ := newSubmissionTestEnv(t)
	_, err := activities.Update(t.Context(), "no-existe", models.ActivityInput{Titulo: "t", Consigna: "c", Fecha: time.Now()})
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestActivityRepository_Delete_CascadesSubmissionsAndAttendance(t *testing.T) {
	subs, activities, attendances, activity := newSubmissionTestEnv(t)
	other, err := activities.Create(t.Context(), "curso-1", models.ActivityInput{Titulo: "Otra", Consigna: "c", Fecha: time.Now().UTC()})
	require.NoError(t, err)
	for _, a := range []*models.Activity{activity, other} {
		_, err := subs.CreateWithAttendance(t.Context(), a, models.Submission{ParticipanteID: "p1"})
		require.NoError(t, err)
	}

	require.NoError(t, activities.Delete(t.Context(), activity.ID))

	_, err = activities.Get(t.Context(), activity.ID)
	assert.True(t, IsNotFound(err))
	list, err := subs.ListByActivity(t.Context(), activity.ID)
	require.NoError(t, err)
	assert.Empty(t, list)
	att, err := attendances.ListByActivity(t.Context(), activity.ID)
	require.NoError(t, err)
	assert.Empty(t, att)

	// La otra actividad queda intacta.
	att, err = attendances.ListByActivity(t.Context(), other.ID)
	require.NoError(t, err)
	assert.Len(t, att, 1)
}

func TestActivityRepository_Delete_NotFound(t *testing.T) {
	_, activities, _, _ := newSubmissionTestEnv(t)
	assert.ErrorIs(t, activities.Delete(t.Context(), "no-existe"), ErrNotFound)
}

func TestEnrollmentRepository_ListByParticipant(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, enrollmentsCollection)
	repo := NewEnrollmentRepository(client)

	for _, e := range []models.Enrollment{
		{ParticipanteID: "p1", CursoID: "c1"},
		{ParticipanteID: "p1", CursoID: "c2", Estado: models.EnrollmentEstadoBaja},
		{ParticipanteID: "p2", CursoID: "c1"},
	} {
		_, err := repo.Create(t.Context(), e)
		require.NoError(t, err)
	}

	got, err := repo.ListByParticipant(t.Context(), "p1")
	require.NoError(t, err)
	require.Len(t, got, 2)
	for _, e := range got {
		assert.Equal(t, "p1", e.ParticipanteID)
	}
}
