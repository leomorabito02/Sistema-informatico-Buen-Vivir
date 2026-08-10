package models

import (
	"time"

	"gorm.io/gorm"
)

type Permission struct {
    PermissionID string    `gorm:"primaryKey" json:"permission_id"`
    Action       string    `gorm:"not null;uniqueIndex:idx_action_resource" json:"action"`
    Resource     string    `gorm:"not null;uniqueIndex:idx_action_resource" json:"resource"`
    Description  string    `json:"description"`
    CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}