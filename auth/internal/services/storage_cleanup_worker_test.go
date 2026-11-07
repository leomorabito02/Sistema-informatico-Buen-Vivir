package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth/internal/models"
	"auth/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStorageCleanupWorker_ProcessDueTasks(t *testing.T) {
	t.Run("Processes task and marks completed on storage delete success", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockStorageCleanupRepo)
		mockStorage := new(MockObjectStorageService)
		worker := services.NewStorageCleanupWorker(mockRepo, mockStorage, 10)

		ctx := context.Background()
		tasks := []models.StorageCleanupTask{
			{
				ID:          "task-1",
				ObjectName:  "documents/part-1/old.pdf",
				Status:      models.CleanupStatusPending,
				Attempts:    0,
				MaxAttempts: 5,
				NextRunAt:   time.Now().Add(-time.Minute),
			},
		}

		mockRepo.On("GetDueTasks", ctx, 10).Return(tasks, nil)
		mockRepo.On("MarkProcessing", ctx, "task-1").Return(nil)
		mockStorage.On("DeleteFile", ctx, "documents/part-1/old.pdf").Return(nil)
		mockRepo.On("MarkCompleted", ctx, "task-1").Return(nil)

		// Act
		processed, err := worker.ProcessDueTasks(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 1, processed)
		mockRepo.AssertCalled(t, "MarkCompleted", ctx, "task-1")
	})

	t.Run("Marks task failed with backoff when storage delete fails", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockStorageCleanupRepo)
		mockStorage := new(MockObjectStorageService)
		worker := services.NewStorageCleanupWorker(mockRepo, mockStorage, 10)

		ctx := context.Background()
		tasks := []models.StorageCleanupTask{
			{
				ID:          "task-retry",
				ObjectName:  "documents/part-1/failed.pdf",
				Status:      models.CleanupStatusPending,
				Attempts:    1,
				MaxAttempts: 5,
				NextRunAt:   time.Now().Add(-time.Minute),
			},
		}

		mockRepo.On("GetDueTasks", ctx, 10).Return(tasks, nil)
		mockRepo.On("MarkProcessing", ctx, "task-retry").Return(nil)
		mockStorage.On("DeleteFile", ctx, "documents/part-1/failed.pdf").Return(errors.New("connection reset by peer"))
		mockRepo.On("MarkFailed", ctx, "task-retry", "connection reset by peer", mock.AnythingOfType("time.Time"), false).Return(nil)

		// Act
		processed, err := worker.ProcessDueTasks(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 0, processed)
		mockRepo.AssertCalled(t, "MarkFailed", ctx, "task-retry", "connection reset by peer", mock.AnythingOfType("time.Time"), false)
	})

	t.Run("Marks task exhausted when max attempts reached", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockStorageCleanupRepo)
		mockStorage := new(MockObjectStorageService)
		worker := services.NewStorageCleanupWorker(mockRepo, mockStorage, 10)

		ctx := context.Background()
		tasks := []models.StorageCleanupTask{
			{
				ID:          "task-exhausted",
				ObjectName:  "documents/part-1/never.pdf",
				Status:      models.CleanupStatusProcessing,
				Attempts:    4,
				MaxAttempts: 5,
				NextRunAt:   time.Now().Add(-time.Minute),
			},
		}

		mockRepo.On("GetDueTasks", ctx, 10).Return(tasks, nil)
		mockRepo.On("MarkProcessing", ctx, "task-exhausted").Return(nil)
		mockStorage.On("DeleteFile", ctx, "documents/part-1/never.pdf").Return(errors.New("permanent storage error"))
		mockRepo.On("MarkFailed", ctx, "task-exhausted", "permanent storage error", mock.AnythingOfType("time.Time"), true).Return(nil)

		// Act
		processed, err := worker.ProcessDueTasks(ctx)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 0, processed)
		mockRepo.AssertCalled(t, "MarkFailed", ctx, "task-exhausted", "permanent storage error", mock.AnythingOfType("time.Time"), true)
	})

	t.Run("Start and Edge Cases", func(t *testing.T) {
		// Nil worker
		nilWorker := services.NewStorageCleanupWorker(nil, nil)
		nilWorker.Start(context.Background(), time.Second)
		processed, err := nilWorker.ProcessDueTasks(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 0, processed)

		// Start with cancellation
		mockRepo := new(MockStorageCleanupRepo)
		mockStorage := new(MockObjectStorageService)
		worker := services.NewStorageCleanupWorker(mockRepo, mockStorage)
		ctx, cancel := context.WithCancel(context.Background())
		mockRepo.On("GetDueTasks", mock.Anything, mock.Anything).Return([]models.StorageCleanupTask{}, nil)
		worker.Start(ctx, 10*time.Millisecond)
		time.Sleep(25 * time.Millisecond)
		cancel()

		// GetDueTasks error on separate worker
		mockRepoErr := new(MockStorageCleanupRepo)
		workerErr := services.NewStorageCleanupWorker(mockRepoErr, mockStorage)
		mockRepoErr.On("GetDueTasks", mock.Anything, 50).Return(nil, errors.New("db fail"))
		_, err = workerErr.ProcessDueTasks(context.Background())
		assert.Error(t, err)

		// MarkProcessing error
		mockRepoProc := new(MockStorageCleanupRepo)
		workerProc := services.NewStorageCleanupWorker(mockRepoProc, mockStorage)
		task := models.StorageCleanupTask{ID: "t-fail", Attempts: 15, MaxAttempts: 5}
		mockRepoProc.On("GetDueTasks", mock.Anything, 50).Return([]models.StorageCleanupTask{task}, nil)
		mockRepoProc.On("MarkProcessing", mock.Anything, "t-fail").Return(errors.New("lock error"))
		processed, err = workerProc.ProcessDueTasks(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 0, processed)
	})
}
