package producers_test

import (
	"context"
	"testing"
	"time"

	"auth/internal/events"
	"auth/internal/events/producers"
	"auth/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type mockOutboxRepo struct {
	mock.Mock
}

func (m *mockOutboxRepo) Enqueue(ctx context.Context, event *models.OutboxEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}
func (m *mockOutboxRepo) EnqueueTx(ctx context.Context, tx *gorm.DB, event *models.OutboxEvent) error {
	args := m.Called(ctx, tx, event)
	return args.Error(0)
}
func (m *mockOutboxRepo) GetDueEvents(ctx context.Context, limit int) ([]models.OutboxEvent, error) {
	return nil, nil
}
func (m *mockOutboxRepo) MarkProcessing(ctx context.Context, id string) error {
	return nil
}
func (m *mockOutboxRepo) MarkCompleted(ctx context.Context, id string) error {
	return nil
}
func (m *mockOutboxRepo) MarkFailed(ctx context.Context, id string, lastErr string, nextRunAt time.Time, dead bool) error {
	return nil
}

func TestUsuariosProducer(t *testing.T) {
	ctx := context.Background()

	t.Run("ParticipanteCreado enqueues event", func(t *testing.T) {
		repo := new(mockOutboxRepo)
		repo.On("EnqueueTx", ctx, (*gorm.DB)(nil), mock.MatchedBy(func(e *models.OutboxEvent) bool {
			return e.Topic == "usuarios-eventos" && e.EventType == events.ParticipanteCreado && e.OrderingKey == "p-1"
		})).Return(nil)

		p := producers.NewUsuariosProducer(repo)
		err := p.ParticipanteCreado(ctx, "p-1", "u-1")
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("ParticipanteEliminado enqueues event", func(t *testing.T) {
		repo := new(mockOutboxRepo)
		repo.On("EnqueueTx", ctx, (*gorm.DB)(nil), mock.MatchedBy(func(e *models.OutboxEvent) bool {
			return e.Topic == "usuarios-eventos" && e.EventType == events.ParticipanteEliminado && e.OrderingKey == "p-2"
		})).Return(nil)

		p := producers.NewUsuariosProducer(repo)
		err := p.ParticipanteEliminado(ctx, "p-2", "u-2")
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("UsuarioActualizado enqueues event", func(t *testing.T) {
		repo := new(mockOutboxRepo)
		repo.On("EnqueueTx", ctx, (*gorm.DB)(nil), mock.MatchedBy(func(e *models.OutboxEvent) bool {
			return e.Topic == "usuarios-eventos" && e.EventType == events.UsuarioActualizado && e.OrderingKey == "p-3"
		})).Return(nil)

		p := producers.NewUsuariosProducer(repo)
		err := p.UsuarioActualizado(ctx, events.UsuarioActualizadoPayload{
			UserID:         "u-3",
			RoleID:         "role_participant",
			ParticipanteID: "p-3",
			Status:         "active",
			ApprovalStatus: "approved",
		})
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("UsuarioCreado enqueues event", func(t *testing.T) {
		repo := new(mockOutboxRepo)
		repo.On("EnqueueTx", ctx, (*gorm.DB)(nil), mock.MatchedBy(func(e *models.OutboxEvent) bool {
			return e.Topic == "usuarios-eventos" && e.EventType == events.UsuarioCreado && e.OrderingKey == "u-10"
		})).Return(nil)

		p := producers.NewUsuariosProducer(repo)
		err := p.UsuarioCreado(ctx, "u-10", "role_participant")
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("EmailSolicitado enqueues event", func(t *testing.T) {
		repo := new(mockOutboxRepo)
		repo.On("EnqueueTx", ctx, (*gorm.DB)(nil), mock.MatchedBy(func(e *models.OutboxEvent) bool {
			return e.Topic == "usuarios-eventos" && e.EventType == events.EmailSolicitado && e.OrderingKey == "u-10"
		})).Return(nil)

		p := producers.NewUsuariosProducer(repo)
		err := p.EmailSolicitado(ctx, "ev-1", "u-10", "verificacion")
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("onEnqueue callback runs after successful enqueue", func(t *testing.T) {
		// Arrange
		repo := new(mockOutboxRepo)
		repo.On("EnqueueTx", ctx, (*gorm.DB)(nil), mock.Anything).Return(nil).Once()
		notified := 0
		p := producers.NewUsuariosProducer(repo, func() { notified++ })

		// Act
		err := p.UsuarioCreado(ctx, "u-11", "role_participant")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 1, notified)
		repo.AssertExpectations(t)
	})
}

