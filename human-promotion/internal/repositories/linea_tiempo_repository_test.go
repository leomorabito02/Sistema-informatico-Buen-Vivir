package repositories

import (
	"context"
	"testing"
)

func TestLineaTiempo_ListByParticipante(t *testing.T) {
	repo := NewLineaTiempoRepository(newTestFirestore(t))
	ctx := context.Background()

	for id, e := range map[string]EntradaLineaTiempo{
		"ev-1": {ParticipanteID: "p1", Tipo: LineaTiempoInscripcion, Fecha: "2026-09-02", CursoID: "c1"},
		"ev-2": {ParticipanteID: "p1", Tipo: LineaTiempoAsistenciaGeneral, Fecha: "2026-09-15"},
		"ev-3": {ParticipanteID: "p2", Tipo: LineaTiempoCursoCompletado, Fecha: "2026-10-01", CursoID: "c1"},
	} {
		if err := repo.Add(ctx, id, e); err != nil {
			t.Fatal(err)
		}
	}
	// Reentrega del mismo evento: pisa la entrada, no la duplica.
	if err := repo.Add(ctx, "ev-1", EntradaLineaTiempo{ParticipanteID: "p1", Tipo: LineaTiempoInscripcion, Fecha: "2026-09-02", CursoID: "c1"}); err != nil {
		t.Fatal(err)
	}

	got, err := repo.ListByParticipante(ctx, "p1")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("esperaba 2 entradas de p1, obtuve %+v", got)
	}
	ids := map[string]EntradaLineaTiempo{}
	for _, e := range got {
		ids[e.ID] = e
	}
	if e, ok := ids["ev-1"]; !ok || e.CursoID != "c1" || e.Tipo != LineaTiempoInscripcion {
		t.Errorf("falta la inscripción con su id y curso: %+v", got)
	}
	if _, ok := ids["ev-2"]; !ok {
		t.Errorf("falta la asistencia: %+v", got)
	}
}

func TestLineaTiempo_SinEntradas(t *testing.T) {
	got, err := NewLineaTiempoRepository(newTestFirestore(t)).ListByParticipante(context.Background(), "nadie")
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("sin entradas debe devolver lista vacía (no nil): %v, %v", got, err)
	}
}
