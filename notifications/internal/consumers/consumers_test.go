package consumers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"buenvivir/notifications/internal/consumers"
	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/repositories"
	"buenvivir/notifications/internal/rules"
)

type dummyAudienceResolver struct{}

func (d *dummyAudienceResolver) ResolveUserIDs(_ context.Context, plan rules.NotificationPlan) ([]string, error) {
	if len(plan.RecipientUserIDs) > 0 {
		return plan.RecipientUserIDs, nil
	}
	if plan.ParticipantID != "" {
		return []string{"usr-" + plan.ParticipantID}, nil
	}
	if plan.TargetRole != nil {
		return []string{"usr-admin-1"}, nil
	}
	return nil, nil
}

func TestEducacionConsumer(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	store := events.NewMemoryStore(100)
	verifier := events.NoVerifier{}
	resolver := &dummyAudienceResolver{}

	registry := rules.NewRegistry()
	registry.Register(&rules.InscripcionCreadaRule{})

	consumer := consumers.NewEducacionConsumer(verifier, store, registry, resolver, repo, nil, nil)

	payload := events.InscripcionCreadaPayload{
		InscripcionID:    "ins-1",
		CursoID:          "cur-1",
		ParticipanteID:   "part-1",
		FechaInscripcion: time.Now(),
	}
	env, err := events.New(context.Background(), "servicio-educacion", events.InscripcionCreada, payload)
	if err != nil {
		t.Fatalf("failed to create envelope: %v", err)
	}

	envBytes, _ := env.Marshal()
	pushBody, err := events.BuildPushBody(envBytes, env.Attributes(), "msg-1", "notificaciones-educacion", 1)
	if err != nil {
		t.Fatalf("failed to build push body: %v", err)
	}

	// Act 1: Send push request
	req := httptest.NewRequest("POST", "/eventos/educacion", bytes.NewReader(pushBody))
	rec := httptest.NewRecorder()
	consumer.ServeHTTP(rec, req)

	// Assert 1: Success status 204
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %d", rec.Code)
	}

	// Verify notification persisted
	recs, err := repo.FindRecipientsByUserID(context.Background(), "usr-part-1", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 persisted recipient, got %d", len(recs))
	}
	if recs[0].Notification == nil || recs[0].Notification.Title != "Inscripción confirmada" {
		t.Errorf("unexpected notification content: %+v", recs[0])
	}

	// Act 2: Idempotent duplicate delivery
	recDup := httptest.NewRecorder()
	reqDup := httptest.NewRequest("POST", "/eventos/educacion", bytes.NewReader(pushBody))
	consumer.ServeHTTP(recDup, reqDup)

	if recDup.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for duplicate, got %d", recDup.Code)
	}

	// Still only 1 record
	recsAfter, _ := repo.FindRecipientsByUserID(context.Background(), "usr-part-1", 10, 0)
	if len(recsAfter) != 1 {
		t.Fatalf("expected still 1 recipient after duplicate, got %d", len(recsAfter))
	}
}

func TestPromocionHumanaConsumer(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	store := events.NewMemoryStore(100)
	verifier := events.NoVerifier{}
	resolver := &dummyAudienceResolver{}

	registry := rules.NewRegistry()
	registry.Register(&rules.HitoRegistradoRule{})

	consumer := consumers.NewPromocionHumanaConsumer(verifier, store, registry, resolver, repo, nil, nil)

	payload := events.HitoPayload{
		HitoID:              "hito-55",
		FichaID:             "f-10",
		ParticipanteID:      "part-99",
		Fecha:               "2026-11-04",
		VisibleParticipante: true,
	}
	env, err := events.New(context.Background(), "servicio-promocion-humana", events.HitoRegistrado, payload)
	if err != nil {
		t.Fatalf("failed to create envelope: %v", err)
	}

	envBytes, _ := env.Marshal()
	pushBody, _ := events.BuildPushBody(envBytes, env.Attributes(), "msg-2", "notificaciones-promocion-humana", 1)

	req := httptest.NewRequest("POST", "/eventos/promocion-humana", bytes.NewReader(pushBody))
	rec := httptest.NewRecorder()
	consumer.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	recs, _ := repo.FindRecipientsByUserID(context.Background(), "usr-part-99", 10, 0)
	if len(recs) != 1 {
		t.Fatalf("expected 1 recipient for participant, got %d", len(recs))
	}
}

func TestUsuariosConsumer(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	store := events.NewMemoryStore(100)
	verifier := events.NoVerifier{}
	resolver := &dummyAudienceResolver{}

	registry := rules.NewRegistry()
	registry.Register(&rules.ParticipanteEliminadoRule{})

	consumer := consumers.NewUsuariosConsumer(verifier, store, registry, resolver, repo, nil, nil)

	payload := events.ParticipantePayload{
		ParticipanteID: "part-del-1",
		UserID:         "usr-part-del-1",
	}
	env, err := events.New(context.Background(), "servicio-autenticacion", events.ParticipanteEliminado, payload)
	if err != nil {
		t.Fatalf("failed to create envelope: %v", err)
	}

	envBytes, _ := env.Marshal()
	pushBody, _ := events.BuildPushBody(envBytes, env.Attributes(), "msg-3", "notificaciones-usuarios", 1)

	req := httptest.NewRequest("POST", "/eventos/usuarios", bytes.NewReader(pushBody))
	rec := httptest.NewRecorder()
	consumer.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	recs, _ := repo.FindRecipientsByUserID(context.Background(), "usr-admin-1", 10, 0)
	if len(recs) != 1 {
		t.Fatalf("expected 1 recipient for admin, got %d", len(recs))
	}
}
