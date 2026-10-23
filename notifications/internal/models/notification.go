package models

import (
	"time"

	"gorm.io/gorm"
)

// Notification represents a system or user-triggered notification event.
type Notification struct {
	ID            string               `gorm:"primaryKey;size:64" json:"id"`
	SourceService string               `gorm:"size:64;index;not null" json:"source_service"`
	EventType     string               `gorm:"size:128;index;not null" json:"event_type"`
	Category      NotificationCategory `gorm:"size:32;index;not null" json:"category"`
	Priority      NotificationPriority `gorm:"size:16;default:'normal'" json:"priority"`

	SenderType SenderType `gorm:"size:16;not null;default:'system'" json:"sender_type"`
	SenderID   string     `gorm:"size:64;index;not null" json:"sender_id"`

	TargetType TargetType `gorm:"size:16;not null;default:'individual'" json:"target_type"`
	TargetRole *string    `gorm:"size:64;index" json:"target_role,omitempty"`
	UserID     string     `gorm:"size:64;index" json:"user_id,omitempty"`

	Title     string  `gorm:"size:255;not null" json:"title"`
	Message   string  `gorm:"type:text;not null" json:"message"`
	ActionURL *string `gorm:"size:512" json:"action_url,omitempty"`
	Data      string  `gorm:"type:jsonb;default:'{}'" json:"data,omitempty"`

	IsCritical bool `gorm:"default:false;index" json:"is_critical"`
	SendEmail  bool `gorm:"default:false" json:"send_email"`
	SendPush   bool `gorm:"default:false" json:"send_push"`
	SendInApp  bool `gorm:"default:true" json:"send_in_app"`
	Read       bool `gorm:"default:false;index" json:"read"`

	Recipients []NotificationRecipient `gorm:"foreignKey:NotificationID;constraint:OnDelete:CASCADE" json:"recipients,omitempty"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
