package repositories

import (
	"context"
	"time"

	"auth/internal/models"

	"gorm.io/gorm"
)

// OutboxRepository manages persisting and querying asynchronous outbox events in PostgreSQL.
type OutboxRepository interface {
	Enqueue(ctx context.Context, event *models.OutboxEvent) error
	EnqueueTx(ctx context.Context, tx *gorm.DB, event *models.OutboxEvent) error
	GetDueEvents(ctx context.Context, limit int) ([]models.OutboxEvent, error)
	MarkProcessing(ctx context.Context, id string) error
	MarkCompleted(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, lastErr string, nextRunAt time.Time, dead bool) error
}

type outboxRepository struct {
	db *gorm.DB
}

// NewOutboxRepository constructs an OutboxRepository using GORM.
func NewOutboxRepository(db *gorm.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) Enqueue(ctx context.Context, event *models.OutboxEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *outboxRepository) EnqueueTx(ctx context.Context, tx *gorm.DB, event *models.OutboxEvent) error {
	if tx == nil {
		return r.Enqueue(ctx, event)
	}
	return tx.WithContext(ctx).Create(event).Error
}

func (r *outboxRepository) GetDueEvents(ctx context.Context, limit int) ([]models.OutboxEvent, error) {
	var events []models.OutboxEvent
	err := r.db.WithContext(ctx).
		Where("status IN ? AND next_run_at <= ?", []string{models.OutboxStatusPending, models.OutboxStatusProcessing}, time.Now()).
		Order("next_run_at ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *outboxRepository) MarkProcessing(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where(queryWhereID, id).
		Updates(map[string]any{
			"status":     models.OutboxStatusProcessing,
			"updated_at": time.Now(),
		}).Error
}

func (r *outboxRepository) MarkCompleted(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where(queryWhereID, id).
		Updates(map[string]any{
			"status":     models.OutboxStatusCompleted,
			"updated_at": time.Now(),
		}).Error
}

func (r *outboxRepository) MarkFailed(ctx context.Context, id string, lastErr string, nextRunAt time.Time, dead bool) error {
	status := models.OutboxStatusPending
	if dead {
		status = models.OutboxStatusDead
	}

	return r.db.WithContext(ctx).
		Model(&models.OutboxEvent{}).
		Where(queryWhereID, id).
		Updates(map[string]any{
			"status":      status,
			"last_error":  lastErr,
			"next_run_at": nextRunAt,
			"attempts":    gorm.Expr("attempts + 1"),
			"updated_at":  time.Now(),
		}).Error
}

