package models

import (
	"time"
)

const (
	OutboxStatusPending    = "pending"
	OutboxStatusProcessing = "processing"
	OutboxStatusCompleted  = "completed"
	OutboxStatusDead       = "dead"
)

// OutboxEvent represents an asynchronous domain event enqueued for publishing to Pub/Sub.
type OutboxEvent struct {
	ID          string     `gorm:"primaryKey;size:36" json:"id"`
	Topic       string     `gorm:"not null;size:100;index" json:"topic"`
	EventType   string     `gorm:"not null;size:100;index" json:"event_type"`
	Payload     []byte     `gorm:"type:jsonb;not null" json:"payload"`
	OrderingKey string     `gorm:"size:100" json:"ordering_key,omitempty"`
	Status      string     `gorm:"not null;size:20;default:'pending';index" json:"status"`
	Attempts    int        `gorm:"not null;default:0" json:"attempts"`
	MaxAttempts int        `gorm:"not null;default:5" json:"max_attempts"`
	LastError   string     `gorm:"type:text" json:"last_error,omitempty"`
	NextRunAt   time.Time  `gorm:"not null;index" json:"next_run_at"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

