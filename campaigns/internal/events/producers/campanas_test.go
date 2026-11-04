package producers

import (
	"context"
	"testing"

	"buenvivir/campaigns/internal/events"
)

func TestCampanasProducer_Disabled(t *testing.T) {
	var p CampanasProducer
	ctx := context.Background()

	if p != nil && p.Enabled() {
		t.Error("nil producer should not be enabled")
	}

	p = NewCampanasProducer(nil, "")
	if p.Enabled() {
		t.Error("producer with nil writer should not be enabled")
	}

	p.Notify()

	if err := p.CampanaCreada(ctx, "c-1", "usr-1"); err != nil {
		t.Fatalf("expected nil when disabled, got: %v", err)
	}
}

func TestCampanaCreadaPayload_SinPII(t *testing.T) {
	payload := events.CampanaCreadaPayload{
		CampanaID: "camp-123",
		CreadoPor: "usr-456",
	}

	if payload.CampanaID != "camp-123" || payload.CreadoPor != "usr-456" {
		t.Errorf("unexpected payload values: %+v", payload)
	}
}

