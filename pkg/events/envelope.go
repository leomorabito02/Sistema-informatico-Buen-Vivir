// Package events es la librería común de mensajería (Pub/Sub): envelope, catálogo,
// handler de suscripciones push y helpers de contexto. Es la copia canónica: cada
// servicio tiene una copia en internal/events generada con scripts/sync-eventlib.sh
// (ver docs/adr/0001-libreria-de-eventos-copiada.md).
package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Atributos del mensaje de Pub/Sub. event_type y version viajan además
// como atributos (no solo dentro del JSON) para poder filtrar las
// suscripciones sin leer el cuerpo del mensaje.
const (
	AttrEventType     = "event_type"
	AttrVersion       = "version"
	AttrProducer      = "producer"
	AttrCorrelationID = "correlation_id"
)

// Envelope es el formato estándar de todos los eventos del sistema.
// Payload lleva SOLO ids y metadatos (nunca datos personales ni
// contenido de fichas): si un consumidor necesita más, se lo pide por
// gRPC/HTTP al servicio dueño del dato.
type Envelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	Version       int             `json:"version"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Producer      string          `json:"producer"`
	CorrelationID string          `json:"correlation_id"`
	ActorID       string          `json:"actor_id"`
	Payload       json.RawMessage `json:"payload"`
}

// ErrInvalidEnvelope señala un mensaje que no cumple el formato estándar.
var ErrInvalidEnvelope = errors.New("envelope inválido")

// New arma un envelope nuevo para eventType con la versión vigente del
// catálogo. El actor y el correlation id salen del contexto (ver
// WithActor / WithCorrelationID); si no hay correlation id se genera
// uno, así todo evento es trazable aunque no venga de un request.
func New(ctx context.Context, producer, eventType string, payload any) (Envelope, error) {
	return NewWithID(ctx, NewUUID(), producer, eventType, payload)
}

// NewWithID es New con un event_id elegido por quien publica. Sirve
// cuando el mismo hecho puede llegar dos veces al productor (por
// ejemplo, Contentful reintenta un webhook): con un id determinístico
// el duplicado se descarta en el consumidor.
func NewWithID(ctx context.Context, eventID, producer, eventType string, payload any) (Envelope, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, fmt.Errorf("no se pudo serializar el payload de %s: %w", eventType, err)
	}
	correlationID := CorrelationIDFrom(ctx)
	if correlationID == "" {
		correlationID = NewUUID()
	}
	env := Envelope{
		EventID:       eventID,
		EventType:     eventType,
		Version:       CurrentVersion(eventType),
		OccurredAt:    time.Now().UTC().Truncate(time.Millisecond),
		Producer:      producer,
		CorrelationID: correlationID,
		ActorID:       ActorFrom(ctx),
		Payload:       data,
	}
	if err := env.Validate(); err != nil {
		return Envelope{}, err
	}
	return env, nil
}

// Validate controla los campos obligatorios del envelope. actor_id
// puede ir vacío: hay eventos que no origina un usuario (por ejemplo,
// un webhook de Contentful).
func (e Envelope) Validate() error {
	var missing []string
	if !IsUUID(e.EventID) {
		missing = append(missing, "event_id (uuid)")
	}
	if strings.TrimSpace(e.EventType) == "" {
		missing = append(missing, "event_type")
	}
	if e.Version < 1 {
		missing = append(missing, "version (>= 1)")
	}
	if e.OccurredAt.IsZero() {
		missing = append(missing, "occurred_at")
	}
	if strings.TrimSpace(e.Producer) == "" {
		missing = append(missing, "producer")
	}
	if strings.TrimSpace(e.CorrelationID) == "" {
		missing = append(missing, "correlation_id")
	}
	if len(e.Payload) == 0 || string(e.Payload) == "null" {
		missing = append(missing, "payload")
	}
	if len(missing) > 0 {
		return fmt.Errorf("%w: faltan o son inválidos %s", ErrInvalidEnvelope, strings.Join(missing, ", "))
	}
	return nil
}

// Marshal serializa el envelope tal como viaja en el cuerpo del mensaje.
func (e Envelope) Marshal() ([]byte, error) {
	return json.Marshal(e)
}

// Attributes son los atributos de Pub/Sub del mensaje (ver Attr*).
func (e Envelope) Attributes() map[string]string {
	return map[string]string{
		AttrEventType:     e.EventType,
		AttrVersion:       strconv.Itoa(e.Version),
		AttrProducer:      e.Producer,
		AttrCorrelationID: e.CorrelationID,
	}
}

// Decode lee un envelope del cuerpo de un mensaje y lo valida.
func Decode(data []byte) (Envelope, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return Envelope{}, fmt.Errorf("%w: %v", ErrInvalidEnvelope, err)
	}
	if err := env.Validate(); err != nil {
		return Envelope{}, err
	}
	return env, nil
}

// DecodePayload lee el payload del envelope en el struct del evento.
func DecodePayload[T any](env Envelope) (T, error) {
	var out T
	if err := json.Unmarshal(env.Payload, &out); err != nil {
		return out, fmt.Errorf("payload inválido para %s v%d: %w", env.EventType, env.Version, err)
	}
	return out, nil
}
