package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PostgresProcessedStore implements events.ProcessedStore backed by PostgreSQL/GORM.
type PostgresProcessedStore struct {
	db          *gorm.DB
	memoryStore *events.MemoryStore
}

// NewProcessedStore creates a ProcessedStore using GORM if db is provided, or an in-memory store as fallback.
func NewProcessedStore(db *gorm.DB) events.ProcessedStore {
	if db == nil {
		return &PostgresProcessedStore{memoryStore: events.NewMemoryStore(1000)}
	}
	return &PostgresProcessedStore{db: db}
}

// IsProcessed checks if the event has already been marked as processed by the given consumer.
func (s *PostgresProcessedStore) IsProcessed(ctx context.Context, consumer, eventID string) (bool, error) {
	if s.memoryStore != nil {
		return s.memoryStore.IsProcessed(ctx, consumer, eventID)
	}

	var count int64
	err := s.db.WithContext(ctx).
		Model(&models.ProcessedEvent{}).
		Where("consumer = ? AND event_id = ?", consumer, eventID).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("failed to query processed event: %w", err)
	}
	return count > 0, nil
}

// MarkProcessed records that the event has been successfully processed.
func (s *PostgresProcessedStore) MarkProcessed(ctx context.Context, consumer, eventID, eventType string) error {
	if s.memoryStore != nil {
		return s.memoryStore.MarkProcessed(ctx, consumer, eventID, eventType)
	}

	record := models.ProcessedEvent{
		ID:          fmt.Sprintf("%s_%s", consumer, eventID),
		Consumer:    consumer,
		EventID:     eventID,
		EventType:   eventType,
		ProcessedAt: time.Now().UTC(),
	}

	// Upsert / On Conflict Do Nothing
	err := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&record).Error
	if err != nil && !errors.Is(err, gorm.ErrDuplicatedKey) {
		return fmt.Errorf("failed to record processed event: %w", err)
	}
	return nil
}

