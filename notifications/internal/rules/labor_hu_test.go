package rules_test

import (
	"context"
	"testing"

	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/rules"
)

// HU-41: se avisa a la empresa solo si la oferta la borró la gestión.
func TestOfertaEliminadaRule(t *testing.T) {
	rule := &rules.OfertaEliminadaRule{}
	env := newTestEnvelope(t, rule.EventType(), events.OfertaEliminadaPayload{OfertaID: "of-1", EmpresaID: "usr-emp", PorGestion: true}, "usr-admin")
	plans, err := rule.Build(context.Background(), env)
	if err != nil || len(plans) != 1 || plans[0].RecipientUserIDs[0] != "usr-emp" {
		t.Fatalf("plan inesperado: %+v, %v", plans, err)
	}
	env = newTestEnvelope(t, rule.EventType(), events.OfertaEliminadaPayload{OfertaID: "of-1", EmpresaID: "usr-emp"}, "usr-emp")
	if plans, _ := rule.Build(context.Background(), env); len(plans) != 0 {
		t.Fatalf("si la borra la empresa no hay aviso: %+v", plans)
	}
}

// HU-44: un aviso por participante, sin duplicados.
func TestOfertaCompatibleRule(t *testing.T) {
	rule := &rules.OfertaCompatibleRule{}
	env := newTestEnvelope(t, rule.EventType(), events.OfertaCompatiblePayload{OfertaID: "of-1", ParticipanteIDs: []string{"part-1", "part-2", "part-1", ""}}, "usr-admin")
	plans, err := rule.Build(context.Background(), env)
	if err != nil || len(plans) != 2 {
		t.Fatalf("esperaba 2 planes: %+v, %v", plans, err)
	}
	if plans[0].AudienceType != rules.AudienceParticipant || plans[0].ParticipantID != "part-1" || plans[0].ActionURL == nil || *plans[0].ActionURL != "/dashboard/ofertas" {
		t.Errorf("plan inesperado: %+v", plans[0])
	}
}
