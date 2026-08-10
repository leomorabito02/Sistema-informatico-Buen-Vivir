package models

import "time"

type RolePermission struct {
    RoleID       string `gorm:"primaryKey" json:"role_id"`
    PermissionID string `gorm:"primaryKey" json:"permission_id"`

    Role       *Role       `json:"role,omitempty"`
    Permission *Permission `json:"permission,omitempty"`

    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
