package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"buenvivir/notifications/api/proto/notificationpb"
	"buenvivir/notifications/internal/controllers"
	"buenvivir/notifications/internal/dto"
	"buenvivir/notifications/internal/repositories"
	"buenvivir/notifications/internal/routes"
	"buenvivir/notifications/internal/services"
)

func setupTestApp() (*http.ServeMux, *controllers.GrpcNotificationController, services.NotificationService) {
	repo := repositories.NewMemoryNotificationRepository()
	svc := services.NewNotificationService(repo)
	httpCtrl := controllers.NewNotificationController(svc)
	grpcCtrl := controllers.NewGrpcNotificationController(svc)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, httpCtrl)

	return mux, grpcCtrl, svc
}

func TestNotificationEndpoints_Flow(t *testing.T) {
	mux, grpcCtrl, svc := setupTestApp()
	ctx := context.Background()

	rcpID := seedFlowNotification(t, grpcCtrl, svc, ctx)

	t.Run("HealthCheck", func(t *testing.T) { testHealthCheck(t, mux) })
	t.Run("GetNotifications_Authorized", func(t *testing.T) { testGetNotifications(t, mux) })
	t.Run("GetNotificationDetail_Authorized", func(t *testing.T) { testGetNotificationDetail(t, mux, rcpID) })
	t.Run("MarkAsRead_Authorized", func(t *testing.T) { testMarkAsRead(t, mux, rcpID) })
	t.Run("GetUnreadCount_Authorized", func(t *testing.T) { testGetUnreadCount(t, mux) })
	t.Run("MarkAllAsRead_Authorized", func(t *testing.T) { testMarkAllAsRead(t, mux) })
	t.Run("DeleteNotification_Authorized", func(t *testing.T) { testDeleteNotification(t, mux, rcpID) })
}

func seedFlowNotification(t *testing.T, grpcCtrl *controllers.GrpcNotificationController, svc services.NotificationService, ctx context.Context) string {
	t.Helper()
	grpcRes, err := grpcCtrl.SendNotification(ctx, &notificationpb.SendNotificationRequest{
		SourceService:    "education",
		EventType:        "educacion.inscripcion.creada",
		Category:         "education",
		Priority:         "normal",
		SenderType:       "system",
		SenderId:         "system",
		RecipientUserIds: []string{"usr-test1"},
		Title:            "Curso Nuevo",
		Message:          "Bienvenido al curso",
		DataJson:         `{"id":"1"}`,
		SendInApp:        true,
	})
	if err != nil || grpcRes.RecipientsCount != 1 {
		t.Fatalf("failed to send gRPC notification: %v", err)
	}

	list, err := svc.GetNotifications(ctx, "usr-test1", 10, 0)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 notification in service, got %d", len(list))
	}
	return list[0].RecipientID
}

