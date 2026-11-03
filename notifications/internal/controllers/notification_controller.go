package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"buenvivir/notifications/internal/dto"
	"buenvivir/notifications/internal/middlewares"
	"buenvivir/notifications/internal/services"
)

type NotificationController struct {
	service services.NotificationService
}

const (
	errMissingNotificationID = "missing notification id"
	errInvalidRequestBody    = "invalid request body"
)

func NewNotificationController(service services.NotificationService) *NotificationController {
	return &NotificationController{service: service}
}

// IngestNotification handles internal HTTP ingestion of notifications.
func (c *NotificationController) IngestNotification(w http.ResponseWriter, r *http.Request) {
	var req dto.SendNotificationRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondError(w, http.StatusBadRequest, errInvalidRequestBody)
		return
	}
	id, count, err := c.service.IngestNotification(r.Context(), req)
	if err != nil {
		c.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	c.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"notification_id":  id,
		"recipients_count": count,
		"status":           "created",
	})
}

// GetNotifications returns the paginated list of notifications for the authenticated user.
func (c *NotificationController) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		c.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	items, err := c.service.GetNotifications(r.Context(), userID, limit, offset)
	if err != nil {
		c.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	c.respondJSON(w, http.StatusOK, items)
}

// GetNotificationDetail returns the complete details of a specific notification for the user.
func (c *NotificationController) GetNotificationDetail(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		c.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	recipientID := extractPathID(r)
	if recipientID == "" {
		c.respondError(w, http.StatusBadRequest, errMissingNotificationID)
		return
	}

	detail, err := c.service.GetNotificationDetail(r.Context(), recipientID, userID)
	if err != nil {
		c.respondError(w, http.StatusNotFound, "notification not found")
		return
	}

	c.respondJSON(w, http.StatusOK, detail)
}

// MarkAsRead marks an individual notification as read.
func (c *NotificationController) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		c.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	recipientID := extractPathID(r)
	if recipientID == "" {
		c.respondError(w, http.StatusBadRequest, errMissingNotificationID)
		return
	}

	if err := c.service.MarkAsRead(r.Context(), recipientID, userID); err != nil {
		c.respondError(w, http.StatusNotFound, "notification not found or update failed")
		return
	}

	c.respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// MarkAllAsRead marks all notifications for the user as read.
