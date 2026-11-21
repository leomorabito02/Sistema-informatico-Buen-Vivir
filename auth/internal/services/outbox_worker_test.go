package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"auth/internal/events"
	"auth/internal/models"
	"auth/internal/services"

	"cloud.google.com/go/pubsub/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type mockOutboxRepository struct {
	mock.Mock
}

func (m *mockOutboxRepository) Enqueue(ctx context.Context, event *models.OutboxEvent) error {
	return nil
}
func (m *mockOutboxRepository) EnqueueTx(ctx context.Context, tx *gorm.DB, event *models.OutboxEvent) error {
	return nil
}
func (m *mockOutboxRepository) GetDueEvents(ctx context.Context, limit int) ([]models.OutboxEvent, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.OutboxEvent), args.Error(1)
}
func (m *mockOutboxRepository) MarkProcessing(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockOutboxRepository) MarkCompleted(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockOutboxRepository) MarkFailed(ctx context.Context, id string, lastErr string, nextRunAt time.Time, dead bool) error {
	args := m.Called(ctx, id, lastErr, nextRunAt, dead)
	return args.Error(0)
}

type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, msg *pubsub.Message) error {
	args := m.Called(ctx, topic, msg)
	return args.Error(0)
}

func TestOutboxWorker_ProcessDueEvents(t *testing.T) {
	ctx := context.Background()

	t.Run("Success publishes and marks completed", func(t *testing.T) {
		repo := new(mockOutboxRepository)
		pub := new(mockPublisher)

		env, err := events.New(ctx, events.ProducerAutenticacion, events.ParticipanteCreado, events.ParticipantePayload{
			ParticipanteID: "part-1",
			UserID:         "user-1",
		})
		assert.NoError(t, err)
		raw, _ := json.Marshal(env)

		event := models.OutboxEvent{
			ID:          "evt-1",
			Topic:       "usuarios-eventos",
			EventType:   events.ParticipanteCreado,
			Payload:     raw,
			OrderingKey: "part-1",
			Attempts:    0,
			MaxAttempts: 5,
		}

		repo.On("GetDueEvents", ctx, 50).Return([]models.OutboxEvent{event}, nil)
		repo.On("MarkProcessing", ctx, "evt-1").Return(nil)
		pub.On("Publish", ctx, "usuarios-eventos", mock.MatchedBy(func(m *pubsub.Message) bool {
			return m.OrderingKey == "part-1" && m.Attributes[events.AttrEventType] == events.ParticipanteCreado
		})).Return(nil)
		repo.On("MarkCompleted", ctx, "evt-1").Return(nil)

		w := services.NewOutboxWorker(repo, pub)
		processed, err := w.ProcessDueEvents(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 1, processed)

		repo.AssertExpectations(t)
		pub.AssertExpectations(t)
	})

	t.Run("Publish failure triggers backoff retry", func(t *testing.T) {
		repo := new(mockOutboxRepository)
		pub := new(mockPublisher)

		env, _ := events.New(ctx, events.ProducerAutenticacion, events.ParticipanteCreado, events.ParticipantePayload{
			ParticipanteID: "part-2",
		})
		raw, _ := json.Marshal(env)

		event := models.OutboxEvent{
			ID:          "evt-2",
			Topic:       "usuarios-eventos",
			EventType:   events.ParticipanteCreado,
			Payload:     raw,
			Attempts:    0,
			MaxAttempts: 5,
		}

		repo.On("GetDueEvents", ctx, 50).Return([]models.OutboxEvent{event}, nil)
		repo.On("MarkProcessing", ctx, "evt-2").Return(nil)
		pub.On("Publish", ctx, "usuarios-eventos", mock.Anything).Return(errors.New("connection failed"))
		repo.On("MarkFailed", ctx, "evt-2", "connection failed", mock.AnythingOfType("time.Time"), false).Return(nil)

		w := services.NewOutboxWorker(repo, pub)
		processed, err := w.ProcessDueEvents(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 0, processed)

		repo.AssertExpectations(t)
		pub.AssertExpectations(t)
	})

	t.Run("Start and Edge Cases", func(t *testing.T) {
		// Nil worker
		nilWorker := services.NewOutboxWorker(nil, nil)
		nilWorker.Start(context.Background(), time.Second)
		processed, err := nilWorker.ProcessDueEvents(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 0, processed)

		// Start with cancellation
		repo := new(mockOutboxRepository)
		pub := new(mockPublisher)
		w := services.NewOutboxWorker(repo, pub)
		ctx, cancel := context.WithCancel(context.Background())
		repo.On("GetDueEvents", mock.Anything, mock.Anything).Return([]models.OutboxEvent{}, nil)
		w.Start(ctx, 10*time.Millisecond)
		time.Sleep(25 * time.Millisecond)
		cancel()

		// GetDueEvents error on separate worker
		repoErr := new(mockOutboxRepository)
		wErr := services.NewOutboxWorker(repoErr, pub)
		repoErr.On("GetDueEvents", mock.Anything, 50).Return(nil, errors.New("db error"))
		_, err = wErr.ProcessDueEvents(context.Background())
		assert.Error(t, err)

		// MarkProcessing error
		repoProc := new(mockOutboxRepository)
		wProc := services.NewOutboxWorker(repoProc, pub)
		event := models.OutboxEvent{ID: "e-fail", Attempts: 15, MaxAttempts: 5}
		repoProc.On("GetDueEvents", mock.Anything, 50).Return([]models.OutboxEvent{event}, nil)
		repoProc.On("MarkProcessing", mock.Anything, "e-fail").Return(errors.New("lock error"))
		processed, err = wProc.ProcessDueEvents(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 0, processed)

		// Corrupted payload
		repoCorrupt := new(mockOutboxRepository)
		wCorrupt := services.NewOutboxWorker(repoCorrupt, pub)
		corruptEvent := models.OutboxEvent{
			ID:          "e-corrupt",
			Payload:     []byte("invalid json"),
			Attempts:    15,
			MaxAttempts: 5,
		}
		repoCorrupt.On("GetDueEvents", mock.Anything, 50).Return([]models.OutboxEvent{corruptEvent}, nil)
		repoCorrupt.On("MarkProcessing", mock.Anything, "e-corrupt").Return(nil)
		repoCorrupt.On("MarkFailed", mock.Anything, "e-corrupt", mock.Anything, mock.AnythingOfType("time.Time"), true).Return(nil)
		processed, err = wCorrupt.ProcessDueEvents(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 0, processed)
	})
}

