package repositories

import (
	"testing"
	"time"

	"buenvivir/education/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeacherContentRepository_CicloCompleto(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, teacherContentCollection)
	wipeCollection(t, client, libraryDownloadsCol)
	repo := NewTeacherContentRepository(client)

	nuevo := func(docente, titulo string, dia int) *models.ContenidoDocente {
		c := &models.ContenidoDocente{
			ID: repo.NewID(), DocenteID: docente, Titulo: titulo,
			ArchivoNombre: "a.pdf", ArchivoTipo: "application/pdf", ArchivoRuta: "contenidos/x/a.pdf",
			EstadoModeracion: models.ModeracionPendiente,
			FechaSubida:      time.Date(2026, 10, dia, 0, 0, 0, 0, time.UTC),
		}
		require.NoError(t, repo.Create(t.Context(), c))
		return c
	}
	viejo := nuevo("doc-1", "Viejo", 1)
	reciente := nuevo("doc-1", "Reciente", 2)
	nuevo("doc-2", "De otro", 3)

	mios, err := repo.ListByDocente(t.Context(), "doc-1")
	require.NoError(t, err)
	require.Len(t, mios, 2)
	assert.Equal(t, "Reciente", mios[0].Titulo, "del más nuevo al más viejo")

	todos, err := repo.ListByEstado(t.Context(), "")
	require.NoError(t, err)
	assert.Len(t, todos, 3)

	viejo.Titulo = "Viejo editado"
	require.NoError(t, repo.Update(t.Context(), viejo))
	got, err := repo.Get(t.Context(), viejo.ID)
	require.NoError(t, err)
	assert.Equal(t, "Viejo editado", got.Titulo)
	assert.Equal(t, "contenidos/x/a.pdf", got.ArchivoRuta)

	mod, err := repo.Moderate(t.Context(), reciente.ID, models.ModeracionAprobado, "", "usr-admin")
	require.NoError(t, err)
	assert.Equal(t, models.ModeracionAprobado, mod.EstadoModeracion)
	require.NotNil(t, mod.FechaModeracion)
	aprobados, err := repo.ListByEstado(t.Context(), models.ModeracionAprobado)
	require.NoError(t, err)
	require.Len(t, aprobados, 1)
	assert.Equal(t, "usr-admin", aprobados[0].ModeradoPor)

	_, err = repo.Moderate(t.Context(), "no-existe", models.ModeracionAprobado, "", "usr-admin")
	assert.ErrorIs(t, err, ErrNotFound)
	_, err = repo.Get(t.Context(), "no-existe")
	assert.ErrorIs(t, err, ErrNotFound)
	assert.ErrorIs(t, repo.Update(t.Context(), &models.ContenidoDocente{ID: "no-existe"}), ErrNotFound)

	require.NoError(t, repo.Delete(t.Context(), viejo.ID))
	_, err = repo.Get(t.Context(), viejo.ID)
	assert.ErrorIs(t, err, ErrNotFound)

	for i, dia := range []int{5, 6} {
		require.NoError(t, repo.CreateDescarga(t.Context(), &models.DescargaBiblioteca{
			ContenidoID: reciente.ID, Nombre: []string{"Ana", "Luis"}[i], Apellido: "Paz",
			Email: "a@gmail.com", Telefono: "3547123456", Fecha: time.Date(2026, 10, dia, 0, 0, 0, 0, time.UTC),
		}))
	}
	descargas, err := repo.ListDescargas(t.Context())
	require.NoError(t, err)
	require.Len(t, descargas, 2)
	assert.Equal(t, "Luis", descargas[0].Nombre)
	assert.NotEmpty(t, descargas[0].ID)
}
