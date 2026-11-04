package producers

import (
	"context"
	"encoding/json"
	"time"

	"auth/internal/events"
	"auth/internal/models"
	"auth/internal/repositories"

	"gorm.io/gorm"
)

const (
	TopicUsuarios = "usuarios-eventos"
)

// UsuariosProducer defines methods to enqueue domain events in Auth's outbox.
type UsuariosProducer interface {
	UsuarioCreado(ctx context.Context, userID, roleID string, tx ...*gorm.DB) error
	EmailSolicitado(ctx context.Context, requestID, userID, tipo string, tx ...*gorm.DB) error
	ParticipanteCreado(ctx context.Context, participanteID, userID string, tx ...*gorm.DB) error
	ParticipanteEliminado(ctx context.Context, participanteID, userID string, tx ...*gorm.DB) error
	UsuarioActualizado(ctx context.Context, payload events.UsuarioActualizadoPayload, tx ...*gorm.DB) error
}

type usuariosProducer struct {
	repo repositories.OutboxRepository
}

// NewUsuariosProducer creates a new UsuariosProducer.
func NewUsuariosProducer(repo repositories.OutboxRepository) UsuariosProducer {
	return &usuariosProducer{repo: repo}
}

func (p *usuariosProducer) emit(ctx context.Context, eventType string, payload any, orderingKey string, tx ...*gorm.DB) error {
	if p == nil || p.repo == nil {
		return nil
	}

	env, err := events.New(ctx, events.ProducerAutenticacion, eventType, payload)
	if err != nil {
		return err
	}

	raw, err := json.Marshal(env)
	if err != nil {
		return err
	}

	var gormTx *gorm.DB
	if len(tx) > 0 {
		gormTx = tx[0]
	}

	return p.repo.EnqueueTx(ctx, gormTx, &models.OutboxEvent{
		ID:          env.EventID,
		Topic:       TopicUsuarios,
		EventType:   eventType,
		Payload:     raw,
		OrderingKey: orderingKey,
		Status:      models.OutboxStatusPending,
		Attempts:    0,
		MaxAttempts: 5,
		NextRunAt:   time.Now(),
	})
}

func (p *usuariosProducer) UsuarioCreado(ctx context.Context, userID, roleID string, tx ...*gorm.DB) error {
	return p.emit(ctx, events.UsuarioCreado, events.UsuarioCreadoPayload{
		UserID: userID,
		RoleID: roleID,
	}, userID, tx...)
}

func (p *usuariosProducer) EmailSolicitado(ctx context.Context, requestID, userID, tipo string, tx ...*gorm.DB) error {
	return p.emit(ctx, events.EmailSolicitado, events.EmailSolicitadoPayload{
		RequestID: requestID,
		UserID:    userID,
		Tipo:      tipo,
	}, userID, tx...)
}

func (p *usuariosProducer) ParticipanteCreado(ctx context.Context, participanteID, userID string, tx ...*gorm.DB) error {
	return p.emit(ctx, events.ParticipanteCreado, events.ParticipantePayload{
		ParticipanteID: participanteID,
		UserID:         userID,
	}, participanteID, tx...)
}

func (p *usuariosProducer) ParticipanteEliminado(ctx context.Context, participanteID, userID string, tx ...*gorm.DB) error {
	return p.emit(ctx, events.ParticipanteEliminado, events.ParticipantePayload{
		ParticipanteID: participanteID,
		UserID:         userID,
	}, participanteID, tx...)
}

func (p *usuariosProducer) UsuarioActualizado(ctx context.Context, payload events.UsuarioActualizadoPayload, tx ...*gorm.DB) error {
	orderingKey := payload.ParticipanteID
	if orderingKey == "" {
		orderingKey = payload.UserID
	}
	return p.emit(ctx, events.UsuarioActualizado, payload, orderingKey, tx...)
}

