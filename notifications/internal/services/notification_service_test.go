package services_test

import (
	"context"
	"testing"

	"buenvivir/notifications/internal/clients"
	"buenvivir/notifications/internal/dto"
	"buenvivir/notifications/internal/models"
	"buenvivir/notifications/internal/repositories"
	"buenvivir/notifications/internal/services"
)

func TestNotificationService_CRUD_ValidationErrors(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	svc := services.NewNotificationService(repo)
	ctx := context.Background()

	if _, err := svc.GetNotifications(ctx, "", 10, 0); err == nil {
		t.Fatalf("expected error on empty userID")
	}
	if _, err := svc.GetNotificationDetail(ctx, "", "usr-1"); err == nil {
		t.Fatalf("expected error on empty recipientID")
	}
	if err := svc.MarkAsRead(ctx, "rec-1", ""); err == nil {
		t.Fatalf("expected error on empty userID")
	}
	if err := svc.DeleteNotification(ctx, "", "usr-1"); err == nil {
		t.Fatalf("expected error on empty recipientID")
	}
}

func TestNotificationService_CRUD_IngestionAndListing(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	svc := services.NewNotificationService(repo)
	ctx := context.Background()

	req := dto.SendNotificationRequestDTO{
		SourceService:    "education",
		EventType:        "educacion.inscripcion.creada",
		Category:         "education",
		Priority:         "high",
		SenderType:       "system",
		SenderID:         "system",
		RecipientUserIDs: []string{"usr-student1", "usr-student2"},
		Title:            "Inscripción Confirmada",
		Message:          "Tu inscripción al curso de Angular ha sido aprobada.",
		Data:             `{"course_id":"crs-101"}`,
		IsCritical:       false,
		SendInApp:        true,
		SendEmail:        true,
	}

	notifID, count, err := svc.IngestNotification(ctx, req)
	if err != nil {
		t.Fatalf("failed to ingest notification: %v", err)
	}
	if notifID == "" || count != 2 {
		t.Fatalf("expected 2 recipients created, got %d", count)
	}

	list1, err := svc.GetNotifications(ctx, "usr-student1", 10, 0)
	if err != nil || len(list1) != 1 {
		t.Fatalf("expected 1 notification for student1, got %d (err: %v)", len(list1), err)
	}
	if list1[0].Title != req.Title || list1[0].IsRead {
		t.Fatalf("unexpected notification item properties")
	}

	detail, err := svc.GetNotificationDetail(ctx, list1[0].RecipientID, "usr-student1")
	if err != nil {
		t.Fatalf("failed to get notification detail: %v", err)
	}
	if detail.Title != req.Title || detail.Data["course_id"] != "crs-101" {
		t.Fatalf("detail data not properly unmarshaled: %v", detail.Data)
	}
}

func TestNotificationService_CRUD_ReadStatusAndDeletion(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	svc := services.NewNotificationService(repo)
	ctx := context.Background()

	req := dto.SendNotificationRequestDTO{
		SourceService:    "education",
		EventType:        "educacion.inscripcion.creada",
		Category:         "education",
		Priority:         "high",
		SenderType:       "system",
		SenderID:         "system",
		RecipientUserIDs: []string{"usr-student1", "usr-student2"},
		Title:            "Inscripción Confirmada",
		Message:          "Tu inscripción al curso de Angular ha sido aprobada.",
		IsCritical:       false,
		SendInApp:        true,
	}

	_, _, err := svc.IngestNotification(ctx, req)
	if err != nil {
		t.Fatalf("failed to ingest notification: %v", err)
	}

	list1, err := svc.GetNotifications(ctx, "usr-student1", 10, 0)
	if err != nil || len(list1) == 0 {
		t.Fatalf("expected notifications for student1")
	}
	recipientID := list1[0].RecipientID

	unread, err := svc.GetUnreadCount(ctx, "usr-student1")
	if err != nil || unread != 1 {
		t.Fatalf("expected unread count 1, got %d", unread)
	}

	if err := svc.MarkAsRead(ctx, recipientID, "usr-student1"); err != nil {
		t.Fatalf("failed to mark as read: %v", err)
	}

	unreadAfter, err := svc.GetUnreadCount(ctx, "usr-student1")
	if err != nil || unreadAfter != 0 {
		t.Fatalf("expected unread count 0, got %d", unreadAfter)
	}

	if err := svc.MarkAllAsRead(ctx, "usr-student2"); err != nil {
		t.Fatalf("failed to mark all as read: %v", err)
	}

	if err := svc.DeleteNotification(ctx, recipientID, "usr-student1"); err != nil {
		t.Fatalf("failed to delete notification: %v", err)
	}

	if _, err = svc.GetNotificationDetail(ctx, recipientID, "usr-student1"); err == nil {
		t.Fatalf("expected error getting detail for deleted notification, got nil")
	}
}

