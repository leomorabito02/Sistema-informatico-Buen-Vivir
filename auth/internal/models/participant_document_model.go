package models

import (
	"time"

	"gorm.io/gorm"
)

type ParticipantDocument struct {
	ID            string         `gorm:"primaryKey" json:"id"`
	ParticipantID string         `gorm:"not null;index" json:"participant_id"`
	Name          string         `gorm:"not null" json:"name"`
	DocURL        string         `gorm:"not null" json:"doc_url"`
	FileType      string         `json:"file_type"`
	FileSize      int64          `json:"file_size"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}
