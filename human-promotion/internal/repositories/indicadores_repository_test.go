package repositories

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"buenvivir/human-promotion/internal/audit"
	"buenvivir/human-promotion/internal/models"
)

func TestIndicadores_FichasEHitosSinDatosConfidenciales(t *testing.T) {
	client := newTestFirestore(t)
	ctx := context.Background()
	fichas := NewFichaRepository(client)
	hitos := NewHitoRepository(client)

	// Ficha nueva: guarda creada_en al crearse.
	antes := time.Now().UTC().Add(-time.Second)
	idNueva, err := fichas.Create(ctx, models.FichaDocument{ParticipanteID: "p-nueva", SituacionActual: "SECRETO-1", Avances: "SECRETO-2", UpdatedAt: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := hitos.Create(ctx, models.HitoDocument{FichaID: idNueva, Fecha: "2026-09-10", Descripcion: "SECRETO-3", ResponsableID: "r1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := hitos.Create(ctx, models.HitoDocument{FichaID: idNueva, Fecha: "2026-10-02", Descripcion: "SECRETO-4", ResponsableID: "r1"}); err != nil {
		t.Fatal(err)
	}

	// Ficha vieja (sin creada_en) con su "crear" en la auditoría.
	altaVieja := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	vieja := client.Collection(fichasCollection).NewDoc()
	if _, err := vieja.Set(ctx, map[string]any{"participante_id": "p-vieja", "situacion_actual": "x", "avances": "y", "updated_at": time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Collection(audit.Collection).NewDoc().Set(ctx, audit.Entry{Accion: audit.AccionCrear, Recurso: audit.RecursoFicha, RecursoID: vieja.ID, Timestamp: altaVieja}); err != nil {
		t.Fatal(err)
	}

	// Ficha vieja sin auditoría: usa la última modificación.
	modif := time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC)
	sinAuditoria := client.Collection(fichasCollection).NewDoc()
	if _, err := sinAuditoria.Set(ctx, map[string]any{"participante_id": "p-sin", "updated_at": modif}); err != nil {
		t.Fatal(err)
	}

	// Ficha dada de baja.
	if _, err := fichas.Create(ctx, models.FichaDocument{ParticipanteID: "p-baja", UpdatedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	if ok, err := fichas.DeactivateByParticipante(ctx, "p-baja"); err != nil || !ok {
		t.Fatalf("baja: %v %v", ok, err)
	}

	// Hito huérfano (ficha inexistente): no suma.
	if _, err := client.Collection(hitosCollection).NewDoc().Set(ctx, map[string]any{"ficha_id": "no-existe", "fecha": "2026-09-01"}); err != nil {
		t.Fatal(err)
	}

	ind, err := NewIndicadoresRepository(client).Indicadores(ctx)
	if err != nil {
		t.Fatal(err)
	}
	porParticipante := map[string]models.IndicadorFicha{}
	for _, f := range ind.Fichas {
		porParticipante[f.ParticipanteID] = f
	}
	if len(porParticipante) != 4 {
		t.Fatalf("esperaba 4 fichas, obtuve %+v", ind.Fichas)
	}
	if f := porParticipante["p-nueva"]; f.Estado != "activa" || f.CreadaEn.Before(antes) {
		t.Errorf("ficha nueva: %+v", f)
	}
	if f := porParticipante["p-vieja"]; !f.CreadaEn.Equal(altaVieja) {
		t.Errorf("ficha vieja: esperaba el alta de la auditoría, obtuve %v", f.CreadaEn)
	}
	if f := porParticipante["p-sin"]; !f.CreadaEn.Equal(modif) {
		t.Errorf("ficha sin auditoría: esperaba la última modificación, obtuve %v", f.CreadaEn)
	}
	if f := porParticipante["p-baja"]; f.Estado != "baja" || f.FechaBaja == nil {
		t.Errorf("ficha de baja: %+v", f)
	}

	if len(ind.Hitos) != 2 || ind.Hitos[0].ParticipanteID != "p-nueva" || ind.Hitos[0].Fecha != "2026-09-10" {
		t.Errorf("hitos: %+v", ind.Hitos)
	}

	// Nada confidencial en la respuesta.
	b, _ := json.Marshal(ind)
	if strings.Contains(string(b), "SECRETO") {
		t.Fatalf("la respuesta incluye textos confidenciales: %s", b)
	}
}

func TestIndicadores_SinDatos(t *testing.T) {
	ind, err := NewIndicadoresRepository(newTestFirestore(t)).Indicadores(context.Background())
	if err != nil || ind.Fichas == nil || ind.Hitos == nil || len(ind.Fichas)+len(ind.Hitos) != 0 {
		t.Fatalf("sin datos debe devolver listas vacías (no null): %+v, %v", ind, err)
	}
}
