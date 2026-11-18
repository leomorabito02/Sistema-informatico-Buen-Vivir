package models

import (
	"time"

	"gorm.io/gorm"
)

// NotificationRecipient represents the fanout delivery receipt and read state for an individual user.
type NotificationRecipient struct {
	RecipientID    string  `gorm:"primaryKey;size:64" json:"recipient_id"`
	NotificationID string  `gorm:"size:64;index;not null" json:"notification_id"`
	UserID         string  `gorm:"size:64;index;not null" json:"user_id"`
	RoleID         *string `gorm:"size:64;index" json:"role_id,omitempty"`

	IsRead bool       `gorm:"default:false;index" json:"is_read"`
	ReadAt *time.Time `json:"read_at,omitempty"`

	EmailStatus    DeliveryStatus `gorm:"size:20;default:'not_requested'" json:"email_status"`
	PushStatus     DeliveryStatus `gorm:"size:20;default:'not_requested'" json:"push_status"`
	EmailMessageID *string        `gorm:"size:128" json:"email_message_id,omitempty"`
	PushMessageID  *string        `gorm:"size:128" json:"push_message_id,omitempty"`
	ErrorMessage   *string        `gorm:"type:text" json:"error_message,omitempty"`
	Attempts       int            `gorm:"default:0" json:"attempts"`
	NextAttemptAt  *time.Time     `gorm:"index" json:"next_attempt_at,omitempty"`
	DedupeKey      string         `gorm:"size:128;index" json:"dedupe_key,omitempty"`

	Notification *Notification `gorm:"foreignKey:NotificationID" json:"notification,omitempty"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
