package repositories

import (
	"context"
	"time"

	"auth/internal/models"

	"gorm.io/gorm"
)

type StorageCleanupRepository interface {
	Enqueue(ctx context.Context, task *models.StorageCleanupTask) error
	GetDueTasks(ctx context.Context, limit int) ([]models.StorageCleanupTask, error)
	MarkProcessing(ctx context.Context, id string) error
	MarkCompleted(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, lastErr string, nextRunAt time.Time, exhausted bool) error
	FindByID(ctx context.Context, id string) (*models.StorageCleanupTask, error)
}

const queryWhereID = "id = ?"

type storageCleanupRepository struct {
	db *gorm.DB
}

func NewStorageCleanupRepository(db *gorm.DB) StorageCleanupRepository {
	return &storageCleanupRepository{db: db}
}

func (r *storageCleanupRepository) Enqueue(ctx context.Context, task *models.StorageCleanupTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *storageCleanupRepository) GetDueTasks(ctx context.Context, limit int) ([]models.StorageCleanupTask, error) {
	var tasks []models.StorageCleanupTask
	err := r.db.WithContext(ctx).
		Where("status IN ? AND next_run_at <= ?", []string{models.CleanupStatusPending, models.CleanupStatusProcessing}, time.Now()).
		Order("next_run_at ASC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

func (r *storageCleanupRepository) MarkProcessing(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.StorageCleanupTask{}).
		Where(queryWhereID, id).
		Updates(map[string]any{
			"status":     models.CleanupStatusProcessing,
			"updated_at": time.Now(),
		}).Error
}

func (r *storageCleanupRepository) MarkCompleted(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&models.StorageCleanupTask{}).
		Where(queryWhereID, id).
		Updates(map[string]any{
			"status":     models.CleanupStatusCompleted,
			"updated_at": time.Now(),
		}).Error
}

func (r *storageCleanupRepository) MarkFailed(ctx context.Context, id string, lastErr string, nextRunAt time.Time, exhausted bool) error {
	status := models.CleanupStatusPending
	if exhausted {
		status = models.CleanupStatusFailed
	}
	return r.db.WithContext(ctx).
		Model(&models.StorageCleanupTask{}).
		Where(queryWhereID, id).
		Updates(map[string]any{
			"status":      status,
			"last_error":  lastErr,
			"attempts":    gorm.Expr("attempts + 1"),
			"next_run_at": nextRunAt,
			"updated_at":  time.Now(),
		}).Error
}

func (r *storageCleanupRepository) FindByID(ctx context.Context, id string) (*models.StorageCleanupTask, error) {
	var task models.StorageCleanupTask
	err := r.db.WithContext(ctx).First(&task, queryWhereID, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}
