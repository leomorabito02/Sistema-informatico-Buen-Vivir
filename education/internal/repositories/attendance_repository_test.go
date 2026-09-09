package repositories

import (
	"testing"

	"buenvivir/education/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAttendanceRepository_RegisterBatch_WritesOneDocPerParticipant(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, attendancesCollection)
	repo := NewAttendanceRepository(client)

	err := repo.RegisterBatch(t.Context(), "act-1", []models.AttendanceInput{
		{ParticipanteID: "p1", Presente: true},
		{ParticipanteID: "p2", Presente: false},
	})
	require.NoError(t, err)

	got, err := repo.ListByActivity(t.Context(), "act-1")
	require.NoError(t, err)
	require.Len(t, got, 2)

	byParticipant := map[string]models.Attendance{}
	for _, a := range got {
		byParticipant[a.ParticipanteID] = a
		assert.Equal(t, "act-1", a.ActividadID)
		assert.False(t, a.Fecha.IsZero())
	}
	assert.True(t, byParticipant["p1"].Presente)
	assert.False(t, byParticipant["p2"].Presente)
}

func TestAttendanceRepository_RegisterBatch_DeterministicID_UpdatesInsteadOfDuplicating(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, attendancesCollection)
	repo := NewAttendanceRepository(client)

	require.NoError(t, repo.RegisterBatch(t.Context(), "act-1", []models.AttendanceInput{
		{ParticipanteID: "p1", Presente: true},
	}))
	// El docente corrige: vuelve a registrar al mismo participante en la
	// misma actividad con otro valor de "presente".
	require.NoError(t, repo.RegisterBatch(t.Context(), "act-1", []models.AttendanceInput{
		{ParticipanteID: "p1", Presente: false},
	}))

	got, err := repo.ListByActivity(t.Context(), "act-1")
	require.NoError(t, err)
	require.Len(t, got, 1, "id determinístico {actividadId}_{participanteId}: se sobrescribe, no se duplica")
	assert.Equal(t, "act-1_p1", got[0].ID)
	assert.False(t, got[0].Presente)
}

func TestAttendanceRepository_RegisterBatch_EmptyInputIsNoop(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, attendancesCollection)
	repo := NewAttendanceRepository(client)

	err := repo.RegisterBatch(t.Context(), "act-1", nil)
	require.NoError(t, err)

	got, err := repo.ListByActivity(t.Context(), "act-1")
	require.NoError(t, err)
	assert.Len(t, got, 0)
}

func TestAttendanceRepository_ListByParticipant_FiltersByParticipant(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, attendancesCollection)
	repo := NewAttendanceRepository(client)

	require.NoError(t, repo.RegisterBatch(t.Context(), "act-1", []models.AttendanceInput{{ParticipanteID: "p1", Presente: true}}))
	require.NoError(t, repo.RegisterBatch(t.Context(), "act-2", []models.AttendanceInput{{ParticipanteID: "p1", Presente: true}}))
	require.NoError(t, repo.RegisterBatch(t.Context(), "act-3", []models.AttendanceInput{{ParticipanteID: "p2", Presente: true}}))

	got, err := repo.ListByParticipant(t.Context(), "p1")
	require.NoError(t, err)
	require.Len(t, got, 2)
	for _, a := range got {
		assert.Equal(t, "p1", a.ParticipanteID)
	}
}