func TestNotificationService_CRUD_CompatibilityFindByUserID(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	svc := services.NewNotificationService(repo)
	ctx := context.Background()

	compat, err := svc.FindByUserID(ctx, "usr-legacy")
	if err != nil || len(compat) == 0 {
		t.Fatalf("expected legacy notifications fallback, got err: %v", err)
	}
}

func TestNotificationService_Preferences_Defaults(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	svc := services.NewNotificationService(repo)
	ctx := context.Background()

	prefs, err := svc.GetPreferences(ctx, "usr-pref-1")
	if err != nil || len(prefs) != 5 {
		t.Fatalf("expected 5 default preferences, got %d (err: %v)", len(prefs), err)
	}
	for _, p := range prefs {
		if !p.EmailEnabled || !p.PushEnabled || !p.InAppEnabled {
			t.Fatalf("expected all channels enabled by default for %s", p.Category)
		}
	}

	persisted, err := repo.GetPreferencesByUserID(ctx, "usr-pref-1")
	if err != nil || len(persisted) != 5 {
		t.Fatalf("expected 5 preferences persisted in repo, got %d", len(persisted))
	}
}

func TestNotificationService_Preferences_InitDirectly(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	svc := services.NewNotificationService(repo)
	ctx := context.Background()

	if err := svc.InitDefaultPreferences(ctx, "usr-new-user"); err != nil {
		t.Fatalf("unexpected error initializing preferences: %v", err)
	}
	persisted, err := repo.GetPreferencesByUserID(ctx, "usr-new-user")
	if err != nil || len(persisted) != 5 {
		t.Fatalf("expected 5 persisted rows for new user, got %d", len(persisted))
	}
}

func TestNotificationService_Preferences_SaveAndGet(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	svc := services.NewNotificationService(repo)
	ctx := context.Background()

	emailOff := false
	saved, err := svc.SavePreference(ctx, "usr-pref-1", dto.UpdatePreferenceRequestDTO{
		Category:     "education",
		EmailEnabled: &emailOff,
	})
	if err != nil || saved.EmailEnabled {
		t.Fatalf("expected saved preference with EmailEnabled=false, got err: %v", err)
	}

	single, err := svc.GetPreferenceByCategory(ctx, "usr-pref-1", "education")
	if err != nil || single.EmailEnabled {
		t.Fatalf("expected single preference with EmailEnabled=false")
	}

	prefs, err := svc.GetPreferences(ctx, "usr-pref-1")
	if err != nil {
		t.Fatalf("failed to get preferences: %v", err)
	}
	var educationFound bool
	for _, p := range prefs {
		if p.Category == "education" && !p.EmailEnabled {
			educationFound = true
		}
	}
	if !educationFound {
		t.Fatalf("expected saved education preference with email disabled")
	}
}

func TestNotificationService_Preferences_Patch(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	svc := services.NewNotificationService(repo)
	ctx := context.Background()

	emailOff := false
	_, err := svc.SavePreference(ctx, "usr-pref-1", dto.UpdatePreferenceRequestDTO{
		Category:     "education",
		EmailEnabled: &emailOff,
	})
	if err != nil {
		t.Fatalf("failed to setup initial preference: %v", err)
	}

	pushOff := false
	patched, err := svc.PatchPreference(ctx, "usr-pref-1", "education", dto.PatchPreferenceRequestDTO{
		PushEnabled: &pushOff,
	})
	if err != nil || patched.PushEnabled || patched.EmailEnabled {
		t.Fatalf("expected patched preference with PushEnabled=false and preserved EmailEnabled=false")
	}
}

