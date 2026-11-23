package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"

	"buenvivir/pkg/events"
)

// Message es lo que el Relay le pasa al Publisher.
type Message struct {
	Data        []byte
	Attributes  map[string]string
	OrderingKey string
}

// Publisher manda un mensaje a un topic y espera la confirmación.
type Publisher interface {
	Publish(ctx context.Context, topic string, msg Message) error
}

// Relay publica los eventos pendientes del outbox, en segundo plano (Run) o a pedido
// (Flush, por ejemplo desde Cloud Scheduler, porque en Cloud Run sin requests no hay CPU).
type Relay struct {
	client    *firestore.Client
	publisher Publisher
	logger    *slog.Logger
	notify    chan struct{}
	mu        sync.Mutex

	BatchSize  int
	Interval   time.Duration // barrido periódico
	Lease      time.Duration // reserva de un evento mientras se publica
	MaxBackoff time.Duration
	Now        func() time.Time
}

func NewRelay(client *firestore.Client, publisher Publisher, logger *slog.Logger) *Relay {
	if logger == nil {
		logger = slog.Default()
	}
	return &Relay{
		client:     client,
		publisher:  publisher,
		logger:     logger,
		notify:     make(chan struct{}, 1),
		BatchSize:  50,
		Interval:   30 * time.Second,
		Lease:      time.Minute,
		MaxBackoff: 10 * time.Minute,
		Now:        time.Now,
	}
}

const (
	defaultInlineFlushTimeout   = 8 * time.Second
	defaultShutdownFlushTimeout = 5 * time.Second
)

// InlineFlushEnabled returns true when outbox events should be flushed inline inside the
// active request (default on Cloud Run via K_SERVICE, or explicitly via OUTBOX_INLINE_FLUSH=true).
// This lets Cloud Run scale to zero without per-minute Cloud Scheduler pings.
func InlineFlushEnabled() bool {
	if raw := strings.TrimSpace(os.Getenv("OUTBOX_INLINE_FLUSH")); raw != "" {
		return strings.EqualFold(raw, "true") || raw == "1"
	}
	return strings.TrimSpace(os.Getenv("K_SERVICE")) != ""
}

// Notify pide un Flush sin bloquear (si ya hay uno pedido, no suma otro).
func (r *Relay) Notify() {
	if r == nil || r.notify == nil {
		return
	}
	select {
	case r.notify <- struct{}{}:
	default:
	}
}

// NotifyInline flushes pending events inline (best-effort) when running on Cloud Run
// (or OUTBOX_INLINE_FLUSH=true), falling back to non-blocking async Notify otherwise.
func (r *Relay) NotifyInline() {
	if r == nil {
		return
	}
	if InlineFlushEnabled() {
		r.FlushBestEffort(context.Background(), defaultInlineFlushTimeout)
		return
	}
	r.Notify()
}

// FlushBestEffort attempts to publish pending outbox events within timeout without
// propagating errors to the caller. Pending events remain safely in Firestore for retry.
func (r *Relay) FlushBestEffort(ctx context.Context, timeout time.Duration) int {
	if r == nil || r.client == nil || r.publisher == nil {
		return 0
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout <= 0 {
		timeout = defaultInlineFlushTimeout
	}
	flushCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	n, err := r.Flush(flushCtx)
	if err != nil {
		logger := r.logger
		if logger == nil {
			logger = slog.Default()
		}
		logger.Warn("outbox: best-effort flush failed", slog.String("error", err.Error()))
	}
	return n
}

// Shutdown drains any pending outbox events during graceful shutdown (Cloud Run SIGTERM).
func (r *Relay) Shutdown(ctx context.Context) {
	r.FlushBestEffort(ctx, defaultShutdownFlushTimeout)
}

// Run publica pendientes ante cada Notify y cada Interval, hasta que se
// cancela ctx.
func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.Interval)
	defer ticker.Stop()
	for {
		if _, err := r.Flush(ctx); err != nil && ctx.Err() == nil {
			r.logger.Error("outbox: error publicando pendientes", slog.String("error", err.Error()))
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-r.notify:
		}
	}
}

// Flush publica los eventos vencidos (hasta BatchSize por vuelta, y
// sigue mientras haya). Devuelve cuántos publicó.
func (r *Relay) Flush(ctx context.Context) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	total := 0
	for {
		recs, err := r.claim(ctx)
		if err != nil {
			return total, err
		}
		if len(recs) == 0 {
			return total, nil
		}
		sent := r.publishBatch(ctx, recs)
		total += sent
		if sent == 0 || len(recs) < r.BatchSize {
			return total, nil
		}
	}
}

