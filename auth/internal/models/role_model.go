package models

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
    RoleID      string    `gorm:"primaryKey" json:"role_id"`
    Name        string    `gorm:"not null;unique" json:"name"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
    DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}