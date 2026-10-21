package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/repositories"
)

var ahora = time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)

func nuevoFichaService() (*fichaService, *memFichas) {
	repo := newMemFichas()
	auth := authConNombres{
		participantes: map[string]models.ParticipanteInfo{
			"p1": {ParticipanteID: "p1", Nombre: "Ana", Apellido: "Pérez", FechaNacimiento: "1996-10-22"},
			"p2": {ParticipanteID: "p2", Nombre: "Bruno", Apellido: "Gómez"},
		},
		usuarios: map[string]string{"u1": "Laura Díaz"},
	}
	return &fichaService{repo: repo, cipher: cifradoFalso{}, auth: auth, now: func() time.Time { return ahora }}, repo
}

func TestCrearFicha_GuardaCifradoYCreador(t *testing.T) {
	svc, repo := nuevoFichaService()
	ficha, err := svc.Create(context.Background(), "u1", models.CreateFichaInput{
		ParticipanteID: "p1", SituacionActual: " Busca trabajo ", Avances: "Terminó la secundaria",
	})
	if err != nil {
		t.Fatal(err)
	}
	if ficha.SituacionActual != "Busca trabajo" || ficha.CreadoPor != "u1" || !ficha.UpdatedAt.Equal(ahora) {
		t.Errorf("ficha creada: %+v", ficha)
	}
	doc := repo.docs[ficha.ID]
	if doc.SituacionActual != "enc:Busca trabajo" || doc.Avances != "enc:Terminó la secundaria" {
		t.Errorf("el contenido confidencial se guarda cifrado: %+v", doc)
	}
}

func TestCrearFicha_Validaciones(t *testing.T) {
	svc, _ := nuevoFichaService()
	ctx := context.Background()
	casos := map[string]struct {
		input models.CreateFichaInput
		msg   string
	}{
		"sin situación":            {models.CreateFichaInput{ParticipanteID: "p1", Avances: "x"}, MsgCamposObligatorios},
		"sin avances":              {models.CreateFichaInput{ParticipanteID: "p1", SituacionActual: "x"}, MsgCamposObligatorios},
		"sin participante":         {models.CreateFichaInput{SituacionActual: "x", Avances: "y"}, MsgCamposObligatorios},
		"participante inválido":    {models.CreateFichaInput{ParticipanteID: "p 1", SituacionActual: "x", Avances: "y"}, "inválido"},
		"participante inexistente": {models.CreateFichaInput{ParticipanteID: "p9", SituacionActual: "x", Avances: "y"}, "no existe"},
		"texto largo":              {models.CreateFichaInput{ParticipanteID: "p1", SituacionActual: strings.Repeat("a", maxTextoFicha+1), Avances: "y"}, "hasta"},
	}
	for nombre, c := range casos {
		_, err := svc.Create(ctx, "u1", c.input)
		if !IsValidationError(err) || !strings.Contains(err.Error(), c.msg) {
			t.Errorf("%s: esperaba error de validación con %q, obtuve %v", nombre, c.msg, err)
		}
	}
}

func TestCrearFicha_UnaPorParticipante(t *testing.T) {
	svc, _ := nuevoFichaService()
	input := models.CreateFichaInput{ParticipanteID: "p1", SituacionActual: "x", Avances: "y"}
	if _, err := svc.Create(context.Background(), "u1", input); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Create(context.Background(), "u1", input); !errors.Is(err, repositories.ErrAlreadyExists) {
		t.Errorf("segunda ficha: esperaba ErrAlreadyExists, obtuve %v", err)
	}
}

