package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/audit"
	"buenvivir/education/internal/events"
	"buenvivir/education/internal/models"
)

func TestCursoBajaRepository(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, cursosBajaCollection)
	wipeCollection(t, client, audit.Collection)
	repo := NewCursoBajaRepository(client)
	ctx := audit.WithClientIP(events.WithActor(context.Background(), "admin-1"), "10.0.0.1:5000")

	dado, err := repo.EstaDadoDeBaja(ctx, "c1")
	require.NoError(t, err)
	assert.False(t, dado)

	require.NoError(t, repo.DarDeBaja(ctx, models.CursoBaja{CursoID: "c1", Titulo: "Huerta", FechaBaja: time.Now().UTC(), InscriptosActivos: 2}))
	assert.ErrorIs(t, repo.DarDeBaja(ctx, models.CursoBaja{CursoID: "c1"}), ErrCursoYaDadoDeBaja)

	dado, err = repo.EstaDadoDeBaja(ctx, "c1")
	require.NoError(t, err)
	assert.True(t, dado)
	ids, err := repo.IDs(ctx)
	require.NoError(t, err)
	assert.Equal(t, map[string]bool{"c1": true}, ids)

	// la baja dejó un único registro de auditoría (el segundo intento no)
	docs, err := client.Collection(audit.Collection).Documents(ctx).GetAll()
	require.NoError(t, err)
	require.Len(t, docs, 1)
	var e audit.Entry
	require.NoError(t, docs[0].DataTo(&e))
	assert.Equal(t, "admin-1", e.UsuarioID)
	assert.Equal(t, audit.AccionBaja, e.Accion)
	assert.Equal(t, audit.RecursoCurso, e.Recurso)
	assert.Equal(t, "c1", e.RecursoID)
	assert.Equal(t, "10.0.0.1", e.IP)

	// y se lee con el repositorio de auditoría
	list, err := NewAuditRepository(client).List(ctx, time.Now().Add(-time.Hour), time.Now().Add(time.Hour), 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "c1", list[0].RecursoID)

	list, err = NewAuditRepository(client).List(ctx, time.Now().Add(time.Hour), time.Now().Add(2*time.Hour), 10)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestGeneralAttendance_RegistraAuditoria(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, generalAttendanceCollection)
	wipeCollection(t, client, audit.Collection)
	repo := NewGeneralAttendanceRepository(client)
	ctx := events.WithActor(context.Background(), "vol-1")

	in := []models.GeneralAttendanceInput{{ParticipanteID: "p1", Presente: true}, {ParticipanteID: "p2", Presente: false}}
	require.NoError(t, repo.RegisterBatch(ctx, "2026-10-20", in))
	// misma carga: no es un cambio. Una edición: un registro más.
	require.NoError(t, repo.RegisterBatch(ctx, "2026-10-20", in))
	require.NoError(t, repo.RegisterBatch(ctx, "2026-10-20", []models.GeneralAttendanceInput{{ParticipanteID: "p2", Presente: true}}))

	docs, err := client.Collection(audit.Collection).Documents(ctx).GetAll()
	require.NoError(t, err)
	acciones := map[string]int{}
	for _, d := range docs {
		var e audit.Entry
		require.NoError(t, d.DataTo(&e))
		assert.Equal(t, audit.RecursoAsistenciaGeneral, e.Recurso)
		assert.Equal(t, "vol-1", e.UsuarioID)
		acciones[e.Accion+":"+e.RecursoID]++
	}
	assert.Equal(t, map[string]int{
		"registrar:2026-10-20_p1":  1,
		"registrar:2026-10-20_p2":  1,
		"actualizar:2026-10-20_p2": 1,
	}, acciones)
}
