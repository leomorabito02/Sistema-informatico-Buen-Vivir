package outbox

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ProcessedCollection guarda los eventos ya procesados por cada
// consumidor (idempotencia). Tiene una política de TTL de Firestore
// sobre expira_en (ver infra/pubsub/setup.sh): pasado ese plazo Pub/Sub
// ya no puede reentregar el mensaje.
const ProcessedCollection = "eventos_procesados"

// ProcessedTTL es cuánto se conserva la marca de un evento procesado.
// Pub/Sub retiene mensajes como máximo 7 días; se deja margen.
const ProcessedTTL = 30 * 24 * time.Hour

// FirestoreProcessedStore implementa events.ProcessedStore en Firestore.
type FirestoreProcessedStore struct {
	client *firestore.Client
	now    func() time.Time
}

func NewFirestoreProcessedStore(client *firestore.Client) *FirestoreProcessedStore {
	return &FirestoreProcessedStore{client: client, now: time.Now}
}

type processedDoc struct {
	Consumer    string    `firestore:"consumidor"`
	EventID     string    `firestore:"event_id"`
	EventType   string    `firestore:"event_type"`
	ProcessedAt time.Time `firestore:"procesado_en"`
	ExpiresAt   time.Time `firestore:"expira_en"`
}

func processedID(consumer, eventID string) string { return consumer + "_" + eventID }

func (s *FirestoreProcessedStore) IsProcessed(ctx context.Context, consumer, eventID string) (bool, error) {
	_, err := s.client.Collection(ProcessedCollection).Doc(processedID(consumer, eventID)).Get(ctx)
	if status.Code(err) == codes.NotFound {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("error consultando eventos procesados: %w", err)
	}
	return true, nil
}

func (s *FirestoreProcessedStore) MarkProcessed(ctx context.Context, consumer, eventID, eventType string) error {
	now := s.now().UTC()
	_, err := s.client.Collection(ProcessedCollection).Doc(processedID(consumer, eventID)).Set(ctx, processedDoc{
		Consumer:    consumer,
		EventID:     eventID,
		EventType:   eventType,
		ProcessedAt: now,
		ExpiresAt:   now.Add(ProcessedTTL),
	})
	if err != nil {
		return fmt.Errorf("error marcando evento procesado: %w", err)
	}
	return nil
}