func TestEditarFicha(t *testing.T) {
	svc, repo := nuevoFichaService()
	ctx := context.Background()
	creada, _ := svc.Create(ctx, "u1", models.CreateFichaInput{ParticipanteID: "p1", SituacionActual: "x", Avances: "y"})

	editada, err := svc.Update(ctx, creada.ID, models.UpdateFichaInput{SituacionActual: "nueva", Avances: "más"})
	if err != nil {
		t.Fatal(err)
	}
	if editada.CreadoPor != "u1" || editada.ParticipanteID != "p1" || editada.Avances != "más" {
		t.Errorf("la edición conserva creador y participante: %+v", editada)
	}
	if repo.docs[creada.ID].SituacionActual != "enc:nueva" {
		t.Errorf("la edición se guarda cifrada: %+v", repo.docs[creada.ID])
	}

	if _, err := svc.Update(ctx, creada.ID, models.UpdateFichaInput{SituacionActual: "nueva"}); !IsValidationError(err) {
		t.Errorf("sin avances: esperaba error de validación, obtuve %v", err)
	}
	if _, err := svc.Update(ctx, "no-existe", models.UpdateFichaInput{SituacionActual: "a", Avances: "b"}); !errors.Is(err, repositories.ErrNotFound) {
		t.Errorf("ficha inexistente: esperaba ErrNotFound, obtuve %v", err)
	}
	if _, err := svc.Update(ctx, "../x", models.UpdateFichaInput{SituacionActual: "a", Avances: "b"}); !errors.Is(err, repositories.ErrNotFound) {
		t.Errorf("id inválido: esperaba ErrNotFound, obtuve %v", err)
	}
}

func TestDetalleDeFicha(t *testing.T) {
	svc, _ := nuevoFichaService()
	ctx := context.Background()
	creada, _ := svc.Create(ctx, "u1", models.CreateFichaInput{ParticipanteID: "p1", SituacionActual: "x", Avances: "y"})

	d, err := svc.GetDetalleByParticipante(ctx, "p1")
	if err != nil {
		t.Fatal(err)
	}
	if d.ID != creada.ID || d.SituacionActual != "x" || d.CreadoPorNombre != "Laura Díaz" {
		t.Errorf("detalle descifrado con el nombre del creador: %+v", d)
	}
	// Cumple 30 el 22/10/2026: el 21/10 todavía tiene 29.
	if d.Participante == nil || d.Participante.Nombre != "Ana" || d.Participante.Edad == nil || *d.Participante.Edad != 29 {
		t.Errorf("datos de Auth y edad: %+v", d.Participante)
	}

	if _, err := svc.GetDetalle(ctx, creada.ID); err != nil {
		t.Errorf("detalle por id: %v", err)
	}
	if _, err := svc.GetDetalleByParticipante(ctx, "p2"); !errors.Is(err, repositories.ErrNotFound) {
		t.Errorf("participante sin ficha: esperaba ErrNotFound, obtuve %v", err)
	}
	if _, err := svc.GetByParticipante(ctx, "p/2"); !IsValidationError(err) {
		t.Errorf("participante inválido: esperaba error de validación, obtuve %v", err)
	}
	if _, err := svc.GetByID(ctx, "a b"); !errors.Is(err, repositories.ErrNotFound) {
		t.Errorf("id inválido: esperaba ErrNotFound, obtuve %v", err)
	}
}

func TestCalcularEdad(t *testing.T) {
	casos := map[string]*int{
		"1996-10-21": intPtr(30), // cumple hoy
		"1996-10-22": intPtr(29), // cumple mañana
		"1996-11-01": intPtr(29),
		"2000-01-15": intPtr(26),
		"2030-01-01": nil, // fecha futura
		"":           nil,
		"21/10/1996": nil,
	}
	for fecha, esperada := range casos {
		got := calcularEdad(fecha, ahora)
		if (got == nil) != (esperada == nil) || (got != nil && *got != *esperada) {
			t.Errorf("%q: esperaba %v, obtuve %v", fecha, deref(esperada), deref(got))
		}
	}
}

func TestSanitizeText(t *testing.T) {
	casos := map[string]string{
		"  hola  ":                       "hola",
		"<script>alert(1)</script>texto": "alert(1)texto",
		"línea 1\r\nlínea 2":             "línea 1\nlínea 2",
		"con\x00control\x07":             "concontrol",
		"tab\tqueda":                     "tab\tqueda",
		"a < b y c > d (no es etiqueta)": "a < b y c > d (no es etiqueta)",
	}
	for in, want := range casos {
		if got := sanitizeText(in); got != want {
			t.Errorf("sanitizeText(%q) = %q, esperaba %q", in, got, want)
		}
	}
}

func intPtr(n int) *int { return &n }

func deref(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}
