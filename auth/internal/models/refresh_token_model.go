package models

import "time"

type RefreshToken struct {
	RefreshTokenID 		string     `json:"refresh_token_id" gorm:"primaryKey"`
	UserID         		string     `json:"user_id" gorm:"index"`
	TokenHash      		string     `json:"token_hash" gorm:"not null;uniqueIndex"`
	ExpiresAt          	time.Time  `json:"expires_at" gorm:"not null"`
	UsedAt             	*time.Time `json:"used_at"`
	RevokedAt          	*time.Time `json:"revoked_at"`
	CreatedAt          	time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          	time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	User               	*User      `json:"user,omitempty"`
}
