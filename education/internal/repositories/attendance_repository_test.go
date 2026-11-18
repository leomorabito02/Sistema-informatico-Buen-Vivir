package repositories

import (
	"testing"
	"time"

	"cloud.google.com/go/firestore"

	"buenvivir/education/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedAttendance escribe una asistencia directo en el emulador: el
// repository es de solo lectura (las asistencias las genera
// SubmissionRepository.CreateWithAttendance).
func seedAttendance(t *testing.T, client *firestore.Client, actividadID, participanteID string, presente bool) {
	t.Helper()
	_, err := client.Collection(attendancesCollection).Doc(SubmissionDocID(actividadID, participanteID)).Set(t.Context(), models.Attendance{
		ParticipanteID: participanteID,
		ActividadID:    actividadID,
		Presente:       presente,
		Fecha:          time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
	})
	require.NoError(t, err)
}

func TestAttendanceRepository_ListByActivity(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, attendancesCollection)
	repo := NewAttendanceRepository(client)

	seedAttendance(t, client, "act-1", "p1", true)
	seedAttendance(t, client, "act-1", "p2", false)
	seedAttendance(t, client, "act-2", "p1", true)

	got, err := repo.ListByActivity(t.Context(), "act-1")
	require.NoError(t, err)
	require.Len(t, got, 2)

	byParticipant := map[string]models.Attendance{}
	for _, a := range got {
		byParticipant[a.ParticipanteID] = a
		assert.Equal(t, "act-1", a.ActividadID)
		assert.Equal(t, "act-1_"+a.ParticipanteID, a.ID)
	}
	assert.True(t, byParticipant["p1"].Presente)
	assert.False(t, byParticipant["p2"].Presente)
}

func TestAttendanceRepository_ListByParticipant_FiltersByParticipant(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, attendancesCollection)
	repo := NewAttendanceRepository(client)

	seedAttendance(t, client, "act-1", "p1", true)
	seedAttendance(t, client, "act-2", "p1", true)
	seedAttendance(t, client, "act-3", "p2", true)

	got, err := repo.ListByParticipant(t.Context(), "p1")
	require.NoError(t, err)
	require.Len(t, got, 2)
	for _, a := range got {
		assert.Equal(t, "p1", a.ParticipanteID)
	}
}
