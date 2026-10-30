package repositories_test

import (
	"context"
	"testing"
	"time"

	"buenvivir/notifications/internal/models"
	"buenvivir/notifications/internal/repositories"
)

func TestNotificationRepository_Default(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	list, err := repo.FindByUserID(context.Background(), "usr-default")
	if err != nil || len(list) == 0 {
		t.Fatalf("expected default list, got err: %v", err)
	}
}

func TestNotificationRepository_RecipientsFlow(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	ctx := context.Background()

	notif := &models.Notification{
		ID:            "ntf-100",
		SourceService: "education",
		Title:         "Inscripción exitosa",
		Message:       "Te has inscrito al curso",
		SendInApp:     true,
		CreatedAt:     time.Now(),
	}
	if err := repo.CreateNotification(ctx, notif); err != nil {
		t.Fatalf("failed to create notification: %v", err)
	}

	recipients := []models.NotificationRecipient{
		{
			RecipientID:    "rec-100",
			NotificationID: "ntf-100",
			UserID:         "usr-user1",
			IsRead:         false,
			EmailStatus:    models.DeliveryStatusPending,
			PushStatus:     models.DeliveryStatusPending,
			CreatedAt:      time.Now(),
		},
		{
			RecipientID:    "rec-101",
			NotificationID: "ntf-100",
			UserID:         "usr-user1",
			IsRead:         false,
			EmailStatus:    models.DeliveryStatusNotRequested,
			PushStatus:     models.DeliveryStatusNotRequested,
			CreatedAt:      time.Now(),
		},
	}
	if err := repo.CreateRecipients(ctx, recipients); err != nil {
		t.Fatalf("failed to create recipients: %v", err)
	}

	userRecipients, err := repo.FindRecipientsByUserID(ctx, "usr-user1", 10, 0)
	if err != nil || len(userRecipients) != 2 {
		t.Fatalf("expected 2 recipients, got %d", len(userRecipients))
	}

	single, err := repo.FindRecipientByID(ctx, "rec-100", "usr-user1")
	if err != nil || single == nil || single.Notification == nil {
		t.Fatalf("expected recipient with notification: %v", err)
	}
}

func TestNotificationRepository_ReadAndUnreadCount(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	ctx := context.Background()

	_ = repo.CreateRecipients(ctx, []models.NotificationRecipient{
		{RecipientID: "rec-100", UserID: "usr-user1", IsRead: false},
		{RecipientID: "rec-101", UserID: "usr-user1", IsRead: false},
	})

	unread, err := repo.CountUnreadByUserID(ctx, "usr-user1")
	if err != nil || unread != 2 {
		t.Fatalf("expected unread count 2, got %d", unread)
	}

	if err := repo.MarkRecipientAsRead(ctx, "rec-100", "usr-user1"); err != nil {
		t.Fatalf("failed to mark read: %v", err)
	}

	if err := repo.MarkAllRecipientsAsRead(ctx, "usr-user1"); err != nil {
		t.Fatalf("failed to mark all read: %v", err)
	}

	unreadAfter, _ := repo.CountUnreadByUserID(ctx, "usr-user1")
	if unreadAfter != 0 {
		t.Fatalf("expected 0 unread, got %d", unreadAfter)
	}
}

func TestNotificationRepository_HardDelete(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	ctx := context.Background()

	_ = repo.CreateRecipients(ctx, []models.NotificationRecipient{
		{RecipientID: "rec-100", UserID: "usr-user1"},
	})

	if err := repo.HardDeleteRecipient(ctx, "rec-100", "usr-user1"); err != nil {
		t.Fatalf("failed to delete: %v", err)
	}

	if _, err := repo.FindRecipientByID(ctx, "rec-100", "usr-user1"); err == nil {
		t.Fatalf("expected error finding deleted record")
	}

	if err := repo.HardDeleteRecipient(ctx, "rec-100", "usr-wrong"); err == nil {
		t.Fatalf("expected error deleting missing/unowned record")
	}
}

func TestNotificationRepository_DeliveryAndPreferences(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	ctx := context.Background()

	_ = repo.CreateRecipients(ctx, []models.NotificationRecipient{
		{RecipientID: "rec-100", UserID: "usr-user1", EmailStatus: models.DeliveryStatusPending},
	})

	pending, _ := repo.FindPendingRecipients(ctx, 10)
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(pending))
	}

	msgID := "resend-msg-123"
	_ = repo.UpdateDeliveryStatus(ctx, "rec-100", models.DeliveryStatusSent, models.DeliveryStatusSent, &msgID, nil, nil)

	pref := &models.NotificationPreference{
		ID:           "pref-1",
		UserID:       "usr-user1",
		Category:     models.CategoryEducation,
		EmailEnabled: true,
	}
	if err := repo.UpsertPreference(ctx, pref); err != nil {
		t.Fatalf("failed upsert: %v", err)
	}

	saved, err := repo.GetPreference(ctx, "usr-user1", models.CategoryEducation)
	if err != nil || saved == nil || !saved.EmailEnabled {
		t.Fatalf("expected saved preference with EmailEnabled=true")
	}

	allPrefs, err := repo.GetPreferencesByUserID(ctx, "usr-user1")
	if err != nil || len(allPrefs) != 1 {
		t.Fatalf("expected 1 preference in list, got %d", len(allPrefs))
	}
}
