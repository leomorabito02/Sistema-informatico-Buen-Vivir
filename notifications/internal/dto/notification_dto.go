package dto

import "time"

// NotificationResponseDTO represents the summarized notification item for UI listings.
type NotificationResponseDTO struct {
	RecipientID    string     `json:"recipient_id"`
	NotificationID string     `json:"notification_id"`
	Title          string     `json:"title"`
	Message        string     `json:"message"`
	Category       string     `json:"category"`
	Priority       string     `json:"priority"`
	SourceService  string     `json:"source_service"`
	ActionURL      *string    `json:"action_url,omitempty"`
	IsCritical     bool       `json:"is_critical"`
	IsRead         bool       `json:"is_read"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// NotificationDetailDTO represents the complete notification payload including entity metadata.
type NotificationDetailDTO struct {
	NotificationResponseDTO
	SenderType string                 `json:"sender_type"`
	SenderID   string                 `json:"sender_id"`
	Data       map[string]interface{} `json:"data,omitempty"`
}

// UnreadCountResponseDTO represents the count of unread notifications for a user.
type UnreadCountResponseDTO struct {
	UnreadCount int64 `json:"unread_count"`
}

// SendNotificationRequestDTO represents an ingestion request to trigger notification distribution.
type SendNotificationRequestDTO struct {
	SourceService    string   `json:"source_service"`
	EventType        string   `json:"event_type"`
	Category         string   `json:"category"`
	Priority         string   `json:"priority"`
	SenderType       string   `json:"sender_type"`
	SenderID         string   `json:"sender_id"`
	TargetType       string   `json:"target_type"`
	TargetRole       *string  `json:"target_role,omitempty"`
	RecipientUserIDs []string `json:"recipient_user_ids"`
	Title            string   `json:"title"`
	Message          string   `json:"message"`
	ActionURL        *string  `json:"action_url,omitempty"`
	Data             string   `json:"data,omitempty"`
	IsCritical       bool     `json:"is_critical"`
	SendEmail        bool     `json:"send_email"`
	SendPush         bool     `json:"send_push"`
	SendInApp        bool     `json:"send_in_app"`
}

// NotificationPreferenceDTO represents a user's notification channel preferences for a category.
type NotificationPreferenceDTO struct {
	Category     string `json:"category"`
	EmailEnabled bool   `json:"email_enabled"`
	PushEnabled  bool   `json:"push_enabled"`
	InAppEnabled bool   `json:"in_app_enabled"`
}

// UpdatePreferenceRequestDTO represents request to create or replace preferences for a category.
type UpdatePreferenceRequestDTO struct {
	Category     string `json:"category"`
	EmailEnabled *bool  `json:"email_enabled,omitempty"`
	PushEnabled  *bool  `json:"push_enabled,omitempty"`
	InAppEnabled *bool  `json:"in_app_enabled,omitempty"`
}

// PatchPreferenceRequestDTO allows partial updates to channel preferences for a category.
type PatchPreferenceRequestDTO struct {
	Category     string `json:"category,omitempty"`
	EmailEnabled *bool  `json:"email_enabled,omitempty"`
	PushEnabled  *bool  `json:"push_enabled,omitempty"`
	InAppEnabled *bool  `json:"in_app_enabled,omitempty"`
}

