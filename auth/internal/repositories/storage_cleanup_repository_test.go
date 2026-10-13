package repositories_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"auth/internal/models"
	"auth/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestStorageCleanupRepository_Enqueue(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewStorageCleanupRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "storage_cleanup_tasks"`)).
			WithArgs("task-1", "documents/part-1/file.pdf", models.CleanupStatusPending, 0, 5, "", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		task := &models.StorageCleanupTask{
			ID:          "task-1",
			ObjectName:  "documents/part-1/file.pdf",
			Status:      models.CleanupStatusPending,
			Attempts:    0,
			MaxAttempts: 5,
			NextRunAt:   time.Now(),
		}

		err := repo.Enqueue(context.Background(), task)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "storage_cleanup_tasks"`)).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		task := &models.StorageCleanupTask{
			ID:         "task-err",
			ObjectName: "documents/err.pdf",
		}

		err := repo.Enqueue(context.Background(), task)
		assert.Error(t, err)
	})
}

func TestStorageCleanupRepository_GetDueTasks(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewStorageCleanupRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "storage_cleanup_tasks" WHERE status IN ($1,$2) AND next_run_at <= $3 ORDER BY next_run_at ASC LIMIT $4`)).
			WithArgs(models.CleanupStatusPending, models.CleanupStatusProcessing, sqlmock.AnyArg(), 10).
			WillReturnRows(sqlmock.NewRows([]string{"id", "object_name", "status"}).
				AddRow("task-1", "documents/file1.pdf", models.CleanupStatusPending))

		tasks, err := repo.GetDueTasks(context.Background(), 10)
		assert.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, "task-1", tasks[0].ID)
	})
}

func TestStorageCleanupRepository_StatusUpdates(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewStorageCleanupRepository(db)

	t.Run("MarkProcessing", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "storage_cleanup_tasks" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.MarkProcessing(context.Background(), "task-1")
		assert.NoError(t, err)
	})

	t.Run("MarkCompleted", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "storage_cleanup_tasks" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.MarkCompleted(context.Background(), "task-1")
		assert.NoError(t, err)
	})

	t.Run("MarkFailed_Retryable", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "storage_cleanup_tasks" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.MarkFailed(context.Background(), "task-1", "network timeout", time.Now().Add(time.Minute), false)
		assert.NoError(t, err)
	})

	t.Run("MarkFailed_Exhausted", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "storage_cleanup_tasks" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.MarkFailed(context.Background(), "task-1", "fatal error", time.Now(), true)
		assert.NoError(t, err)
	})
}
