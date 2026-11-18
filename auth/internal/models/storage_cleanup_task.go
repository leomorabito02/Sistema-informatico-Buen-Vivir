package models

import (
	"time"
)

const (
	CleanupStatusPending    = "pending"
	CleanupStatusProcessing = "processing"
	CleanupStatusCompleted  = "completed"
	CleanupStatusFailed     = "failed"
)

// StorageCleanupTask represents an outbox asynchronous task to purge orphaned objects from object storage.
type StorageCleanupTask struct {
	ID          string    `gorm:"primaryKey;size:36" json:"id"`
	ObjectName  string    `gorm:"not null;size:512;index" json:"object_name"`
	Status      string    `gorm:"not null;size:20;default:'pending';index" json:"status"`
	Attempts    int       `gorm:"not null;default:0" json:"attempts"`
	MaxAttempts int       `gorm:"not null;default:5" json:"max_attempts"`
	LastError   string    `gorm:"type:text" json:"last_error,omitempty"`
	NextRunAt   time.Time `gorm:"not null;index" json:"next_run_at"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