func TestOutboxInlineFlushEnabled(t *testing.T) {
	t.Run("disabled by default locally", func(t *testing.T) {
		// Arrange
		t.Setenv("K_SERVICE", "")
		t.Setenv("OUTBOX_INLINE_FLUSH", "")

		// Act & Assert
		assert.False(t, services.OutboxInlineFlushEnabled())
	})

	t.Run("enabled automatically on Cloud Run", func(t *testing.T) {
		// Arrange
		t.Setenv("K_SERVICE", "auth-testing")
		t.Setenv("OUTBOX_INLINE_FLUSH", "")

		// Act & Assert
		assert.True(t, services.OutboxInlineFlushEnabled())
	})

	t.Run("explicit false disables on Cloud Run", func(t *testing.T) {
		// Arrange
		t.Setenv("K_SERVICE", "auth-testing")
		t.Setenv("OUTBOX_INLINE_FLUSH", "false")

		// Act & Assert
		assert.False(t, services.OutboxInlineFlushEnabled())
	})

	t.Run("explicit true enables locally", func(t *testing.T) {
		// Arrange
		t.Setenv("K_SERVICE", "")
		t.Setenv("OUTBOX_INLINE_FLUSH", "true")

		// Act & Assert
		assert.True(t, services.OutboxInlineFlushEnabled())
	})
}

func TestOutboxWorker_InlineFlushAndBestEffort(t *testing.T) {
	t.Run("FlushIfDirty skips repository when clean", func(t *testing.T) {
		// Arrange
		repo := new(mockOutboxRepository)
		pub := new(mockPublisher)
		w := services.NewOutboxWorker(repo, pub)

		// Act
		processed := w.FlushIfDirty(context.Background(), time.Second)

		// Assert
		assert.Equal(t, 0, processed)
		repo.AssertNotCalled(t, "GetDueEvents", mock.Anything, mock.Anything)
	})

	t.Run("MarkDirty triggers FlushIfDirty once", func(t *testing.T) {
		// Arrange
		repo := new(mockOutboxRepository)
		pub := new(mockPublisher)
		env, _ := events.New(context.Background(), events.ProducerAutenticacion, events.ParticipanteCreado, events.ParticipantePayload{ParticipanteID: "p-inline"})
		raw, _ := json.Marshal(env)
		event := models.OutboxEvent{ID: "evt-inline", Topic: "usuarios-eventos", EventType: events.ParticipanteCreado, Payload: raw, MaxAttempts: 5}

		repo.On("GetDueEvents", mock.Anything, 50).Return([]models.OutboxEvent{event}, nil).Once()
		repo.On("MarkProcessing", mock.Anything, "evt-inline").Return(nil).Once()
		pub.On("Publish", mock.Anything, "usuarios-eventos", mock.Anything).Return(nil).Once()
		repo.On("MarkCompleted", mock.Anything, "evt-inline").Return(nil).Once()

		w := services.NewOutboxWorker(repo, pub)
		w.MarkDirty()

		// Act
		first := w.FlushIfDirty(context.Background(), time.Second)
		second := w.FlushIfDirty(context.Background(), time.Second)

		// Assert
		assert.Equal(t, 1, first)
		assert.Equal(t, 0, second)
		repo.AssertExpectations(t)
		pub.AssertExpectations(t)
	})

	t.Run("FlushBestEffort swallows errors and handles nil dependencies", func(t *testing.T) {
		// Arrange
		nilWorker := services.NewOutboxWorker(nil, nil)
		repoErr := new(mockOutboxRepository)
		pub := new(mockPublisher)
		repoErr.On("GetDueEvents", mock.Anything, 50).Return(nil, errors.New("temporary db error")).Once()
		wErr := services.NewOutboxWorker(repoErr, pub)

		// Act & Assert
		assert.Equal(t, 0, nilWorker.FlushBestEffort(context.Background(), 0))
		assert.Equal(t, 0, wErr.FlushBestEffort(nil, 0))
		repoErr.AssertExpectations(t)
	})
}

