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

func TestOutboxRepository_Enqueue(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewOutboxRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "outbox_events"`)).
			WithArgs("evt-1", "usuarios-eventos", "usuarios.participante.creado", []byte(`{}`), "p-1", models.OutboxStatusPending, 0, 5, "", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		event := &models.OutboxEvent{
			ID:          "evt-1",
			Topic:       "usuarios-eventos",
			EventType:   "usuarios.participante.creado",
			Payload:     []byte(`{}`),
			OrderingKey: "p-1",
			Status:      models.OutboxStatusPending,
			Attempts:    0,
			MaxAttempts: 5,
			NextRunAt:   time.Now(),
		}

		err := repo.Enqueue(context.Background(), event)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "outbox_events"`)).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		event := &models.OutboxEvent{
			ID:    "evt-err",
			Topic: "usuarios-eventos",
		}

		err := repo.Enqueue(context.Background(), event)
		assert.Error(t, err)
	})
}

func TestOutboxRepository_GetDueEvents(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewOutboxRepository(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "outbox_events" WHERE status IN ($1,$2) AND next_run_at <= $3 ORDER BY next_run_at ASC LIMIT $4`)).
		WithArgs(models.OutboxStatusPending, models.OutboxStatusProcessing, sqlmock.AnyArg(), 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "topic", "event_type"}).
			AddRow("evt-1", "usuarios-eventos", "usuarios.participante.creado"))

	events, err := repo.GetDueEvents(context.Background(), 10)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "evt-1", events[0].ID)
}

func TestOutboxRepository_StatusUpdates(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewOutboxRepository(db)

	t.Run("MarkProcessing", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "outbox_events" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.MarkProcessing(context.Background(), "evt-1")
		assert.NoError(t, err)
	})

	t.Run("MarkCompleted", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "outbox_events" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.MarkCompleted(context.Background(), "evt-1")
		assert.NoError(t, err)
	})

	t.Run("MarkFailed_Dead", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "outbox_events" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.MarkFailed(context.Background(), "evt-1", "permanent fail", time.Now(), true)
		assert.NoError(t, err)
	})
}

