package rules_test

import (
	"context"
	"testing"
	"time"

	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/models"
	"buenvivir/notifications/internal/rules"
)

func newTestEnvelope(t *testing.T, eventType string, payload any, actor string) events.Envelope {
	t.Helper()
	ctx := context.Background()
	if actor != "" {
		ctx = events.WithActor(ctx, actor)
	}
	env, err := events.New(ctx, "test-producer", eventType, payload)
	if err != nil {
		t.Fatalf("failed to create test envelope: %v", err)
	}
	return env
}

func TestEducationRules(t *testing.T) {
	ctx := context.Background()

	t.Run("InscripcionCreadaRule_WithTeacher", func(t *testing.T) {
		rule := &rules.InscripcionCreadaRule{}
		payload := events.InscripcionCreadaPayload{
			InscripcionID:    "ins-1",
			CursoID:          "cur-10",
			ParticipanteID:   "part-99",
			DocenteID:        "usr-teacher-1",
			FechaInscripcion: time.Now(),
		}
		env := newTestEnvelope(t, rule.EventType(), payload, "usr-volunteer")

		plans, err := rule.Build(ctx, env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(plans) != 2 {
			t.Fatalf("expected 2 plans (participant + teacher), got %d", len(plans))
		}
		if plans[0].ParticipantID != "part-99" || plans[0].Category != models.CategoryEducation {
			t.Errorf("unexpected participant plan: %+v", plans[0])
		}
		if len(plans[1].RecipientUserIDs) != 1 || plans[1].RecipientUserIDs[0] != "usr-teacher-1" {
			t.Errorf("unexpected teacher plan: %+v", plans[1])
		}
	})

	t.Run("InscripcionCreadaRule_WithTitleEnriched", func(t *testing.T) {
		rule := &rules.InscripcionCreadaRule{}
		payload := events.InscripcionCreadaPayload{
			InscripcionID:    "ins-2",
			CursoID:          "cur-10",
			ParticipanteID:   "part-99",
			Titulo:           "Electricidad Básica",
			FechaInscripcion: time.Now(),
		}
		env := newTestEnvelope(t, rule.EventType(), payload, "usr-volunteer")

		plans, err := rule.Build(ctx, env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if plans[0].Title != "Inscripción confirmada: Electricidad Básica" {
			t.Errorf("expected enriched title, got %q", plans[0].Title)
		}
	})

	t.Run("ActividadRecordatorioRule", func(t *testing.T) {
		rule := &rules.ActividadRecordatorioRule{}
		payload := events.ActividadRecordatorioPayload{
			ActividadID:    "act-1",
			CursoID:        "cur-10",
			ParticipanteID: "part-99",
			Fecha:          "2026-11-05",
			Titulo:         "Clase Práctica",
		}
		env := newTestEnvelope(t, rule.EventType(), payload, "")

		plans, err := rule.Build(ctx, env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(plans) != 1 {
			t.Fatalf("expected 1 plan, got %d", len(plans))
		}
		if plans[0].Title != "Recordatorio: Clase Práctica" || plans[0].ParticipantID != "part-99" {
			t.Errorf("unexpected reminder plan: %+v", plans[0])
		}
	})

	t.Run("ContenidoDocenteSubidoRule", func(t *testing.T) {
		rule := &rules.ContenidoDocenteSubidoRule{}
		payload := events.ContenidoDocenteSubidoPayload{
			ContenidoID: "mat-1",
			DocenteID:   "usr-t1",
			Tipo:        "material",
		}
		env := newTestEnvelope(t, rule.EventType(), payload, "usr-t1")

		plans, err := rule.Build(ctx, env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(plans) != 1 {
			t.Fatalf("expected 1 plan, got %d", len(plans))
		}
		if plans[0].TargetRole == nil || *plans[0].TargetRole != "role_admin" {
			t.Errorf("expected role_admin target, got %v", plans[0].TargetRole)
		}
	})

	t.Run("ContenidoDocenteModeradoRule", func(t *testing.T) {
		rule := &rules.ContenidoDocenteModeradoRule{}
		payload := events.ContenidoDocenteModeradoPayload{
			ContenidoID: "mat-1",
			DocenteID:   "usr-t1",
			Tipo:        "material",
			Resultado:   "aprobado",
		}
		env := newTestEnvelope(t, rule.EventType(), payload, "usr-admin")

		plans, err := rule.Build(ctx, env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(plans) != 1 || len(plans[0].RecipientUserIDs) != 1 || plans[0].RecipientUserIDs[0] != "usr-t1" {
			t.Errorf("expected recipient usr-t1, got %+v", plans)
		}
	})
}

func TestHumanPromotionRules(t *testing.T) {
	ctx := context.Background()

	t.Run("FichaCreadaRule", func(t *testing.T) {
		rule := &rules.FichaCreadaRule{}
		payload := events.FichaPayload{
			FichaID:        "f-1",
			ParticipanteID: "part-1",
			CreadoPor:      "vol-1",
		}
		env := newTestEnvelope(t, rule.EventType(), payload, "vol-1")

		plans, err := rule.Build(ctx, env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(plans) != 1 || *plans[0].TargetRole != "role_admin" {
			t.Fatalf("expected role_admin plan, got %+v", plans)
		}
	})

	t.Run("FichaActualizadaRule_DifferentActor", func(t *testing.T) {
		rule := &rules.FichaActualizadaRule{}
		payload := events.FichaPayload{
			FichaID:        "f-1",
			ParticipanteID: "part-1",
			CreadoPor:      "vol-original",
		}
		env := newTestEnvelope(t, rule.EventType(), payload, "vol-different")

		plans, err := rule.Build(ctx, env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(plans) != 1 || plans[0].RecipientUserIDs[0] != "vol-original" {
			t.Fatalf("expected plan for vol-original, got %+v", plans)
		}
	})

	t.Run("HitoRegistradoRule_VisibleAndAssigned", func(t *testing.T) {
		rule := &rules.HitoRegistradoRule{}
		payload := events.HitoPayload{
			HitoID:              "h-1",
			FichaID:             "f-1",
			ParticipanteID:      "part-2",
			VisibleParticipante: true,
			ResponsableID:       "vol-responsable",
		}
		env := newTestEnvelope(t, rule.EventType(), payload, "vol-creator")

		plans, err := rule.Build(ctx, env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(plans) != 2 {
			t.Fatalf("expected 2 plans, got %d", len(plans))
		}
		if plans[0].ParticipantID != "part-2" {
			t.Errorf("expected participant plan, got %+v", plans[0])
		}
		if plans[1].RecipientUserIDs[0] != "vol-responsable" {
			t.Errorf("expected responsible plan, got %+v", plans[1])
		}
	})
}

func TestRegistry(t *testing.T) {
	reg := rules.NewRegistry()
	reg.Register(&rules.InscripcionCreadaRule{})
	reg.Register(&rules.FichaCreadaRule{})

	rule, ok := reg.Get(events.InscripcionCreada)
	if !ok || rule.EventType() != events.InscripcionCreada {
		t.Errorf("expected InscripcionCreada rule, got %v", rule)
	}

	types := reg.EventTypes()
	if len(types) != 2 {
		t.Errorf("expected 2 registered types, got %d", len(types))
	}
}

func TestCampaignRules(t *testing.T) {
	ctx := context.Background()

	t.Run("CampanaCreadaRule", func(t *testing.T) {
		rule := &rules.CampanaCreadaRule{}
		payload := events.CampanaCreadaPayload{
			CampanaID: "camp-123",
			CreadoPor: "vol-1",
		}
		env := newTestEnvelope(t, rule.EventType(), payload, "vol-1")

		plans, err := rule.Build(ctx, env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(plans) != 3 {
			t.Fatalf("expected 3 plans (admin, participant, volunteer), got %d", len(plans))
		}
		if *plans[0].TargetRole != "role_admin" || *plans[0].ActionURL != "/dashboard/campanas" {
			t.Errorf("expected role_admin target with /dashboard/campanas, got %v %v", *plans[0].TargetRole, *plans[0].ActionURL)
		}
		if *plans[1].TargetRole != "role_participant" || *plans[1].ActionURL != "/campanas-publico" {
			t.Errorf("expected role_participant target with /campanas-publico, got %v %v", *plans[1].TargetRole, *plans[1].ActionURL)
		}
		if *plans[2].TargetRole != "role_volunteer" || *plans[2].ActionURL != "/dashboard/campanas" {
			t.Errorf("expected role_volunteer target with /dashboard/campanas, got %v %v", *plans[2].TargetRole, *plans[2].ActionURL)
		}
	})
}

func TestAccountRules(t *testing.T) {
	ctx := context.Background()

	t.Run("UsuarioCreadoRule_WithApprovalRequiredRole", func(t *testing.T) {
		rule := &rules.UsuarioCreadoRule{}
		payload := events.UsuarioCreadoPayload{
			UserID: "usr-new-1",
			RoleID: "role_volunteer",
		}
		env := newTestEnvelope(t, rule.EventType(), payload, "usr-new-1")

		plans, err := rule.Build(ctx, env)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(plans) != 2 {
			t.Fatalf("expected 2 plans (welcome + admin review), got %d", len(plans))
		}
		if *plans[0].ActionURL != "/dashboard/perfil" {
			t.Errorf("expected /dashboard/perfil, got %s", *plans[0].ActionURL)
		}
		if plans[1].TargetRole == nil || *plans[1].TargetRole != "role_admin" || *plans[1].ActionURL != "/dashboard/usuarios" {
			t.Errorf("unexpected admin review plan: %+v", plans[1])
		}
	})
}


