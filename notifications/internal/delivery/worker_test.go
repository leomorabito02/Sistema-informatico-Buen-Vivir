package delivery_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"buenvivir/notifications/internal/clients"
	"buenvivir/notifications/internal/delivery"
	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/models"
	"buenvivir/notifications/internal/repositories"
	"buenvivir/notifications/internal/services"
	"buenvivir/notifications/internal/templates"
)

type mockEmailService struct {
	sendErr error
	sentMsgs []services.EmailMessage
}

func (m *mockEmailService) SendEmail(_ context.Context, msg services.EmailMessage) (string, error) {
	if m.sendErr != nil {
		return "", m.sendErr
	}
	m.sentMsgs = append(m.sentMsgs, msg)
	return "msg-id-123", nil
}

type mockPushService struct {
	sendErr error
}

func (m *mockPushService) SendMulticast(_ context.Context, tokens []string, _, _ string, _ map[string]string) (*services.MulticastResult, error) {
	if m.sendErr != nil {
		return nil, m.sendErr
	}
	return &services.MulticastResult{SuccessCount: len(tokens)}, nil
}

type mockAuthClient struct {
	tokens []string
}

func (m *mockAuthClient) GetUserDeviceTokens(_ context.Context, _ string) ([]string, error) {
	return m.tokens, nil
}
func (m *mockAuthClient) RevokeDeviceTokens(_ context.Context, _ string, _ []string) error {
	return nil
}
func (m *mockAuthClient) GetCampaignRecipients(_ context.Context) ([]clients.RecipientContact, error) {
	return nil, nil
}
func (m *mockAuthClient) IsUserEmailVerified(_ context.Context, _ string) (bool, error) {
	return true, nil
}
func (m *mockAuthClient) GetUserContact(_ context.Context, userID string) (*clients.RecipientContact, error) {
	return &clients.RecipientContact{
		UserID:          userID,
		Email:           "user@example.com",
		FirstName:       "Juan",
		LastName:        "Perez",
		IsEmailVerified: true,
	}, nil
}
func (m *mockAuthClient) GetActiveUserIDsByRoles(_ context.Context, _ []string) ([]string, error) {
	return nil, nil
}
func (m *mockAuthClient) GetParticipantUserID(_ context.Context, _ string) (string, error) {
	return "usr-resolved", nil
}
func (m *mockAuthClient) IssueActionLink(_ context.Context, _, _, _ string) (string, error) {
	return "https://example.com/action", nil
}
func (m *mockAuthClient) Close() error { return nil }

func TestDeliveryWorker_Flush_Success(t *testing.T) {
	// Arrange
	repo := repositories.NewMemoryNotificationRepository()
	emailSvc := &mockEmailService{}
	pushSvc := &mockPushService{}
	authCl := &mockAuthClient{tokens: []string{"fcm-tok-1"}}

	notif := &models.Notification{
		ID:        "ntf-1",
		Category:  models.CategoryEducation,
		Title:     "Test Title",
		Message:   "Test Message",
		SendEmail: true,
		SendPush:  true,
	}
	_ = repo.CreateNotification(context.Background(), notif)

	rec := models.NotificationRecipient{
		RecipientID:    "rcp-1",
		NotificationID: notif.ID,
		UserID:         "usr-1",
		EmailStatus:    models.DeliveryStatusPending,
		PushStatus:     models.DeliveryStatusPending,
	}
	_ = repo.CreateRecipients(context.Background(), []models.NotificationRecipient{rec})

	worker := delivery.NewWorker(repo, emailSvc, pushSvc, authCl, nil, nil)

	// Act
	count, err := worker.Flush(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 processed recipient, got %d", count)
	}

	saved, err := repo.FindRecipientByID(context.Background(), "rcp-1", "usr-1")
	if err != nil {
		t.Fatalf("failed to find recipient: %v", err)
	}
	if saved.EmailStatus != models.DeliveryStatusSent || saved.PushStatus != models.DeliveryStatusSent {
		t.Errorf("expected both channels sent, got email=%s push=%s", saved.EmailStatus, saved.PushStatus)
	}
}

