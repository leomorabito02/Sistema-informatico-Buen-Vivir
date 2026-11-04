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
}

