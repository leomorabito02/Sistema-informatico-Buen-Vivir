package rules_test

import (
	"context"
	"encoding/json"
	"testing"

	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/rules"
)

// HU-12 y HU-38: el rechazo de una cuenta manda un email propio, sin
// request_id (no hay enlace que pedirle a Auth).
func TestEmailSolicitadoRule_CuentaRechazada(t *testing.T) {
	rule := &rules.EmailSolicitadoRule{}
	env := newTestEnvelope(t, rule.EventType(), events.EmailSolicitadoPayload{RequestID: "rej-usr-1", UserID: "usr-1", Tipo: "cuenta_rechazada"}, "usr-admin")

	plans, err := rule.Build(context.Background(), env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plans) != 1 || !plans[0].SendEmail || plans[0].RecipientUserIDs[0] != "usr-1" {
		t.Fatalf("plan inesperado: %+v", plans)
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(plans[0].Data), &data); err != nil {
		t.Fatal(err)
	}
	if _, ok := data["request_id"]; ok {
		t.Errorf("no debe llevar request_id: %v", data)
	}
	if !plans[0].IsCritical {
		t.Errorf("el rechazo tiene que ser crítico para llegar a emails sin verificar")
	}
	if data["template_id"] != "general_alert" {
		t.Errorf("template inesperado: %v", data["template_id"])
	}
}

func TestEmailSolicitadoRule_CuentaAprobada(t *testing.T) {
	// Arrange
	rule := &rules.EmailSolicitadoRule{}
	env := newTestEnvelope(t, rule.EventType(), events.EmailSolicitadoPayload{
		RequestID: "app-usr-1",
		UserID:    "usr-1",
		Tipo:      "cuenta_aprobada",
	}, "usr-admin")

	// Act
	plans, err := rule.Build(context.Background(), env)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plans) != 1 || !plans[0].SendEmail || plans[0].RecipientUserIDs[0] != "usr-1" {
		t.Fatalf("unexpected plan: %+v", plans)
	}
	if !plans[0].IsCritical {
		t.Errorf("expected IsCritical=true for cuenta_aprobada")
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(plans[0].Data), &data); err != nil {
		t.Fatal(err)
	}
	if data["template_id"] != "account_approved" {
		t.Errorf("expected account_approved template, got %v", data["template_id"])
	}
}