func TestDeliveryWorker_Flush_RetryOnTransientFailure(t *testing.T) {
	// Arrange
	repo := repositories.NewMemoryNotificationRepository()
	emailSvc := &mockEmailService{sendErr: errors.New("smtp connection failed")}
	authCl := &mockAuthClient{}

	notif := &models.Notification{
		ID:        "ntf-2",
		Category:  models.CategoryEducation,
		Title:     "Alert",
		Message:   "Message",
		SendEmail: true,
	}
	_ = repo.CreateNotification(context.Background(), notif)

	rec := models.NotificationRecipient{
		RecipientID:    "rcp-2",
		NotificationID: notif.ID,
		UserID:         "usr-2",
		EmailStatus:    models.DeliveryStatusPending,
		Attempts:       0,
	}
	_ = repo.CreateRecipients(context.Background(), []models.NotificationRecipient{rec})

	worker := delivery.NewWorker(repo, emailSvc, nil, authCl, nil, nil)
	worker.Now = func() time.Time { return time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC) }

	// Act
	_, err := worker.Flush(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}

	saved, _ := repo.FindRecipientByID(context.Background(), "rcp-2", "usr-2")
	if saved.Attempts != 1 {
		t.Errorf("expected attempts = 1, got %d", saved.Attempts)
	}
	if saved.NextAttemptAt == nil || !saved.NextAttemptAt.After(worker.Now()) {
		t.Errorf("expected NextAttemptAt to be set in the future, got %v", saved.NextAttemptAt)
	}
}

func TestDeliveryWorker_Flush_DeadStatusAfterMaxAttempts(t *testing.T) {
	// Arrange
	repo := repositories.NewMemoryNotificationRepository()
	emailSvc := &mockEmailService{sendErr: errors.New("fatal provider rejected")}
	authCl := &mockAuthClient{}

	notif := &models.Notification{
		ID:        "ntf-3",
		Category:  models.CategoryEducation,
		Title:     "Alert",
		Message:   "Message",
		SendEmail: true,
	}
	_ = repo.CreateNotification(context.Background(), notif)

	rec := models.NotificationRecipient{
		RecipientID:    "rcp-3",
		NotificationID: notif.ID,
		UserID:         "usr-3",
		EmailStatus:    models.DeliveryStatusPending,
		Attempts:       4, // Next attempt is 5, which reaches defaultMaxAttempts (5)
	}
	_ = repo.CreateRecipients(context.Background(), []models.NotificationRecipient{rec})

	worker := delivery.NewWorker(repo, emailSvc, nil, authCl, nil, nil)

	// Act
	_, err := worker.Flush(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}

	saved, _ := repo.FindRecipientByID(context.Background(), "rcp-3", "usr-3")
	if saved.EmailStatus != models.DeliveryStatusDead {
		t.Errorf("expected EmailStatus dead, got %s", saved.EmailStatus)
	}
}

func TestDeliveryWorker_FlushHandler(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	worker := delivery.NewWorker(repo, nil, nil, nil, nil, nil)

	verifier := events.NoVerifier{}
	handler := worker.FlushHandler(verifier)

	req := httptest.NewRequest("POST", "/internal/delivery/flush", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rec.Code)
	}
}

