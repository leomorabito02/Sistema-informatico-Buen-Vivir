package repositories

import (
	"context"
	"testing"
	"time"

	"buenvivir/human-promotion/internal/audit"
	"buenvivir/human-promotion/internal/events"
	"buenvivir/human-promotion/internal/models"
)

func TestAuditoria_ListaLoQueRegistranFichasEHitos(t *testing.T) {
	client := newTestFirestore(t)
	ctx := audit.WithClientIP(events.WithActor(context.Background(), "admin-aud"), "10.1.1.1")
	inicio := time.Now().UTC().Add(-time.Second)

	fichaID, err := NewFichaRepository(client).Create(ctx, models.FichaDocument{ParticipanteID: "p-aud", SituacionActual: "SECRETO", UpdatedAt: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}

	list, err := NewAuditoriaRepository(client).List(ctx, inicio, time.Now().UTC().Add(time.Minute), 50)
	if err != nil {
		t.Fatal(err)
	}
	var encontrado *models.RegistroAuditoria
	for i := range list {
		if list[i].RecursoID == fichaID {
			encontrado = &list[i]
		}
	}
	if encontrado == nil {
		t.Fatalf("no aparece el alta de la ficha %s en %+v", fichaID, list)
	}
	if encontrado.UsuarioID != "admin-aud" || encontrado.Accion != audit.AccionCrear || encontrado.Recurso != audit.RecursoFicha || encontrado.IP != "10.1.1.1" {
		t.Errorf("registro inesperado: %+v", *encontrado)
	}

	vacio, err := NewAuditoriaRepository(client).List(ctx, time.Now().Add(time.Hour), time.Now().Add(2*time.Hour), 50)
	if err != nil || len(vacio) != 0 {
		t.Errorf("rango futuro: %+v %v", vacio, err)
	}
}
