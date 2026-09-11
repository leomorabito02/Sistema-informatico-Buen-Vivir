package repositories

import (
	"testing"
	"time"

	"buenvivir/education/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnrollmentRepository_ListByCourse_FiltersByCourseID(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, enrollmentsCollection)
	repo := NewEnrollmentRepository(client)

	ctx := t.Context()
	now := time.Now().UTC()
	// inscripciones es de solo lectura desde este servicio: se siembra
	// directo con el cliente, no hay repo.Create.
	_, _, err := client.Collection(enrollmentsCollection).Add(ctx, models.Enrollment{
		ParticipanteID: "p1", CursoID: "curso-1", FechaInscripcion: now, Estado: models.EnrollmentEstadoActiva,
	})
	require.NoError(t, err)
	_, _, err = client.Collection(enrollmentsCollection).Add(ctx, models.Enrollment{
		ParticipanteID: "p2", CursoID: "curso-1", FechaInscripcion: now, Estado: models.EnrollmentEstadoFinalizada,
	})
	require.NoError(t, err)
	_, _, err = client.Collection(enrollmentsCollection).Add(ctx, models.Enrollment{
		ParticipanteID: "p3", CursoID: "curso-2", FechaInscripcion: now, Estado: models.EnrollmentEstadoActiva,
	})
	require.NoError(t, err)

	got, err := repo.ListByCourse(ctx, "curso-1")
	require.NoError(t, err)
	require.Len(t, got, 2)
	for _, e := range got {
		assert.Equal(t, "curso-1", e.CursoID)
		assert.NotEmpty(t, e.ID)
		assert.NotEmpty(t, e.Estado)
	}
}

func TestEnrollmentRepository_ListByCourse_EmptyWhenNoMatch(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, enrollmentsCollection)
	repo := NewEnrollmentRepository(client)

	got, err := repo.ListByCourse(t.Context(), "curso-sin-inscriptos")
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Len(t, got, 0)
}
