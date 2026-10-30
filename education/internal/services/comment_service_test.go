package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"buenvivir/education/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCommentRepo struct {
	created   []*models.Comentario
	listTipo  string
	listID    string
	deletedID string
	err       error
}

func (f *fakeCommentRepo) ListByEntidad(_ context.Context, tipo, id string) ([]models.Comentario, error) {
	f.listTipo, f.listID = tipo, id
	return []models.Comentario{}, f.err
}

func (f *fakeCommentRepo) Create(_ context.Context, c *models.Comentario) error {
	if f.err != nil {
		return f.err
	}
	c.ID = "com-1"
	f.created = append(f.created, c)
	return nil
}

func (f *fakeCommentRepo) Delete(_ context.Context, id string) error {
	f.deletedID = id
	return f.err
}

func TestSanitizarComentario(t *testing.T) {
	casos := map[string]string{
		"  Muy linda la campaña  ":                   "Muy linda la campaña",
		"<script>alert('x')</script>Hola":            "alert('x')Hola",
		`<img src=x onerror="alert(1)">Bien`:         "Bien",
		"Me <b>encantó</b>":                          "Me encantó",
		"2 < 3 y 5 > 4":                              "2  3 y 5  4",
		"Línea 1\nLínea 2\x00\x07":                   "Línea 1\nLínea 2",
		"<<script>script>alert(1)<</script>/script>": "scriptalert(1)/script",
	}
	for entrada, quiero := range casos {
		got := SanitizarComentario(entrada)
		assert.Equal(t, quiero, got, entrada)
		assert.NotContains(t, got, "<")
		assert.NotContains(t, got, ">")
	}
}

func TestCommentService_Create(t *testing.T) {
	repo := &fakeCommentRepo{}
	svc := &commentService{repo: repo, now: func() time.Time { return time.Date(2026, 10, 22, 15, 0, 0, 0, time.FixedZone("ART", -3*3600)) }}

	c, err := svc.Create(context.Background(), models.ComentarioInput{
		EntidadTipo: "Campaña", EntidadID: "camp-1", Nombre: " Ana ", Apellido: "Gómez", Texto: "Me <b>encantó</b>",
	}, "")
	require.NoError(t, err)
	assert.Equal(t, "campana", c.EntidadTipo, "campaña con ñ se guarda sin ñ")
	assert.Equal(t, "Ana", c.NombreAutor)
	assert.Equal(t, "Me encantó", c.Texto)
	assert.Equal(t, time.UTC, c.Fecha.Location())
	assert.Empty(t, c.UsuarioID, "se puede comentar sin cuenta")
}

func TestCommentService_Create_Validaciones(t *testing.T) {
	base := models.ComentarioInput{EntidadTipo: "curso", EntidadID: "c-1", Nombre: "Ana", Apellido: "Gómez", Texto: "Hola"}
	casos := map[string]func(*models.ComentarioInput){
		"tipo inválido":              func(in *models.ComentarioInput) { in.EntidadTipo = "oferta" },
		"id vacío":                   func(in *models.ComentarioInput) { in.EntidadID = "" },
		"id con caracteres raros":    func(in *models.ComentarioInput) { in.EntidadID = "../x" },
		"sin nombre":                 func(in *models.ComentarioInput) { in.Nombre = "  " },
		"sin apellido":               func(in *models.ComentarioInput) { in.Apellido = "<b></b>" },
		"nombre largo":               func(in *models.ComentarioInput) { in.Nombre = strings.Repeat("a", 51) },
		"texto vacío tras sanitizar": func(in *models.ComentarioInput) { in.Texto = "<script></script>" },
		"texto largo":                func(in *models.ComentarioInput) { in.Texto = strings.Repeat("a", 1001) },
	}
	for nombre, cambiar := range casos {
		t.Run(nombre, func(t *testing.T) {
			repo := &fakeCommentRepo{}
			in := base
			cambiar(&in)
			_, err := NewCommentService(repo).Create(context.Background(), in, "")
			assert.True(t, IsValidationError(err), "quiero error de validación, vino %v", err)
			assert.Empty(t, repo.created)
		})
	}
}

func TestCommentService_ListYDelete(t *testing.T) {
	repo := &fakeCommentRepo{}
	svc := NewCommentService(repo)

	_, err := svc.List(context.Background(), "campaña", "camp-1")
	require.NoError(t, err)
	assert.Equal(t, "campana", repo.listTipo)
	assert.Equal(t, "camp-1", repo.listID)

	_, err = svc.List(context.Background(), "", "camp-1")
	assert.True(t, IsValidationError(err))

	require.NoError(t, svc.Delete(context.Background(), "com-1"))
	assert.Equal(t, "com-1", repo.deletedID)
	assert.True(t, IsValidationError(svc.Delete(context.Background(), "")))

	repo.err = errors.New("firestore caído")
	_, err = svc.Create(context.Background(), models.ComentarioInput{EntidadTipo: "curso", EntidadID: "c", Nombre: "A", Apellido: "B", Texto: "x"}, "")
	assert.Error(t, err)
}