func TestNotificationService_Preferences_ValidationErrors(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	svc := services.NewNotificationService(repo)
	ctx := context.Background()

	if _, err := svc.GetPreferences(ctx, ""); err == nil {
		t.Fatalf("expected error on empty user ID")
	}
	if _, err := svc.GetPreferenceByCategory(ctx, "usr-1", "invalid_category"); err == nil {
		t.Fatalf("expected error on invalid category")
	}
	if _, err := svc.SavePreference(ctx, "usr-1", dto.UpdatePreferenceRequestDTO{Category: "crypto"}); err == nil {
		t.Fatalf("expected error on invalid category")
	}
	if _, err := svc.PatchPreference(ctx, "usr-1", "invalid", dto.PatchPreferenceRequestDTO{}); err == nil {
		t.Fatalf("expected error on invalid category")
	}
}

type testEmailMock struct {
	sentMessages []services.EmailMessage
}

func (m *testEmailMock) SendEmail(ctx context.Context, msg services.EmailMessage) (string, error) {
	m.sentMessages = append(m.sentMessages, msg)
	return "resend-msg-123", nil
}

type testPushMock struct {
	sentTokens    []string
	invalidTokens []string
}

func (m *testPushMock) SendMulticast(ctx context.Context, tokens []string, title, body string, data map[string]string) (*services.MulticastResult, error) {
	m.sentTokens = append(m.sentTokens, tokens...)
	return &services.MulticastResult{
		SuccessCount:  len(tokens) - len(m.invalidTokens),
		FailureCount:  len(m.invalidTokens),
		InvalidTokens: m.invalidTokens,
	}, nil
}

type testAuthMock struct {
	tokens         []string
	tokensByUser   map[string][]string
	revokedTokens  []string
	recipients     []clients.RecipientContact
	verifiedByUser map[string]bool
}

func (m *testAuthMock) GetUserDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	if m.tokensByUser != nil {
		return m.tokensByUser[userID], nil
	}
	return m.tokens, nil
}

func (m *testAuthMock) RevokeDeviceTokens(ctx context.Context, userID string, tokens []string) error {
	m.revokedTokens = append(m.revokedTokens, tokens...)
	return nil
}

func (m *testAuthMock) GetCampaignRecipients(ctx context.Context) ([]clients.RecipientContact, error) {
	return m.recipients, nil
}

func (m *testAuthMock) IsUserEmailVerified(ctx context.Context, userID string) (bool, error) {
	if m.verifiedByUser != nil {
		return m.verifiedByUser[userID], nil
	}
	return true, nil
}

func (m *testAuthMock) Close() error {
	return nil
}

func TestNotificationService_Dispatch_FCMTokenCleanupAndEmail(t *testing.T) {
	// Arrange
	repo := repositories.NewMemoryNotificationRepository()
	emailMock := &testEmailMock{}
	pushMock := &testPushMock{
		invalidTokens: []string{"bad-token-1"},
	}
	authMock := &testAuthMock{
		tokens: []string{"good-token-1", "bad-token-1"},
	}

	svc := services.NewNotificationService(
		repo,
		services.WithEmailService(emailMock),
		services.WithPushService(pushMock),
		services.WithAuthClient(authMock),
		services.WithSyncDispatch(true),
	)

	req := dto.SendNotificationRequestDTO{
		SourceService:    "auth",
		EventType:        "auth.security.alert",
		Category:         "security",
		Priority:         "critical",
		RecipientUserIDs: []string{"user-alice@example.com"},
		Title:            "Nuevo inicio de sesión",
		Message:          "Se ha detectado un inicio de sesión desde un nuevo dispositivo.",
		Data:             `{"recipient_name":"Alice","action_url":"https://buenvivir.org"}`,
		IsCritical:       true,
		SendEmail:        true,
		SendPush:         true,
	}

	// Act
	notifID, count, err := svc.IngestNotification(context.Background(), req)

	// Assert
	if err != nil {
		t.Fatalf("unexpected ingest error: %v", err)
	}
	if notifID == "" || count != 1 {
		t.Fatalf("expected 1 recipient, got %d", count)
	}

	// Give async worker brief instant to complete
	recipients, err := repo.FindRecipientsByUserID(context.Background(), "user-alice@example.com", 10, 0)
	if err != nil || len(recipients) == 0 {
		t.Fatalf("expected recipient found, got err: %v", err)
	}

	// Verify token cleanup was triggered for unregistered token
	if len(authMock.revokedTokens) != 1 || authMock.revokedTokens[0] != "bad-token-1" {
		t.Errorf("expected bad-token-1 to be revoked, got %v", authMock.revokedTokens)
	}

	// Verify email was sent with EmailSenderAlert due to security category
	if len(emailMock.sentMessages) != 1 {
		t.Fatalf("expected 1 email sent, got %d", len(emailMock.sentMessages))
	}
	sentEmail := emailMock.sentMessages[0]
	if sentEmail.SenderType != services.EmailSenderAlert {
		t.Errorf("expected EmailSenderAlert for security notification, got %s", sentEmail.SenderType)
	}
	if sentEmail.To[0] != "user-alice@example.com" {
		t.Errorf("expected recipient user-alice@example.com, got %s", sentEmail.To[0])
	}
}

