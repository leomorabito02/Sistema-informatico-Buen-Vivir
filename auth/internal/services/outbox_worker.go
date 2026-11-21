package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"auth/internal/events"
	"auth/internal/models"
	"auth/internal/repositories"

	"cloud.google.com/go/pubsub/v2"
)

const defaultOutboxInlineTimeout = 2 * time.Second

// OutboxInlineFlushEnabled returns true on Cloud Run (K_SERVICE set) or when
// OUTBOX_INLINE_FLUSH=true, allowing pending outbox rows to flush before the HTTP
// response finishes so Cloud Run can scale to zero without external polling.
func OutboxInlineFlushEnabled() bool {
	if raw := strings.TrimSpace(os.Getenv("OUTBOX_INLINE_FLUSH")); raw != "" {
		return strings.EqualFold(raw, "true") || raw == "1"
	}
	return strings.TrimSpace(os.Getenv("K_SERVICE")) != ""
}

// EventPublisher abstracts publishing a single message to Pub/Sub.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, msg *pubsub.Message) error
}

// OutboxWorker manages publishing pending outbox events to Pub/Sub with retries.
type OutboxWorker interface {
	Start(ctx context.Context, interval time.Duration)
	ProcessDueEvents(ctx context.Context) (int, error)
	MarkDirty()
	FlushBestEffort(ctx context.Context, timeout time.Duration) int
	FlushIfDirty(ctx context.Context, timeout time.Duration) int
}

type outboxWorker struct {
	repo      repositories.OutboxRepository
	publisher EventPublisher
	batchSize int
	mu        sync.Mutex
	dirty     atomic.Bool
}

// NewOutboxWorker creates a new OutboxWorker instance.
func NewOutboxWorker(repo repositories.OutboxRepository, publisher EventPublisher, batchSize ...int) OutboxWorker {
	size := 50
	if len(batchSize) > 0 && batchSize[0] > 0 {
		size = batchSize[0]
	}
	return &outboxWorker{
		repo:      repo,
		publisher: publisher,
		batchSize: size,
	}
}

func (w *outboxWorker) Start(ctx context.Context, interval time.Duration) {
	if w.repo == nil || w.publisher == nil {
		slog.Warn("Outbox worker not started: repository or publisher is nil")
		return
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				slog.Info("Outbox worker stopped gracefully")
				return
			case <-ticker.C:
				if _, err := w.ProcessDueEvents(ctx); err != nil {
					slog.Error("Outbox worker batch processing error", "error", err)
				}
			}
		}
	}()
	slog.Info("Outbox worker started", "interval", interval.String())
}

func (w *outboxWorker) handleFailure(ctx context.Context, event models.OutboxEvent, pubErr error) {
	nextAttempts := event.Attempts + 1
	exhausted := nextAttempts >= event.MaxAttempts
	backoffSeconds := math.Pow(2, float64(nextAttempts)) * 5
	if backoffSeconds > 3600 {
		backoffSeconds = 3600
	}
	nextRunAt := time.Now().Add(time.Duration(backoffSeconds) * time.Second)

	if failErr := w.repo.MarkFailed(ctx, event.ID, pubErr.Error(), nextRunAt, exhausted); failErr != nil {
		slog.Error("Failed to update failed outbox event status", "id", event.ID, "error", failErr)
	}
	slog.Warn("Failed to publish outbox event", "id", event.ID, "topic", event.Topic, "type", event.EventType, "attempt", nextAttempts, "exhausted", exhausted, "error", pubErr)
}

func (w *outboxWorker) processSingleEvent(ctx context.Context, event models.OutboxEvent) bool {
	if err := w.repo.MarkProcessing(ctx, event.ID); err != nil {
		slog.Warn("Failed to mark outbox event as processing", "id", event.ID, "error", err)
		return false
	}

	var env events.Envelope
	if err := json.Unmarshal(event.Payload, &env); err != nil {
		w.handleFailure(ctx, event, fmt.Errorf("corrupted outbox payload: %w", err))
		return false
	}

	msg := &pubsub.Message{
		Data:        event.Payload,
		Attributes:  env.Attributes(),
		OrderingKey: event.OrderingKey,
	}

	if err := w.publisher.Publish(ctx, event.Topic, msg); err != nil {
		w.handleFailure(ctx, event, err)
		return false
	}

	if err := w.repo.MarkCompleted(ctx, event.ID); err != nil {
		slog.Error("Failed to mark outbox event as completed", "id", event.ID, "error", err)
	}
	return true
}

func (w *outboxWorker) MarkDirty() {
	if w == nil {
		return
	}
	w.dirty.Store(true)
}

func (w *outboxWorker) FlushBestEffort(ctx context.Context, timeout time.Duration) int {
	if w == nil || w.repo == nil || w.publisher == nil {
		return 0
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		timeout = defaultOutboxInlineTimeout
	}
	flushCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	processed, err := w.ProcessDueEvents(flushCtx)
	if err != nil {
		slog.Warn("Outbox best-effort flush failed", "error", err)
		return processed
	}
	w.dirty.Store(false)
	return processed
}

func (w *outboxWorker) FlushIfDirty(ctx context.Context, timeout time.Duration) int {
	if w == nil || !w.dirty.Swap(false) {
		return 0
	}
	return w.FlushBestEffort(ctx, timeout)
}

func (w *outboxWorker) ProcessDueEvents(ctx context.Context) (int, error) {
	if w.repo == nil || w.publisher == nil {
		return 0, nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	events, err := w.repo.GetDueEvents(ctx, w.batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch due outbox events: %w", err)
	}

	processed := 0
	for _, event := range events {
		if ctx.Err() != nil {
			break
		}
		if w.processSingleEvent(ctx, event) {
			processed++
		}
	}

	return processed, nil
}

