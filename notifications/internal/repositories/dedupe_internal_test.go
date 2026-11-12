package repositories

import (
	"context"
	"testing"
	"time"

	"buenvivir/notifications/internal/models"
)

// Problema 3 de la auditoría (2026-11-11): si Pub/Sub reintenta el mismo
// evento, el destinatario no recibe la notificación dos veces y tampoco
// queda guardada una notificación sin destinatarios.
func TestCreateNotificationAndRecipients_ReintentoNoDuplica(t *testing.T) {
	repo := NewMemoryNotificationRepository().(*memoryNotificationRepository)
	ctx := context.Background()
	antes := len(repo.notifications)

	guardar := func(id string) {
		t.Helper()
		n := &models.Notification{ID: id, Title: "Nuevo hito en tu proceso", SendInApp: true, CreatedAt: time.Now()}
		recs := []models.NotificationRecipient{{
			RecipientID: "rcp-" + id, NotificationID: id, UserID: "usr-part-1",
			DedupeKey: "evt-1_participant_part-1", CreatedAt: time.Now(),
		}}
		if err := repo.CreateNotificationAndRecipients(ctx, n, recs); err != nil {
			t.Fatal(err)
		}
	}
	guardar("ntf-a")
	guardar("ntf-b") // reintento del mismo evento: nuevo id, mismo dedupe

	n := 0
	for _, r := range repo.recipients {
		if r.UserID == "usr-part-1" && r.DedupeKey == "evt-1_participant_part-1" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("el participante recibió %d copias, esperaba 1", n)
	}
	if got := len(repo.notifications) - antes; got != 1 {
		t.Fatalf("se guardaron %d notificaciones, esperaba 1 (sin huérfanas)", got)
	}
	if _, ok := repo.notifications["ntf-b"]; ok {
		t.Fatal("el reintento dejó una notificación huérfana")
	}
}
