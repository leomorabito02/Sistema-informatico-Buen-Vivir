package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/repositories"
)

func ptr(b bool) *bool { return &b }

func nuevoHitoService(t *testing.T, auth authConNombres) (HitoService, *memHitos, string) {
	t.Helper()
	fichas := newMemFichas()
	fichaID, _ := fichas.Create(context.Background(), models.FichaDocument{ParticipanteID: "p1"})
	hitos := newMemHitos()
	return NewHitoService(hitos, fichas, cifradoFalso{}, auth), hitos, fichaID
}

var equipo = authConNombres{usuarios: map[string]string{"u1": "Laura Díaz", "u2": "Pedro Ruiz"}}

func TestCrearHito_ResponsablePorDefectoYCifrado(t *testing.T) {
	svc, repo, fichaID := nuevoHitoService(t, equipo)
	hito, err := svc.Create(context.Background(), "u1", fichaID, models.HitoInput{
		Fecha: "2026-09-10", Descripcion: "  <b>Empezó</b> el taller  ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if hito.ResponsableID != "u1" || hito.VisibleParticipante {
		t.Errorf("responsable por defecto = usuario logueado y no visible: %+v", hito)
	}
	if hito.Descripcion != "Empezó el taller" {
		t.Errorf("la descripción tiene que quedar sin HTML ni espacios: %q", hito.Descripcion)
	}
	if guardado := repo.docs[hito.ID].Descripcion; guardado != "enc:Empezó el taller" {
		t.Errorf("la descripción se guarda cifrada, se guardó %q", guardado)
	}
}

func TestCrearHito_Validaciones(t *testing.T) {
	svc, _, fichaID := nuevoHitoService(t, equipo)
	ctx := context.Background()
	casos := map[string]struct {
		ficha string
		input models.HitoInput
		msg   string
	}{
		"sin descripción":         {fichaID, models.HitoInput{Fecha: "2026-09-10"}, MsgCamposObligatorios},
		"solo HTML":               {fichaID, models.HitoInput{Fecha: "2026-09-10", Descripcion: "<p></p>"}, MsgCamposObligatorios},
		"sin fecha":               {fichaID, models.HitoInput{Descripcion: "x"}, MsgCamposObligatorios},
		"fecha mal escrita":       {fichaID, models.HitoInput{Fecha: "10/09/2026", Descripcion: "x"}, "AAAA-MM-DD"},
		"fecha fuera de rango":    {fichaID, models.HitoInput{Fecha: "1850-01-01", Descripcion: "x"}, "AAAA-MM-DD"},
		"descripción larga":       {fichaID, models.HitoInput{Fecha: "2026-09-10", Descripcion: strings.Repeat("a", maxDescripcionHito+1)}, "hasta"},
		"responsable inexistente": {fichaID, models.HitoInput{Fecha: "2026-09-10", Descripcion: "x", ResponsableID: "u9"}, "no existe"},
		"responsable inválido":    {fichaID, models.HitoInput{Fecha: "2026-09-10", Descripcion: "x", ResponsableID: "u/9"}, "inválido"},
	}
	for nombre, c := range casos {
		_, err := svc.Create(ctx, "u1", c.ficha, c.input)
		if !IsValidationError(err) || !strings.Contains(err.Error(), c.msg) {
			t.Errorf("%s: esperaba error de validación con %q, obtuve %v", nombre, c.msg, err)
		}
	}

	if _, err := svc.Create(ctx, "u1", "no-existe", models.HitoInput{Fecha: "2026-09-10", Descripcion: "x"}); !errors.Is(err, repositories.ErrNotFound) {
		t.Errorf("ficha inexistente: esperaba ErrNotFound, obtuve %v", err)
	}
}

func TestCrearHito_OtroResponsableVisible(t *testing.T) {
	svc, _, fichaID := nuevoHitoService(t, equipo)
	hito, err := svc.Create(context.Background(), "u1", fichaID, models.HitoInput{
		Fecha: "2026-09-10", Descripcion: "x", ResponsableID: "u2", VisibleParticipante: ptr(true),
	})
	if err != nil || hito.ResponsableID != "u2" || !hito.VisibleParticipante {
		t.Fatalf("esperaba responsable u2 y visible: %+v, %v", hito, err)
	}
}

func TestEditarHito(t *testing.T) {
	svc, repo, fichaID := nuevoHitoService(t, equipo)
	ctx := context.Background()
	creado, _ := svc.Create(ctx, "u1", fichaID, models.HitoInput{Fecha: "2026-09-10", Descripcion: "x", VisibleParticipante: ptr(true)})

	// Sin responsable ni visibilidad en el body: se conservan.
	editado, err := svc.Update(ctx, fichaID, creado.ID, models.HitoInput{Fecha: "2026-09-11", Descripcion: "y"})
	if err != nil {
		t.Fatal(err)
	}
	if editado.ResponsableID != "u1" || !editado.VisibleParticipante || editado.Fecha != "2026-09-11" {
		t.Errorf("edición parcial: %+v", editado)
	}
	if repo.docs[creado.ID].Descripcion != "enc:y" {
		t.Errorf("la edición se guarda cifrada: %q", repo.docs[creado.ID].Descripcion)
	}

	editado, err = svc.Update(ctx, fichaID, creado.ID, models.HitoInput{Fecha: "2026-09-11", Descripcion: "y", ResponsableID: "u2", VisibleParticipante: ptr(false)})
	if err != nil || editado.ResponsableID != "u2" || editado.VisibleParticipante {
		t.Errorf("cambio de responsable y visibilidad: %+v, %v", editado, err)
	}
}

func TestEditarHito_SoloDesdeSuFicha(t *testing.T) {
	svc, _, fichaID := nuevoHitoService(t, equipo)
	ctx := context.Background()
	creado, _ := svc.Create(ctx, "u1", fichaID, models.HitoInput{Fecha: "2026-09-10", Descripcion: "x"})

	input := models.HitoInput{Fecha: "2026-09-10", Descripcion: "y"}
	if _, err := svc.Update(ctx, "otra-ficha", creado.ID, input); !errors.Is(err, repositories.ErrNotFound) {
		t.Errorf("hito de otra ficha: esperaba ErrNotFound, obtuve %v", err)
	}
	if _, err := svc.Update(ctx, fichaID, "../h1", input); !errors.Is(err, repositories.ErrNotFound) {
		t.Errorf("id inválido: esperaba ErrNotFound, obtuve %v", err)
	}
	if _, err := svc.Update(ctx, fichaID, creado.ID, models.HitoInput{Fecha: "2026-09-10"}); !IsValidationError(err) {
		t.Errorf("sin descripción: esperaba error de validación, obtuve %v", err)
	}
	if _, err := svc.Update(ctx, fichaID, creado.ID, models.HitoInput{Fecha: "2026-09-10", Descripcion: "y", ResponsableID: "u9"}); !IsValidationError(err) {
		t.Errorf("responsable inexistente: esperaba error de validación, obtuve %v", err)
	}
}

func TestListarHitos_OrdenYNombres(t *testing.T) {
	svc, _, fichaID := nuevoHitoService(t, equipo)
	ctx := context.Background()
	for _, h := range []models.HitoInput{
		{Fecha: "2026-09-20", Descripcion: "tercero", ResponsableID: "u2"},
		{Fecha: "2026-09-01", Descripcion: "primero"},
		{Fecha: "2026-09-10", Descripcion: "segundo"},
	} {
		if _, err := svc.Create(ctx, "u1", fichaID, h); err != nil {
			t.Fatal(err)
		}
	}

	detalle, err := svc.ListDetalle(ctx, fichaID)
	if err != nil {
		t.Fatal(err)
	}
	var orden []string
	for _, h := range detalle {
		orden = append(orden, h.Descripcion)
	}
	if strings.Join(orden, ",") != "primero,segundo,tercero" {
		t.Errorf("orden cronológico: %v", orden)
	}
	if detalle[0].ResponsableNombre != "Laura Díaz" || detalle[2].ResponsableNombre != "Pedro Ruiz" {
		t.Errorf("nombres de responsables: %+v", detalle)
	}

	if _, err := svc.ListDetalle(ctx, "no-existe"); !errors.Is(err, repositories.ErrNotFound) {
		t.Errorf("ficha inexistente: esperaba ErrNotFound, obtuve %v", err)
	}
}

func TestListarHitos_AuthCaidoSinNombres(t *testing.T) {
	fichas := newMemFichas()
	fichaID, _ := fichas.Create(context.Background(), models.FichaDocument{ParticipanteID: "p1"})
	hitos := newMemHitos()
	_, _ = hitos.Create(context.Background(), models.HitoDocument{FichaID: fichaID, Fecha: "2026-09-01", Descripcion: "enc:x", ResponsableID: "u1"})

	svc := NewHitoService(hitos, fichas, cifradoFalso{}, authConNombres{err: errors.New("auth caído")})
	detalle, err := svc.ListDetalle(context.Background(), fichaID)
	if err != nil || len(detalle) != 1 || detalle[0].ResponsableNombre != "" {
		t.Fatalf("la línea sale igual, sin nombres: %+v, %v", detalle, err)
	}
}

func TestListarHitos_DescripcionCorrupta(t *testing.T) {
	fichas := newMemFichas()
	fichaID, _ := fichas.Create(context.Background(), models.FichaDocument{ParticipanteID: "p1"})
	hitos := newMemHitos()
	_, _ = hitos.Create(context.Background(), models.HitoDocument{FichaID: fichaID, Fecha: "2026-09-01", Descripcion: "texto sin cifrar"})

	svc := NewHitoService(hitos, fichas, cifradoFalso{}, equipo)
	if _, err := svc.ListByFicha(context.Background(), fichaID); err == nil {
		t.Fatal("un hito que no se puede descifrar tiene que devolver error, no texto basura")
	}
}
