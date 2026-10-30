package services

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"buenvivir/human-promotion/internal/models"
	"buenvivir/human-promotion/internal/repositories"
)

type lineaFalsa struct {
	entradas map[string][]repositories.EntradaLineaTiempo
	err      error
}

func (l lineaFalsa) Add(context.Context, string, repositories.EntradaLineaTiempo) error { return nil }
func (l lineaFalsa) ListByParticipante(_ context.Context, id string) ([]repositories.EntradaLineaTiempo, error) {
	return l.entradas[id], l.err
}

// fichasFalsas implementa solo lo que usan estos tests; el resto de la
// interfaz queda sin implementar (entra en pánico si se llama).
type fichasFalsas struct {
	repositories.FichaRepository
	porID           map[string]models.FichaDocument
	porParticipante map[string]string // participante -> ficha
}

func (f fichasFalsas) GetByID(_ context.Context, id string) (*models.FichaDocument, error) {
	doc, ok := f.porID[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return &doc, nil
}

func (f fichasFalsas) GetByParticipante(_ context.Context, participanteID string) (string, *models.FichaDocument, error) {
	id, ok := f.porParticipante[participanteID]
	if !ok {
		return "", nil, repositories.ErrNotFound
	}
	doc := f.porID[id]
	return id, &doc, nil
}

func TestAgruparActividad_AsistenciaPorMes(t *testing.T) {
	got := agruparActividad([]repositories.EntradaLineaTiempo{
		{ID: "a3", Tipo: repositories.LineaTiempoAsistenciaGeneral, Fecha: "2026-09-20"},
		{ID: "i1", Tipo: repositories.LineaTiempoInscripcion, Descripcion: "Se inscribió en un curso", Fecha: "2026-09-02", CursoID: "c1"},
		{ID: "a1", Tipo: repositories.LineaTiempoAsistenciaGeneral, Fecha: "2026-09-05"},
		{ID: "a2", Tipo: repositories.LineaTiempoAsistenciaGeneral, Fecha: "2026-10-01"},
		{ID: "c1", Tipo: repositories.LineaTiempoCursoCompletado, Descripcion: "Completó un curso", Fecha: "2026-09-30", CursoID: "c1"},
		{ID: "mal", Tipo: repositories.LineaTiempoInscripcion, Fecha: "30/09/2026"},
	})

	want := []models.EntradaActividad{
		{ID: "i1", Tipo: repositories.LineaTiempoInscripcion, Fecha: "2026-09-02", Descripcion: "Se inscribió en un curso", CursoID: "c1"},
		{ID: "asistencia-2026-09", Tipo: repositories.LineaTiempoAsistenciaGeneral, Fecha: "2026-09-05", Descripcion: "Asistió al centro en septiembre de 2026"},
		{ID: "c1", Tipo: repositories.LineaTiempoCursoCompletado, Fecha: "2026-09-30", Descripcion: "Completó un curso", CursoID: "c1"},
		{ID: "asistencia-2026-10", Tipo: repositories.LineaTiempoAsistenciaGeneral, Fecha: "2026-10-01", Descripcion: "Asistió al centro en octubre de 2026"},
	}
	if len(got) != len(want) {
		t.Fatalf("esperaba %d entradas, obtuve %+v", len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("entrada %d: esperaba %+v, obtuve %+v", i, want[i], got[i])
		}
	}
}

// Pedido del cliente: la línea de tiempo no muestra cantidades. El año
// es el único número permitido en la asistencia.
func TestAgruparActividad_SinContarDias(t *testing.T) {
	var entradas []repositories.EntradaLineaTiempo
	for _, d := range []string{"01", "02", "03", "04", "05"} {
		entradas = append(entradas, repositories.EntradaLineaTiempo{ID: d, Tipo: repositories.LineaTiempoAsistenciaGeneral, Fecha: "2026-08-" + d})
	}
	got := agruparActividad(entradas)
	if len(got) != 1 {
		t.Fatalf("cinco días del mismo mes tienen que ser una sola entrada: %+v", got)
	}
	if regexp.MustCompile(`\d+`).ReplaceAllString(got[0].Descripcion, "") != "Asistió al centro en agosto de " {
		t.Errorf("la descripción no debería tener números salvo el año: %q", got[0].Descripcion)
	}
}

func TestAgruparActividad_SinEntradas(t *testing.T) {
	if got := agruparActividad(nil); got == nil || len(got) != 0 {
		t.Fatalf("esperaba lista vacía (no nil), obtuve %v", got)
	}
}

func TestActividadDeFicha(t *testing.T) {
	svc := NewActividadService(
		lineaFalsa{entradas: map[string][]repositories.EntradaLineaTiempo{
			"p1": {{ID: "i1", Tipo: repositories.LineaTiempoInscripcion, Fecha: "2026-09-02", CursoID: "c1"}},
		}},
		fichasFalsas{porID: map[string]models.FichaDocument{"f1": {ParticipanteID: "p1"}}},
	)
	ctx := context.Background()

	got, err := svc.DeFicha(ctx, "f1")
	if err != nil || len(got) != 1 || got[0].CursoID != "c1" {
		t.Fatalf("actividad de la ficha f1: %+v, %v", got, err)
	}
	if _, err := svc.DeFicha(ctx, "no-existe"); !errors.Is(err, repositories.ErrNotFound) {
		t.Errorf("ficha inexistente: esperaba ErrNotFound, obtuve %v", err)
	}
	if _, err := svc.DeFicha(ctx, "../f1"); !IsValidationError(err) {
		t.Errorf("id inválido: esperaba error de validación, obtuve %v", err)
	}
}

func TestActividadDeParticipante_ErrorDelRepositorio(t *testing.T) {
	svc := NewActividadService(lineaFalsa{err: errors.New("firestore caído")}, fichasFalsas{})
	if _, err := svc.DeParticipante(context.Background(), "p1"); err == nil {
		t.Fatal("esperaba el error del repositorio")
	}
}
