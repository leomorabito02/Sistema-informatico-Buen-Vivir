package repositories

import (
	"testing"
	"time"

	"buenvivir/education/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivityRepository_Create_PersistsAndReturnsID(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, activitiesCollection)
	repo := NewActivityRepository(client)

	fecha := time.Date(2026, 9, 1, 14, 0, 0, 0, time.UTC)
	created, err := repo.Create(t.Context(), "curso-1", models.ActivityInput{
		Titulo: "Práctica 1",
		Fecha:  fecha,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "curso-1", created.CursoID)
	assert.Equal(t, "Práctica 1", created.Titulo)
	assert.False(t, created.CreatedAt.IsZero(), "el repository setea created_at")

	got, err := repo.Get(t.Context(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, "Práctica 1", got.Titulo)
	assert.True(t, fecha.Equal(got.Fecha))
}

func TestActivityRepository_Get_NotFound(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, activitiesCollection)
	repo := NewActivityRepository(client)

	got, err := repo.Get(t.Context(), "no-existe")
	assert.Nil(t, got)
	assert.ErrorIs(t, err, ErrNotFound)
	assert.True(t, IsNotFound(err))
}

func TestActivityRepository_ListByCourse_FiltersByCourseID(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, activitiesCollection)
	repo := NewActivityRepository(client)

	_, err := repo.Create(t.Context(), "curso-1", models.ActivityInput{Titulo: "A1", Fecha: time.Now().UTC()})
	require.NoError(t, err)
	_, err = repo.Create(t.Context(), "curso-1", models.ActivityInput{Titulo: "A2", Fecha: time.Now().UTC()})
	require.NoError(t, err)
	_, err = repo.Create(t.Context(), "curso-2", models.ActivityInput{Titulo: "otro curso", Fecha: time.Now().UTC()})
	require.NoError(t, err)

	got, err := repo.ListByCourse(t.Context(), "curso-1")
	require.NoError(t, err)
	require.Len(t, got, 2)
	for _, a := range got {
		assert.Equal(t, "curso-1", a.CursoID)
		assert.NotEmpty(t, a.ID)
	}
}

func TestActivityRepository_ListByCourse_EmptyWhenNoMatch(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, activitiesCollection)
	repo := NewActivityRepository(client)

	got, err := repo.ListByCourse(t.Context(), "curso-sin-actividades")
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Len(t, got, 0)
}
