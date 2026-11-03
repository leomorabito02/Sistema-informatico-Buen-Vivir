package models

import (
	"time"

	"gorm.io/gorm"
)

// NotificationPreference stores per-user delivery channel settings by category.
type NotificationPreference struct {
	ID       string               `gorm:"primaryKey;size:64" json:"id"`
	UserID   string               `gorm:"size:64;not null;uniqueIndex:idx_user_category" json:"user_id"`
	Category NotificationCategory `gorm:"size:32;not null;uniqueIndex:idx_user_category" json:"category"`

	EmailEnabled bool `gorm:"default:true" json:"email_enabled"`
	PushEnabled  bool `gorm:"default:true" json:"push_enabled"`
	InAppEnabled bool `gorm:"default:true" json:"in_app_enabled"`

	CreatedAt time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// AllowsChannel returns true if the user allows delivery through the specified channel or if notification is critical/security.
func (p *NotificationPreference) AllowsChannel(channel string, isCritical bool) bool {
	if isCritical || p.Category == CategorySecurity {
		return true
	}
	switch channel {
	case "email":
		return p.EmailEnabled
	case "push":
		return p.PushEnabled
	case "in_app":
		return p.InAppEnabled && (p.EmailEnabled || p.PushEnabled)
	default:
		return false
	}
}