// claim reserva, en una transacción, los próximos eventos vencidos:
// les corre next_attempt_at hacia adelante (Lease) para que otra
// instancia no los publique a la vez. Si la instancia se cae a mitad de
// camino, vencido el lease otro Relay los vuelve a tomar.
func (r *Relay) claim(ctx context.Context) ([]Record, error) {
	var out []Record
	now := r.Now().UTC()
	err := r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		out = out[:0]
		q := r.client.Collection(Collection).
			Where("next_attempt_at", "<=", now).
			OrderBy("next_attempt_at", firestore.Asc).
			Limit(r.BatchSize)
		docs, err := tx.Documents(q).GetAll()
		if err != nil {
			return err
		}
		for _, d := range docs {
			var rec Record
			if err := d.DataTo(&rec); err != nil {
				return err
			}
			out = append(out, rec)
		}
		for _, d := range docs {
			if err := tx.Update(d.Ref, []firestore.Update{{Path: "next_attempt_at", Value: now.Add(r.Lease)}}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("outbox: no se pudieron reservar pendientes: %w", err)
	}
	return out, nil
}

// publishBatch publica en orden. Si falla un evento con ordering key,
// los siguientes con la misma clave se postergan junto con él, para no
// adelantarse al que falló.
func (r *Relay) publishBatch(ctx context.Context, recs []Record) int {
	sent := 0
	blocked := map[string]time.Time{}
	for _, rec := range recs {
		ref := r.client.Collection(Collection).Doc(rec.EventID)
		if retryAt, ok := blocked[rec.OrderingKey]; ok && rec.OrderingKey != "" {
			r.reschedule(ctx, ref, rec, retryAt, "postergado: falló un evento anterior con la misma ordering key")
			continue
		}
		err := r.publisher.Publish(ctx, rec.Topic, Message{Data: []byte(rec.Data), Attributes: rec.Attributes, OrderingKey: rec.OrderingKey})
		if err != nil {
			retryAt := r.Now().UTC().Add(r.backoff(rec.Attempts + 1))
			if rec.OrderingKey != "" {
				blocked[rec.OrderingKey] = retryAt
			}
			r.reschedule(ctx, ref, rec, retryAt, err.Error())
			r.log(slog.LevelError, rec, "error", err)
			continue
		}
		if _, err := ref.Delete(ctx); err != nil {
			// Se publicó pero no se pudo borrar: se volverá a publicar
			// cuando venza el lease. Los consumidores descartan el
			// duplicado por event_id.
			r.log(slog.LevelWarn, rec, "publicado_sin_borrar", err)
		}
		r.log(slog.LevelInfo, rec, "publicado", nil)
		sent++
	}
	return sent
}

func (r *Relay) reschedule(ctx context.Context, ref *firestore.DocumentRef, rec Record, at time.Time, reason string) {
	_, err := ref.Update(ctx, []firestore.Update{
		{Path: "attempts", Value: rec.Attempts + 1},
		{Path: "last_error", Value: reason},
		{Path: "next_attempt_at", Value: at},
	})
	if err != nil {
		r.log(slog.LevelError, rec, "error_reprogramando", err)
	}
}

// backoff exponencial: 5s, 10s, 20s... hasta MaxBackoff.
func (r *Relay) backoff(attempt int) time.Duration {
	d := 5 * time.Second
	for i := 1; i < attempt && d < r.MaxBackoff; i++ {
		d *= 2
	}
	return min(d, r.MaxBackoff)
}

func (r *Relay) log(level slog.Level, rec Record, result string, err error) {
	attrs := []any{
		slog.String("event_id", rec.EventID),
		slog.String("event_type", rec.EventType),
		slog.String("correlation_id", rec.Attributes[events.AttrCorrelationID]),
		slog.String("topic", rec.Topic),
		slog.Int("intentos", rec.Attempts+1),
		slog.String("resultado", result),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}
	r.logger.Log(context.Background(), level, "outbox", attrs...)
}

// FlushHandler expone Flush por HTTP (POST) para Cloud Scheduler. Se
// protege con el mismo verificador OIDC que los push de Pub/Sub.
func (r *Relay) FlushHandler(verifier events.Verifier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := verifier.Verify(req); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		n, err := r.Flush(req.Context())
		if err != nil {
			r.logger.Error("outbox: flush programado con error", slog.String("error", err.Error()))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"publicados":%d}`, n)
	})
}
