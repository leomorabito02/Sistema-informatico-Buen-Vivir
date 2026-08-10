package models

import "time"

type PasswordRecovery struct {
	PasswordRecoveryID string    `json:"password_recovery_id" gorm:"primaryKey"`
	UserID             string    `gorm:"not null" json:"user_id"`
	User               *User      `json:"user,omitempty"`
	TokenHash          string     `json:"token_hash" gorm:"not null"`
	ExpiresAt          time.Time  `json:"expires_at" gorm:"not null"`
	UsedAt             *time.Time `json:"used_at"`
	CreatedAt          time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	
}
