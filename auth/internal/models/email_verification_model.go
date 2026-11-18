package models

import "time"

type EmailVerification struct {
	EmailVerificationID string     `json:"email_verification_id" gorm:"primaryKey"`
	UserID              string     `json:"user_id" gorm:"not null;index"`
	User                *User      `json:"user,omitempty"`
	TokenHash           string     `json:"token_hash" gorm:"not null;index"`
	ExpiresAt           time.Time  `json:"expires_at" gorm:"not null"`
	UsedAt              *time.Time `json:"used_at"`
	CreatedAt           time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt           time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}
