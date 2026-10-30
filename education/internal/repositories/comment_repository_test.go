package repositories

import (
	"testing"
	"time"

	"buenvivir/education/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentRepository_CreateListDelete(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, commentCollection)
	repo := NewCommentRepository(client)

	viejo := &models.Comentario{EntidadTipo: "campana", EntidadID: "camp-1", NombreAutor: "Ana", ApellidoAutor: "Gómez", Texto: "Primero", Fecha: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)}
	nuevo := &models.Comentario{EntidadTipo: "campana", EntidadID: "camp-1", NombreAutor: "Luis", ApellidoAutor: "Paz", Texto: "Segundo", Fecha: time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)}
	otro := &models.Comentario{EntidadTipo: "curso", EntidadID: "camp-1", NombreAutor: "Eva", ApellidoAutor: "Sosa", Texto: "De un curso con el mismo id", Fecha: time.Now()}
	for _, c := range []*models.Comentario{viejo, nuevo, otro} {
		require.NoError(t, repo.Create(t.Context(), c))
		require.NotEmpty(t, c.ID)
	}

	got, err := repo.ListByEntidad(t.Context(), "campana", "camp-1")
	require.NoError(t, err)
	require.Len(t, got, 2, "no mezcla un curso con una campaña del mismo id")
	assert.Equal(t, "Segundo", got[0].Texto, "del más nuevo al más viejo")
	assert.Equal(t, "Primero", got[1].Texto)

	require.NoError(t, repo.Delete(t.Context(), nuevo.ID))
	got, err = repo.ListByEntidad(t.Context(), "campana", "camp-1")
	require.NoError(t, err)
	assert.Len(t, got, 1)

	assert.ErrorIs(t, repo.Delete(t.Context(), nuevo.ID), ErrNotFound)

	vacio, err := repo.ListByEntidad(t.Context(), "curso", "sin-comentarios")
	require.NoError(t, err)
	assert.NotNil(t, vacio)
	assert.Empty(t, vacio)
}
