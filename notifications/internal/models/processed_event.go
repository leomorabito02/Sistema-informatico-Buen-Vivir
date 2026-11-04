package models

import "time"

// ProcessedEvent tracks events already consumed by Pub/Sub push endpoints for idempotency.
type ProcessedEvent struct {
	ID          string    `gorm:"primaryKey;size:64" json:"id"`
	Consumer    string    `gorm:"size:64;uniqueIndex:idx_consumer_event;not null" json:"consumer"`
	EventID     string    `gorm:"size:64;uniqueIndex:idx_consumer_event;not null" json:"event_id"`
	EventType   string    `gorm:"size:128;not null" json:"event_type"`
	ProcessedAt time.Time `gorm:"autoCreateTime" json:"processed_at"`
}

