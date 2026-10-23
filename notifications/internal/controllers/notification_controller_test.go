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

	// Ingest via gRPC
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

	// Fetch recipient ID
	list, err := svc.GetNotifications(ctx, "usr-test1", 10, 0)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 notification in service, got %d", len(list))
	}
	rcpID := list[0].RecipientID

	t.Run("HealthCheck", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("GetNotifications_Authorized", func(t *testing.T) {
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
	})

	t.Run("GetNotificationDetail_Authorized", func(t *testing.T) {
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
	})

	t.Run("MarkAsRead_Authorized", func(t *testing.T) {
		req := httptest.NewRequest("PATCH", "/api/v1/notifications/"+rcpID+"/read", nil)
		req.Header.Set("X-User-ID", "usr-test1")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("GetUnreadCount_Authorized", func(t *testing.T) {
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
	})

	t.Run("MarkAllAsRead_Authorized", func(t *testing.T) {
		req := httptest.NewRequest("PATCH", "/api/v1/notifications/read-all", nil)
		req.Header.Set("X-User-ID", "usr-test1")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("DeleteNotification_Authorized", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/notifications/"+rcpID, nil)
		req.Header.Set("X-User-ID", "usr-test1")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected status 204 No Content, got %d", rec.Code)
		}

		// Verify deletion
		reqAfter := httptest.NewRequest("GET", "/api/v1/notifications/"+rcpID, nil)
		reqAfter.Header.Set("X-User-ID", "usr-test1")
		recAfter := httptest.NewRecorder()
		mux.ServeHTTP(recAfter, reqAfter)

		if recAfter.Code != http.StatusNotFound {
			t.Fatalf("expected status 404 after deletion, got %d", recAfter.Code)
		}
	})
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

	t.Run("GetPreferences_Unauthorized", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/notifications/preferences", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401, got %d", rec.Code)
		}
	})

	t.Run("GetPreferences_Defaults", func(t *testing.T) {
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
		if len(prefs) != 5 {
			t.Fatalf("expected 5 default preferences, got %d", len(prefs))
		}
	})

	t.Run("SavePreferences_Single_POST", func(t *testing.T) {
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
	})

	t.Run("SavePreferences_Batch_POST", func(t *testing.T) {
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
	})

	t.Run("PatchPreference_QueryCategory", func(t *testing.T) {
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
	})

	t.Run("GetPreferenceByCategory", func(t *testing.T) {
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
	})

	t.Run("SavePreferences_InvalidCategory", func(t *testing.T) {
		body := `{"category":"unknown_service"}`
		req := httptest.NewRequest("POST", "/api/v1/notifications/preferences", strings.NewReader(body))
		req.Header.Set("X-User-ID", "usr-pref-test")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 for invalid category, got %d", rec.Code)
		}
	})

	t.Run("InitDefaultPreferences_POST", func(t *testing.T) {
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
		if len(list) != 5 {
			t.Fatalf("expected 5 initialized preferences, got %d", len(list))
		}
	})
}

// TestGrpcNotificationController_InitPreferences tests default preferences initialization via gRPC.
func TestGrpcNotificationController_InitPreferences(t *testing.T) {
	_, grpcCtrl, svc := setupTestApp()
	ctx := context.Background()

	res, err := grpcCtrl.InitDefaultPreferences(ctx, &notificationpb.InitDefaultPreferencesRequest{
		UserId: "usr-grpc-pref",
	})
	if err != nil || !res.Success || res.Count != 5 {
		t.Fatalf("failed to init preferences via gRPC: %v", err)
	}

	prefs, err := svc.GetPreferences(ctx, "usr-grpc-pref")
	if err != nil || len(prefs) != 5 {
		t.Fatalf("expected 5 preferences in service, got %d", len(prefs))
	}
}