func TestNotificationService_CampaignBroadcast_AdheresToPreferences(t *testing.T) {
	// Arrange
	repo := repositories.NewMemoryNotificationRepository()
	ctx := context.Background()

	// Preferences setup:
	// usr-all: both email and push enabled (default)
	// usr-no-email: email disabled, push enabled
	_ = repo.UpsertPreference(ctx, &models.NotificationPreference{
		ID:           "prf-1",
		UserID:       "usr-no-email",
		Category:     models.CategoryCampaigns,
		EmailEnabled: false,
		PushEnabled:  true,
		InAppEnabled: true,
	})
	// usr-no-push: email enabled, push disabled
	_ = repo.UpsertPreference(ctx, &models.NotificationPreference{
		ID:           "prf-2",
		UserID:       "usr-no-push",
		Category:     models.CategoryCampaigns,
		EmailEnabled: true,
		PushEnabled:  false,
		InAppEnabled: true,
	})
	// usr-none: both email and push disabled
	_ = repo.UpsertPreference(ctx, &models.NotificationPreference{
		ID:           "prf-3",
		UserID:       "usr-none",
		Category:     models.CategoryCampaigns,
		EmailEnabled: false,
		PushEnabled:  false,
		InAppEnabled: true,
	})

	emailMock := &testEmailMock{}
	pushMock := &testPushMock{}
	authMock := &testAuthMock{
		tokensByUser: map[string][]string{
			"usr-all":        {"fcm-token-all"},
			"usr-no-email":   {"fcm-token-no-email"},
			"usr-no-push":    {"fcm-token-push"},
			"usr-none":       {"fcm-token-none"},
			"usr-unverified": {"fcm-token-unverified"},
		},
		recipients: []clients.RecipientContact{
			{UserID: "usr-all", Email: "all@example.com", FirstName: "Ana", LastName: "López", IsEmailVerified: true},
			{UserID: "usr-no-email", Email: "noemail@example.com", FirstName: "Bob", LastName: "Marley", IsEmailVerified: true},
			{UserID: "usr-no-push", Email: "nopush@example.com", FirstName: "Carlos", LastName: "Gómez", IsEmailVerified: true},
			{UserID: "usr-none", Email: "none@example.com", FirstName: "Diana", LastName: "Prince", IsEmailVerified: true},
			{UserID: "usr-unverified", Email: "unverified@example.com", FirstName: "Eva", LastName: "Duarte", IsEmailVerified: false},
		},
	}

	svc := services.NewNotificationService(
		repo,
		services.WithEmailService(emailMock),
		services.WithPushService(pushMock),
		services.WithAuthClient(authMock),
		services.WithSyncDispatch(true),
	)

	req := dto.SendNotificationRequestDTO{
		SourceService: "campaigns",
		EventType:     "campanas.campana.creada",
		Category:      "campaigns",
		Priority:      "normal",
		TargetType:    "broadcast",
		Title:         "Nueva Campaña de Salud",
		Message:       "Se ha publicado una nueva campaña de vacunación comunitaria.",
		IsCritical:    false,
		SendEmail:     true,
		SendPush:      true,
		SendInApp:     true,
	}

	// Act
	notifID, count, err := svc.IngestNotification(ctx, req)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error during ingest: %v", err)
	}
	if notifID == "" || count != 5 {
		t.Fatalf("expected 5 recipients created, got count=%d", count)
	}

	// Verify emails sent: should only be usr-all and usr-no-push (unverified user must be skipped!)
	if len(emailMock.sentMessages) != 2 {
		t.Fatalf("expected exactly 2 emails sent according to preferences and verification, got %d", len(emailMock.sentMessages))
	}
	emailRecipients := map[string]bool{}
	for _, m := range emailMock.sentMessages {
		if len(m.To) > 0 {
			emailRecipients[m.To[0]] = true
		}
	}
	if !emailRecipients["all@example.com"] || !emailRecipients["nopush@example.com"] {
		t.Errorf("unexpected email recipients: %v", emailRecipients)
	}
	if emailRecipients["unverified@example.com"] {
		t.Errorf("unverified user received email but should have been skipped")
	}

	// Verify push tokens targeted: usr-all, usr-no-email, and usr-unverified
	if len(pushMock.sentTokens) != 3 {
		t.Fatalf("expected exactly 3 push tokens, got %d", len(pushMock.sentTokens))
	}
	pushTokens := map[string]bool{}
	for _, tk := range pushMock.sentTokens {
		pushTokens[tk] = true
	}
	if !pushTokens["fcm-token-all"] || !pushTokens["fcm-token-no-email"] || !pushTokens["fcm-token-unverified"] {
		t.Errorf("unexpected push tokens: %v", pushTokens)
	}
}