func TestDeliveryWorker_Flush_IssueActionLinkOnDemand(t *testing.T) {
	// Arrange
	repo := repositories.NewMemoryNotificationRepository()
	emailSvc := &mockEmailService{}
	authCl := &mockAuthClient{}

	notif := &models.Notification{
		ID:        "ntf-sec-1",
		Category:  models.CategorySecurity,
		Title:     "Verifica tu correo",
		Message:   "Por favor verifica tu correo",
		SendEmail: true,
		Data:      `{"request_id":"req-999","tipo":"verificacion","template_id":"email_verification"}`,
	}
	_ = repo.CreateNotification(context.Background(), notif)

	rec := models.NotificationRecipient{
		RecipientID:    "rcp-sec-1",
		NotificationID: notif.ID,
		UserID:         "usr-sec-1",
		EmailStatus:    models.DeliveryStatusPending,
	}
	_ = repo.CreateRecipients(context.Background(), []models.NotificationRecipient{rec})

	tmplEngine, err := templates.NewTemplateEngine()
	if err != nil {
		t.Fatalf("failed to create template engine: %v", err)
	}
	worker := delivery.NewWorker(repo, emailSvc, nil, authCl, tmplEngine, nil)

	// Act
	n, err := worker.Flush(context.Background())

	// Assert
	if err != nil || n != 1 {
		t.Fatalf("expected flush 1 recipient, got %d, err: %v", n, err)
	}
	if len(emailSvc.sentMsgs) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(emailSvc.sentMsgs))
	}
	if !strings.Contains(emailSvc.sentMsgs[0].Html, "https://example.com/action") {
		t.Errorf("expected html to contain action url, got: %s", emailSvc.sentMsgs[0].Html)
	}
}

func TestDeliveryWorker_InlineFlushEnabled(t *testing.T) {
	t.Run("disabled by default locally", func(t *testing.T) {
		// Arrange
		t.Setenv("K_SERVICE", "")
		t.Setenv("DELIVERY_INLINE_FLUSH", "")

		// Act & Assert
		if delivery.InlineFlushEnabled() {
			t.Fatal("expected inline flush disabled by default")
		}
	})

	t.Run("enabled automatically on Cloud Run", func(t *testing.T) {
		// Arrange
		t.Setenv("K_SERVICE", "notifications-testing")
		t.Setenv("DELIVERY_INLINE_FLUSH", "")

		// Act & Assert
		if !delivery.InlineFlushEnabled() {
			t.Fatal("expected inline flush enabled when K_SERVICE is set")
		}
	})

	t.Run("explicit false overrides K_SERVICE", func(t *testing.T) {
		// Arrange
		t.Setenv("K_SERVICE", "notifications-testing")
		t.Setenv("DELIVERY_INLINE_FLUSH", "false")

		// Act & Assert
		if delivery.InlineFlushEnabled() {
			t.Fatal("expected explicit false to disable inline flush")
		}
	})

	t.Run("explicit true enables locally", func(t *testing.T) {
		// Arrange
		t.Setenv("K_SERVICE", "")
		t.Setenv("DELIVERY_INLINE_FLUSH", "true")

		// Act & Assert
		if !delivery.InlineFlushEnabled() {
			t.Fatal("expected explicit true to enable inline flush")
		}
	})
}

func TestDeliveryWorker_NotifyInlineAndShutdown(t *testing.T) {
	// Arrange
	t.Setenv("K_SERVICE", "notifications-testing")
	t.Setenv("DELIVERY_INLINE_FLUSH", "")

	repo := repositories.NewMemoryNotificationRepository()
	emailSvc := &mockEmailService{}
	authCl := &mockAuthClient{}

	notif := &models.Notification{
		ID:        "ntf-inline-1",
		Category:  models.CategoryEducation,
		Title:     "Inline Title",
		Message:   "Inline Message",
		SendEmail: true,
	}
	_ = repo.CreateNotification(context.Background(), notif)
	_ = repo.CreateRecipients(context.Background(), []models.NotificationRecipient{{
		RecipientID:    "rcp-inline-1",
		NotificationID: notif.ID,
		UserID:         "usr-inline-1",
		EmailStatus:    models.DeliveryStatusPending,
	}})

	worker := delivery.NewWorker(repo, emailSvc, nil, authCl, nil, nil)

	// Act
	worker.NotifyInline(context.Background())
	worker.Shutdown(context.Background())

	// Assert
	saved, err := repo.FindRecipientByID(context.Background(), "rcp-inline-1", "usr-inline-1")
	if err != nil {
		t.Fatalf("failed to find recipient: %v", err)
	}
	if saved.EmailStatus != models.DeliveryStatusSent {
		t.Fatalf("expected email sent inline, got %s", saved.EmailStatus)
	}

	var nilWorker *delivery.Worker
	nilWorker.Notify()
	nilWorker.NotifyInline(context.Background())
	nilWorker.Shutdown(context.Background())
}
