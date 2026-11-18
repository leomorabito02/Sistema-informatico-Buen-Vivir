package audit

import (
	"context"
	"testing"
)

// RecordTx se prueba contra el emulador en repositories/events_test.go.
func TestClientIP(t *testing.T) {
	if ClientIP(context.Background()) != "" {
		t.Error("sin IP en el contexto")
	}
	if ClientIP(WithClientIP(context.Background(), "10.0.0.1")) != "10.0.0.1" {
		t.Error("no guardó la IP")
	}
}
