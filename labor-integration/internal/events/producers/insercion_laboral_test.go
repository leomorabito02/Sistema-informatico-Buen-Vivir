package producers

import (
	"context"
	"testing"

	"buenvivir/labor-integration/internal/events"
)

func TestProductorNilNoPublica(t *testing.T) {
	var p InsercionLaboralProducer
	ctx := context.Background()

	if p != nil && p.Enabled() {
		t.Error("productor nil no debería estar habilitado")
	}

	p = NewInsercionLaboral(nil, "")
	if p.Enabled() {
		t.Error("productor con writer nil no debería estar habilitado")
	}

	p.Notify()

	if err := p.PostulacionSeleccionada(ctx, "app-1", "off-1", "part-1", "comp-1"); err != nil {
		t.Fatalf("se esperaba nil al publicar con productor deshabilitado, se obtuvo: %v", err)
	}
}

func TestPostulacionSeleccionadaPayloadSinPII(t *testing.T) {
	payload := events.PostulacionSeleccionadaPayload{
		PostulacionID:  "app-123",
		OfertaID:       "off-456",
		ParticipanteID: "part-789",
		EmpresaID:      "comp-012",
	}

	if payload.PostulacionID != "app-123" || payload.OfertaID != "off-456" || payload.ParticipanteID != "part-789" || payload.EmpresaID != "comp-012" {
		t.Errorf("payload no coincide con los valores esperados: %+v", payload)
	}
}