func testHealthCheck(t *testing.T, mux http.Handler) {
	t.Helper()
	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func testGetNotifications(t *testing.T, mux http.Handler) {
	t.Helper()
	req := httptest.NewRequest("GET", "/api/v1/notifications?limit=10", nil)
	req.Header.Set("X-User-ID", "usr-test1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var items []dto.NotificationResponseDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Curso Nuevo" {
		t.Fatalf("unexpected items returned: %v", items)
	}
}

func testGetNotificationDetail(t *testing.T, mux http.Handler, rcpID string) {
	t.Helper()
	req := httptest.NewRequest("GET", "/api/v1/notifications/"+rcpID, nil)
	req.Header.Set("X-User-ID", "usr-test1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var detail dto.NotificationDetailDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &detail); err != nil {
		t.Fatalf("failed to parse detail JSON: %v", err)
	}
	if detail.Title != "Curso Nuevo" || detail.Data["id"] != "1" {
		t.Fatalf("unexpected detail: %v", detail)
	}
}

func testMarkAsRead(t *testing.T, mux http.Handler, rcpID string) {
	t.Helper()
	req := httptest.NewRequest("PATCH", "/api/v1/notifications/"+rcpID+"/read", nil)
	req.Header.Set("X-User-ID", "usr-test1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func testGetUnreadCount(t *testing.T, mux http.Handler) {
	t.Helper()
	req := httptest.NewRequest("GET", "/api/v1/notifications/unread-count", nil)
	req.Header.Set("X-User-ID", "usr-test1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res dto.UnreadCountResponseDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse unread count: %v", err)
	}
	if res.UnreadCount != 0 {
		t.Fatalf("expected unread count 0, got %d", res.UnreadCount)
	}
}

func testMarkAllAsRead(t *testing.T, mux http.Handler) {
	t.Helper()
	req := httptest.NewRequest("PATCH", "/api/v1/notifications/read-all", nil)
	req.Header.Set("X-User-ID", "usr-test1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func testDeleteNotification(t *testing.T, mux http.Handler, rcpID string) {
	t.Helper()
	req := httptest.NewRequest("DELETE", "/api/v1/notifications/"+rcpID, nil)
	req.Header.Set("X-User-ID", "usr-test1")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 No Content, got %d", rec.Code)
	}

	reqAfter := httptest.NewRequest("GET", "/api/v1/notifications/"+rcpID, nil)
	reqAfter.Header.Set("X-User-ID", "usr-test1")
	recAfter := httptest.NewRecorder()
	mux.ServeHTTP(recAfter, reqAfter)

	if recAfter.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 after deletion, got %d", recAfter.Code)
	}
}

func TestGrpcNotificationController_Batch(t *testing.T) {
	_, grpcCtrl, _ := setupTestApp()
	ctx := context.Background()

	batchRes, err := grpcCtrl.SendBatchNotifications(ctx, &notificationpb.SendBatchNotificationsRequest{
		Notifications: []*notificationpb.SendNotificationRequest{
			{
				SourceService:    "auth",
				EventType:        "auth.user.approved",
				RecipientUserIds: []string{"usr-batch1"},
				Title:            "Usuario Aprobado",
				Message:          "Tu cuenta ha sido aprobada.",
			},
			{
				SourceService:    "auth",
				EventType:        "auth.password.changed",
				RecipientUserIds: []string{"usr-batch2"},
				Title:            "Contraseña Cambiada",
				Message:          "Tu clave se actualizó con éxito.",
			},
		},
	})

	if err != nil {
		t.Fatalf("failed batch notification: %v", err)
	}
	if batchRes.CreatedCount != 2 || len(batchRes.NotificationIds) != 2 {
		t.Fatalf("expected 2 notifications created, got %d", batchRes.CreatedCount)
	}
}

func TestNotificationEndpoints_PreferencesFlow(t *testing.T) {
	mux, _, _ := setupTestApp()

	t.Run("GetPreferences_Unauthorized", func(t *testing.T) { testGetPreferencesUnauthorized(t, mux) })
	t.Run("GetPreferences_Defaults", func(t *testing.T) { testGetPreferencesDefaults(t, mux) })
	t.Run("SavePreferences_Single_POST", func(t *testing.T) { testSavePreferencesSingle(t, mux) })
	t.Run("SavePreferences_Batch_POST", func(t *testing.T) { testSavePreferencesBatch(t, mux) })
	t.Run("PatchPreference_QueryCategory", func(t *testing.T) { testPatchPreferenceQueryCategory(t, mux) })
	t.Run("GetPreferenceByCategory", func(t *testing.T) { testGetPreferenceByCategory(t, mux) })
	t.Run("SavePreferences_InvalidCategory", func(t *testing.T) { testSavePreferencesInvalidCategory(t, mux) })
	t.Run("InitDefaultPreferences_POST", func(t *testing.T) { testInitDefaultPreferencesPOST(t, mux) })
	t.Run("InitDefaultPreferences_POST_WithReset", func(t *testing.T) { testInitDefaultPreferencesWithReset(t, mux) })
}

func testGetPreferencesUnauthorized(t *testing.T, mux http.Handler) {
	t.Helper()
	req := httptest.NewRequest("GET", "/api/v1/notifications/preferences", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func testGetPreferencesDefaults(t *testing.T, mux http.Handler) {
	t.Helper()
	req := httptest.NewRequest("GET", "/api/v1/notifications/preferences", nil)
	req.Header.Set("X-User-ID", "usr-pref-test")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var prefs []dto.NotificationPreferenceDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &prefs); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(prefs) != 6 {
		t.Fatalf("expected 6 default preferences, got %d", len(prefs))
	}
}

func testSavePreferencesSingle(t *testing.T, mux http.Handler) {
	t.Helper()
	body := `{"category":"education","email_enabled":false,"push_enabled":true,"in_app_enabled":true}`
	req := httptest.NewRequest("POST", "/api/v1/notifications/preferences", strings.NewReader(body))
	req.Header.Set("X-User-ID", "usr-pref-test")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var res dto.NotificationPreferenceDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if res.Category != "education" || res.EmailEnabled {
		t.Fatalf("expected education preference with email disabled")
	}
}

func testSavePreferencesBatch(t *testing.T, mux http.Handler) {
	t.Helper()
	body := `[{"category":"labor","email_enabled":false},{"category":"campaigns","push_enabled":false}]`
	req := httptest.NewRequest("POST", "/api/v1/notifications/preferences", strings.NewReader(body))
	req.Header.Set("X-User-ID", "usr-pref-test")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var list []dto.NotificationPreferenceDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 2 {
		t.Fatalf("expected 2 saved preferences, got %d", len(list))
	}
}

func testPatchPreferenceQueryCategory(t *testing.T, mux http.Handler) {
	t.Helper()
	body := `{"push_enabled":false}`
	req := httptest.NewRequest("PATCH", "/api/v1/notifications/preferences?category=education", bytes.NewBufferString(body))
	req.Header.Set("X-User-ID", "usr-pref-test")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var res dto.NotificationPreferenceDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if res.PushEnabled || res.EmailEnabled {
		t.Fatalf("expected PushEnabled=false and preserved EmailEnabled=false")
	}
}

func testGetPreferenceByCategory(t *testing.T, mux http.Handler) {
	t.Helper()
	req := httptest.NewRequest("GET", "/api/v1/notifications/preferences/education", nil)
	req.Header.Set("X-User-ID", "usr-pref-test")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	var res dto.NotificationPreferenceDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &res)
	if res.Category != "education" {
		t.Fatalf("expected education category, got %s", res.Category)
	}
}

func testSavePreferencesInvalidCategory(t *testing.T, mux http.Handler) {
	t.Helper()
	body := `{"category":"unknown_service"}`
	req := httptest.NewRequest("POST", "/api/v1/notifications/preferences", strings.NewReader(body))
	req.Header.Set("X-User-ID", "usr-pref-test")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid category, got %d", rec.Code)
	}
}

