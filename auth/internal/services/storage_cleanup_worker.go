package services

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"auth/internal/models"
	"auth/internal/repositories"
	"auth/internal/storage"
)

// StorageCleanupWorker processes asynchronous cleanup tasks to reconcile orphaned objects in storage.
type StorageCleanupWorker interface {
	Start(ctx context.Context, interval time.Duration)
	ProcessDueTasks(ctx context.Context) (int, error)
}

type storageCleanupWorker struct {
	repo      repositories.StorageCleanupRepository
	storage   storage.ObjectStorageService
	batchSize int
}

func NewStorageCleanupWorker(repo repositories.StorageCleanupRepository, storage storage.ObjectStorageService, batchSize ...int) StorageCleanupWorker {
	size := 50
	if len(batchSize) > 0 && batchSize[0] > 0 {
		size = batchSize[0]
	}
	return &storageCleanupWorker{
		repo:      repo,
		storage:   storage,
		batchSize: size,
	}
}

func (w *storageCleanupWorker) Start(ctx context.Context, interval time.Duration) {
	if w.repo == nil || w.storage == nil {
		slog.Warn("Storage cleanup worker not started: repository or storage service is nil")
		return
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				slog.Info("Storage cleanup worker stopped gracefully")
				return
			case <-ticker.C:
				if _, err := w.ProcessDueTasks(ctx); err != nil {
					slog.Error("Storage cleanup worker batch processing error", "error", err)
				}
			}
		}
	}()
	slog.Info("Storage cleanup worker started", "interval", interval.String())
}

func (w *storageCleanupWorker) handleTaskFailure(ctx context.Context, task models.StorageCleanupTask, delErr error) {
	nextAttempts := task.Attempts + 1
	exhausted := nextAttempts >= task.MaxAttempts
	backoffSeconds := math.Pow(2, float64(nextAttempts)) * 10
	if backoffSeconds > 3600 {
		backoffSeconds = 3600
	}
	nextRunAt := time.Now().Add(time.Duration(backoffSeconds) * time.Second)

	if failErr := w.repo.MarkFailed(ctx, task.ID, delErr.Error(), nextRunAt, exhausted); failErr != nil {
		slog.Error("Failed to update failed cleanup task status", "id", task.ID, "error", failErr)
	}
	slog.Warn("Failed to delete object from storage in cleanup task", "object", task.ObjectName, "attempt", nextAttempts, "exhausted", exhausted, "error", delErr)
}

func (w *storageCleanupWorker) processSingleTask(ctx context.Context, task models.StorageCleanupTask) bool {
	if err := w.repo.MarkProcessing(ctx, task.ID); err != nil {
		slog.Warn("Failed to mark cleanup task as processing", "id", task.ID, "error", err)
		return false
	}

	if delErr := w.storage.DeleteFile(ctx, task.ObjectName); delErr != nil {
		w.handleTaskFailure(ctx, task, delErr)
		return false
	}

	if compErr := w.repo.MarkCompleted(ctx, task.ID); compErr != nil {
		slog.Error("Failed to mark cleanup task as completed", "id", task.ID, "error", compErr)
	}
	return true
}

func (w *storageCleanupWorker) ProcessDueTasks(ctx context.Context) (int, error) {
	if w.repo == nil || w.storage == nil {
		return 0, nil
	}

	tasks, err := w.repo.GetDueTasks(ctx, w.batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch due cleanup tasks: %w", err)
	}

	processed := 0
	for _, task := range tasks {
		if ctx.Err() != nil {
			break
		}
		if w.processSingleTask(ctx, task) {
			processed++
		}
	}

	return processed, nil
}