func (c *NotificationController) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		c.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if err := c.service.MarkAllAsRead(r.Context(), userID); err != nil {
		c.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	c.respondJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// DeleteNotification permanently hard-deletes a notification recipient record.
func (c *NotificationController) DeleteNotification(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		c.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	recipientID := extractPathID(r)
	if recipientID == "" {
		c.respondError(w, http.StatusBadRequest, errMissingNotificationID)
		return
	}

	if err := c.service.DeleteNotification(r.Context(), recipientID, userID); err != nil {
		c.respondError(w, http.StatusNotFound, "notification not found or delete failed")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUnreadCount returns the count of unread notifications for the user.
func (c *NotificationController) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		c.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	count, err := c.service.GetUnreadCount(r.Context(), userID)
	if err != nil {
		c.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	c.respondJSON(w, http.StatusOK, dto.UnreadCountResponseDTO{UnreadCount: count})
}

// HealthCheck returns service health status.
func (c *NotificationController) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	c.respondJSON(w, http.StatusOK, map[string]string{
		"status":  "UP",
		"service": "notifications-service",
	})
}

// GetPreferences returns all notification category preferences for the user.
func (c *NotificationController) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		c.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	prefs, err := c.service.GetPreferences(r.Context(), userID)
	if err != nil {
		c.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	c.respondJSON(w, http.StatusOK, prefs)
}

// GetPreferenceByCategory returns preferences for a single category.
func (c *NotificationController) GetPreferenceByCategory(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		c.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	category := extractPathCategory(r)
	if category == "" {
		c.respondError(w, http.StatusBadRequest, "category is required")
		return
	}

	pref, err := c.service.GetPreferenceByCategory(r.Context(), userID, category)
	if err != nil {
		c.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	c.respondJSON(w, http.StatusOK, pref)
}

// SavePreferences creates or updates category notification preferences (supports single or array).
func (c *NotificationController) SavePreferences(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		c.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		c.respondError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	trimmed := bytes.TrimSpace(body)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		c.handleSavePreferencesBatch(w, r, userID, trimmed)
		return
	}

	var req dto.UpdatePreferenceRequestDTO
	if err := json.Unmarshal(trimmed, &req); err != nil {
		c.respondError(w, http.StatusBadRequest, errInvalidRequestBody)
		return
	}

	res, err := c.service.SavePreference(r.Context(), userID, req)
	if err != nil {
		c.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	c.respondJSON(w, http.StatusOK, res)
}

func (c *NotificationController) handleSavePreferencesBatch(w http.ResponseWriter, r *http.Request, userID string, raw []byte) {
	var reqs []dto.UpdatePreferenceRequestDTO
	if err := json.Unmarshal(raw, &reqs); err != nil {
		c.respondError(w, http.StatusBadRequest, "invalid request body array")
		return
	}
	results := make([]dto.NotificationPreferenceDTO, 0, len(reqs))
	for _, req := range reqs {
		res, err := c.service.SavePreference(r.Context(), userID, req)
		if err != nil {
			c.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		results = append(results, *res)
	}
	c.respondJSON(w, http.StatusOK, results)
}

// PatchPreference partially updates channel settings for a category.
func (c *NotificationController) PatchPreference(w http.ResponseWriter, r *http.Request) {
	userID, err := middlewares.GetUserID(r.Context())
	if err != nil {
		c.respondError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req dto.PatchPreferenceRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondError(w, http.StatusBadRequest, errInvalidRequestBody)
		return
	}

	category := extractPathCategory(r)
	if category == "" {
		category = req.Category
	}
	if category == "" {
		c.respondError(w, http.StatusBadRequest, "category is required")
		return
	}

	res, err := c.service.PatchPreference(r.Context(), userID, category, req)
	if err != nil {
		c.respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	c.respondJSON(w, http.StatusOK, res)
}

// InitDefaultPreferences initializes or resets default preferences in DB for a user.
func (c *NotificationController) InitDefaultPreferences(w http.ResponseWriter, r *http.Request) {
	userID, _ := middlewares.GetUserID(r.Context())
	shouldReset := r.URL.Query().Get("reset") == "true"

	if r.Body != nil {
		var body struct {
			UserID string `json:"user_id"`
			Reset  bool   `json:"reset"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			if body.UserID != "" {
				userID = body.UserID
			}
			if body.Reset {
				shouldReset = true
			}
		}
	}

	if strings.TrimSpace(userID) == "" {
		c.respondError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	var err error
	if shouldReset {
		err = c.service.ResetPreferences(r.Context(), userID)
	} else {
		err = c.service.InitDefaultPreferences(r.Context(), userID)
	}
	if err != nil {
		c.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	prefs, err := c.service.GetPreferences(r.Context(), userID)
	if err != nil {
		c.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	c.respondJSON(w, http.StatusCreated, prefs)
}


func extractPathID(r *http.Request) string {
	id := r.PathValue("id")
	if id != "" {
		return id
	}
	// Fallback path trimming if standard mux pattern is used with trailing path
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 4 { // e.g., api/v1/notifications/{id}
		return parts[3]
	}
	return ""
}

func extractPathCategory(r *http.Request) string {
	cat := r.PathValue("category")
	if cat != "" {
		return cat
	}
	if q := r.URL.Query().Get("category"); q != "" {
		return q
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 5 && parts[3] == "preferences" {
		return parts[4]
	}
	return ""
}

func (c *NotificationController) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (c *NotificationController) respondError(w http.ResponseWriter, status int, message string) {
	c.respondJSON(w, status, map[string]string{"error": message})
}
