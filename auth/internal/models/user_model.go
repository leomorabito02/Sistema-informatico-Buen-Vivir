package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	UserID         string    `gorm:"primaryKey" json:"user_id"`
	FirstName      string    `json:"first_name" gorm:"not null"`
	LastName       string    `json:"last_name" gorm:"not null"`
	Email          string    `json:"email" gorm:"not null;unique"`
	PasswordHash   string    `json:"-"`
	Phone          string    `json:"phone" gorm:"not null"`
	Status         string    `json:"status" gorm:"not null"`
	GoogleAuthID   string    `json:"google_auth_id"`
	RoleID         string    `json:"role_id" gorm:"index;not null"`
	Role               *Role     `json:"role,omitempty"`
	RoleApprovalStatus string    `json:"role_approval_status" gorm:"->;-:migration"`
	ImageURL           string    `json:"image_url" gorm:"->;-:migration"`
	TwofaEnabled       bool      `json:"twofa_enabled" gorm:"not null"`
	TwofaSecret    string    `json:"-" gorm:"not null"`
	TwofaSeed      string    `json:"-" gorm:"not null"`
	IsEmailVerified bool      `json:"is_email_verified" gorm:"default:false;not null"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}
