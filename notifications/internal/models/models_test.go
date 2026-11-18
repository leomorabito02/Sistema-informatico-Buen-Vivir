package models_test

import (
	"testing"
	"time"

	"buenvivir/notifications/internal/models"
)

func TestNotificationModel_Instantiation(t *testing.T) {
	// Arrange
	targetRole := "role_admin"
	now := time.Now()

	// Act
	notif := &models.Notification{
		ID:            "ntf-100",
		SourceService: "auth",
		EventType:     "auth.password.changed",
		Category:      models.CategorySecurity,
		Priority:      models.PriorityCritical,
		SenderType:    models.SenderTypeSystem,
		SenderID:      "system",
		TargetType:    models.TargetTypeRole,
		TargetRole:    &targetRole,
		Title:         "Contraseña modificada",
		Message:       "Se modificó la contraseña de su cuenta.",
		IsCritical:    true,
		SendEmail:     true,
		SendPush:      true,
		SendInApp:     true,
		CreatedAt:     now,
	}

	// Assert
	if notif.ID != "ntf-100" {
		t.Errorf("expected ID 'ntf-100', got %s", notif.ID)
	}
	if notif.SourceService != "auth" {
		t.Errorf("expected SourceService 'auth', got %s", notif.SourceService)
	}
	if notif.EventType != "auth.password.changed" {
		t.Errorf("expected EventType 'auth.password.changed', got %s", notif.EventType)
	}
	if notif.Category != models.CategorySecurity {
		t.Errorf("expected CategorySecurity, got %s", notif.Category)
	}
	if notif.Priority != models.PriorityCritical {
		t.Errorf("expected PriorityCritical, got %s", notif.Priority)
	}
	if notif.SenderType != models.SenderTypeSystem || notif.SenderID != "system" {
		t.Errorf("expected sender system/system, got %s/%s", notif.SenderType, notif.SenderID)
	}
	if notif.TargetType != models.TargetTypeRole || notif.TargetRole == nil || *notif.TargetRole != targetRole {
		t.Errorf("expected TargetRole 'role_admin', got %v", notif.TargetRole)
	}
	if notif.Title != "Contraseña modificada" || notif.Message != "Se modificó la contraseña de su cuenta." {
		t.Errorf("unexpected Title or Message")
	}
	if !notif.IsCritical || !notif.SendEmail || !notif.SendPush || !notif.SendInApp {
		t.Errorf("expected flags to be true")
	}
	if notif.CreatedAt != now {
		t.Errorf("expected CreatedAt %v, got %v", now, notif.CreatedAt)
	}
}

func TestNotificationPreference_AllowsChannel_NormalNotification(t *testing.T) {
	// Arrange
	pref := &models.NotificationPreference{
		ID:           "pref-1",
		UserID:       "usr-1",
		Category:     models.CategoryEducation,
		EmailEnabled: false,
		PushEnabled:  true,
		InAppEnabled: true,
	}

	// Act & Assert (Normal non-critical notification)
	if pref.AllowsChannel("email", false) {
		t.Errorf("expected email to be disabled by user preferences")
	}
	if !pref.AllowsChannel("push", false) {
		t.Errorf("expected push to be enabled by user preferences")
	}
	if !pref.AllowsChannel("in_app", false) {
		t.Errorf("expected in_app to be enabled by user preferences")
	}
}

func TestNotificationPreference_AllowsChannel_NingunaOmitted(t *testing.T) {
	// Arrange (Category not security, normal notification, both channels false)
	pref := &models.NotificationPreference{
		ID:           "pref-none",
		UserID:       "usr-none",
		Category:     models.CategoryCampaigns,
		EmailEnabled: false,
		PushEnabled:  false,
		InAppEnabled: true,
	}

	// Act & Assert (Neither email nor push enabled -> in_app omitted unless critical)
	if pref.AllowsChannel("email", false) {
		t.Errorf("expected email to be disabled")
	}
	if pref.AllowsChannel("push", false) {
		t.Errorf("expected push to be disabled")
	}
	if pref.AllowsChannel("in_app", false) {
		t.Errorf("expected in_app to be omitted when Ninguna (email and push false)")
	}
	// Critical still allows
	if !pref.AllowsChannel("in_app", true) {
		t.Errorf("expected critical notification to bypass Ninguna preference")
	}
}

func TestNotificationPreference_AllowsChannel_CriticalBypass(t *testing.T) {
	// Arrange
	pref := &models.NotificationPreference{
		ID:           "pref-2",
		UserID:       "usr-2",
		Category:     models.CategorySecurity,
		EmailEnabled: false,
		PushEnabled:  false,
		InAppEnabled: false,
	}

	// Act & Assert (Critical notification overrides user preferences)
	if !pref.AllowsChannel("email", true) {
		t.Errorf("expected critical notification to bypass email preference")
	}
	if !pref.AllowsChannel("push", true) {
		t.Errorf("expected critical notification to bypass push preference")
	}
	if !pref.AllowsChannel("in_app", true) {
		t.Errorf("expected critical notification to bypass in_app preference")
	}
}

func TestNotificationRecipient_Instantiation(t *testing.T) {
	// Arrange
	msgID := "re_resend_999"

	// Act
	rcp := &models.NotificationRecipient{
		RecipientID:    "rcp-1",
		NotificationID: "ntf-1",
		UserID:         "usr-1",
		IsRead:         false,
		EmailStatus:    models.DeliveryStatusSent,
		EmailMessageID: &msgID,
	}

	// Assert
	if rcp.RecipientID != "rcp-1" {
		t.Errorf("expected RecipientID 'rcp-1', got %s", rcp.RecipientID)
	}
	if rcp.NotificationID != "ntf-1" {
		t.Errorf("expected NotificationID 'ntf-1', got %s", rcp.NotificationID)
	}
	if rcp.UserID != "usr-1" {
		t.Errorf("expected UserID 'usr-1', got %s", rcp.UserID)
	}
	if rcp.IsRead {
		t.Errorf("expected IsRead false")
	}
	if rcp.EmailStatus != models.DeliveryStatusSent {
		t.Errorf("expected DeliveryStatusSent, got %s", rcp.EmailStatus)
	}
	if rcp.EmailMessageID == nil || *rcp.EmailMessageID != msgID {
		t.Errorf("expected EmailMessageID %s, got %v", msgID, rcp.EmailMessageID)
	}
}