func TestNotificationService_EmailDispatch_RequiresEmailVerification(t *testing.T) {
	repo := repositories.NewMemoryNotificationRepository()
	emailMock := &testEmailMock{}
	authMock := &testAuthMock{
		verifiedByUser: map[string]bool{
			"usr-verified":   true,
			"usr-unverified": false,
		},
	}

	svc := services.NewNotificationService(
		repo,
		services.WithEmailService(emailMock),
		services.WithAuthClient(authMock),
		services.WithSyncDispatch(true),
	)

	t.Run("UnverifiedUser_EmailSkipped", func(t *testing.T) {
		req := dto.SendNotificationRequestDTO{
			SourceService:    "education",
			EventType:        "education.alert",
			Category:         "education",
			Priority:         "normal",
			RecipientUserIDs: []string{"usr-unverified"},
			Title:            "Curso disponible",
			Message:          "Nuevo curso disponible",
			Data:             `{"email":"unverified@test.com"}`,
			SendEmail:        true,
		}
		_, _, err := svc.IngestNotification(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(emailMock.sentMessages) != 0 {
			t.Fatalf("expected 0 emails sent for unverified user, got %d", len(emailMock.sentMessages))
		}
	})

	t.Run("VerifiedUser_EmailSent", func(t *testing.T) {
		emailMock.sentMessages = nil
		req := dto.SendNotificationRequestDTO{
			SourceService:    "education",
			EventType:        "education.alert",
			Category:         "education",
			Priority:         "normal",
			RecipientUserIDs: []string{"usr-verified"},
			Title:            "Curso disponible",
			Message:          "Nuevo curso disponible",
			Data:             `{"email":"verified@test.com"}`,
			SendEmail:        true,
		}
		_, _, err := svc.IngestNotification(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(emailMock.sentMessages) != 1 {
			t.Fatalf("expected 1 email sent for verified user, got %d", len(emailMock.sentMessages))
		}
	})

	t.Run("VerificationEmail_ExemptedAndSentEvenIfUnverified", func(t *testing.T) {
		emailMock.sentMessages = nil
		req := dto.SendNotificationRequestDTO{
			SourceService:    "auth",
			EventType:        "auth.email_verification",
			Category:         "security",
			Priority:         "high",
			RecipientUserIDs: []string{"usr-unverified"},
			Title:            "Verifica tu cuenta",
			Message:          "Haz clic para verificar",
			Data:             `{"email":"unverified@test.com","template_id":"email_verification"}`,
			SendEmail:        true,
		}
		_, _, err := svc.IngestNotification(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(emailMock.sentMessages) != 1 {
			t.Fatalf("expected verification email to be exempted and sent, got %d", len(emailMock.sentMessages))
		}
	})
}