func testInitDefaultPreferencesPOST(t *testing.T, mux http.Handler) {
	t.Helper()
	body := `{"user_id":"usr-init-test"}`
	req := httptest.NewRequest("POST", "/api/v1/notifications/preferences/init", strings.NewReader(body))
	req.Header.Set("X-User-ID", "usr-init-test")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rec.Code)
	}
	var list []dto.NotificationPreferenceDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 6 {
		t.Fatalf("expected 6 initialized preferences, got %d", len(list))
	}
}

func testInitDefaultPreferencesWithReset(t *testing.T, mux http.Handler) {
	t.Helper()
	body := `{"user_id":"usr-init-test","reset":true}`
	req := httptest.NewRequest("POST", "/api/v1/notifications/preferences/init", strings.NewReader(body))
	req.Header.Set("X-User-ID", "usr-init-test")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rec.Code)
	}
	var list []dto.NotificationPreferenceDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	if len(list) != 6 {
		t.Fatalf("expected 6 reset preferences, got %d", len(list))
	}
	for _, p := range list {
		if !p.EmailEnabled || !p.PushEnabled {
			t.Fatalf("expected all channels true after reset for category %s", p.Category)
		}
	}
}

// TestGrpcNotificationController_InitPreferences tests default preferences initialization via gRPC.
func TestGrpcNotificationController_InitPreferences(t *testing.T) {
	_, grpcCtrl, svc := setupTestApp()
	ctx := context.Background()

	res, err := grpcCtrl.InitDefaultPreferences(ctx, &notificationpb.InitDefaultPreferencesRequest{
		UserId: "usr-grpc-pref",
	})
	if err != nil || !res.Success || res.Count != 6 {
		t.Fatalf("failed to init preferences via gRPC: %v", err)
	}

	prefs, err := svc.GetPreferences(ctx, "usr-grpc-pref")
	if err != nil || len(prefs) != 6 {
		t.Fatalf("expected 6 preferences in service, got %d", len(prefs))
	}
}


