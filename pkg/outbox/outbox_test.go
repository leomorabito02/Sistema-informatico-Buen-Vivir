package outbox

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/firestore"

	"buenvivir/pkg/events"
)

// Proyecto propio del emulador para que estos tests no choquen con los
// de otros paquetes que corren en paralelo contra el mismo emulador.
const testProject = "outbox-test"

// newTestFirestore se conecta al emulador de Firestore
// (FIRESTORE_EMULATOR_HOST) o saltea el test, igual que el resto de los
// tests de repositorios del proyecto.
func newTestFirestore(t *testing.T) *firestore.Client {
	t.Helper()
	host := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if host == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST no seteado: se omiten los tests contra el emulador de Firestore")
	}
	conn, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		t.Skipf("emulador de Firestore no accesible en %s: %v", host, err)
	}
	_ = conn.Close()
	client, err := firestore.NewClient(context.Background(), testProject)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	wipe(t, client, Collection)
	wipe(t, client, ProcessedCollection)
	return client
}

func wipe(t *testing.T, client *firestore.Client, name string) {
	t.Helper()
	docs, err := client.Collection(name).Documents(context.Background()).GetAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range docs {
		if _, err := d.Ref.Delete(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
}

func pending(t *testing.T, client *firestore.Client) []Record {
	t.Helper()
	docs, err := client.Collection(Collection).Documents(context.Background()).GetAll()
	if err != nil {
		t.Fatal(err)
	}
	out := make([]Record, 0, len(docs))
	for _, d := range docs {
		var r Record
		if err := d.DataTo(&r); err != nil {
			t.Fatal(err)
		}
		out = append(out, r)
	}
	return out
}

type fakePublisher struct {
	mu   sync.Mutex
	got  []Message
	fail map[string]bool // ordering keys que fallan
	sent chan struct{}
}

func (p *fakePublisher) Publish(_ context.Context, _ string, msg Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.fail[msg.OrderingKey] {
		return errors.New("pub/sub no disponible")
	}
	p.got = append(p.got, msg)
	if p.sent != nil {
		select {
		case p.sent <- struct{}{}:
		default:
		}
	}
	return nil
}

func quiet() *slog.Logger { return slog.New(slog.NewJSONHandler(io.Discard, nil)) }

func TestWriterNilYSinTopic(t *testing.T) {
	var w *Writer
	if err := w.Add(context.Background(), Event{Topic: "t", Type: "x"}); err != nil {
		t.Fatal("un writer nil no hace nada")
	}
	if err := w.AddTx(context.Background(), nil, Event{}); err != nil {
		t.Fatal("un writer nil no hace nada")
	}
	w.Notify()
	if _, err := (&Writer{producer: "p"}).record(context.Background(), Event{Type: "x", Payload: 1}); err == nil {
		t.Fatal("sin topic debe fallar")
	}
	if _, err := (&Writer{producer: ""}).record(context.Background(), Event{Topic: "t", Type: "x", Payload: 1}); err == nil {
		t.Fatal("un envelope sin producer es inválido")
	}
}

func TestWriterAddTxYAdd(t *testing.T) {
	client := newTestFirestore(t)
	ctx := events.WithActor(events.WithCorrelationID(context.Background(), "corr-9"), "admin-1")
	notified := 0
	w := NewWriter(client, events.ProducerEducacion, func() { notified++ })

	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		return w.AddTx(ctx, tx, Event{Topic: "educacion-eventos", Type: events.InscripcionCreada, Payload: events.InscripcionCreadaPayload{InscripcionID: "i1"}, OrderingKey: "p1"})
	})
	if err != nil {
		t.Fatal(err)
	}
	id := events.NameUUID("mismo-hecho")
	ev := Event{Topic: "contenido-eventos", Type: events.ContenidoPublicado, Payload: events.ContenidoPayload{EntryID: "e1"}, ID: id}
	if err := w.Add(ctx, ev); err != nil {
		t.Fatal(err)
	}
	if err := w.Add(ctx, ev); err != nil {
		t.Fatalf("el mismo id determinístico se descarta sin error: %v", err)
	}

	recs := pending(t, client)
	if len(recs) != 2 {
		t.Fatalf("se esperaban 2 pendientes, hay %d", len(recs))
	}
	for _, r := range recs {
		env, err := events.Decode([]byte(r.Data))
		if err != nil {
			t.Fatalf("el outbox guarda un envelope válido: %v", err)
		}
		if env.CorrelationID != "corr-9" || env.ActorID != "admin-1" || r.Attributes[events.AttrEventType] != r.EventType {
			t.Errorf("envelope/atributos incorrectos: %+v %+v", env, r.Attributes)
		}
	}
	if notified != 2 {
		t.Errorf("Add debe avisar al relay: %d", notified)
	}
}

func TestRelayPublicaYBorra(t *testing.T) {
	client := newTestFirestore(t)
	w := NewWriter(client, events.ProducerEducacion, nil)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := w.Add(ctx, Event{Topic: "t", Type: events.CursoCreado, Payload: events.CursoPayload{CursoID: "c"}}); err != nil {
			t.Fatal(err)
		}
	}
	pub := &fakePublisher{}
	relay := NewRelay(client, pub, quiet())
	relay.BatchSize = 2 // fuerza dos vueltas
	n, err := relay.Flush(ctx)
	if err != nil || n != 3 {
		t.Fatalf("publicados %d, err %v", n, err)
	}
	if len(pending(t, client)) != 0 {
		t.Fatal("lo publicado se borra del outbox")
	}
}

func TestRelayReintentaYRespetaOrderingKey(t *testing.T) {
	client := newTestFirestore(t)
	w := NewWriter(client, events.ProducerEducacion, nil)
	ctx := context.Background()
	for _, key := range []string{"p1", "p1", "p2"} {
		if err := w.Add(ctx, Event{Topic: "t", Type: events.AsistenciaRegistrada, Payload: events.AsistenciaRegistradaPayload{ParticipanteID: key}, OrderingKey: key}); err != nil {
			t.Fatal(err)
		}
		time.Sleep(5 * time.Millisecond) // created_at distintos: orden estable
	}
	pub := &fakePublisher{fail: map[string]bool{"p1": true}}
	relay := NewRelay(client, pub, quiet())
	n, err := relay.Flush(ctx)
	if err != nil || n != 1 {
		t.Fatalf("solo p2 se publica: %d, %v", n, err)
	}
	recs := pending(t, client)
	if len(recs) != 2 {
		t.Fatalf("los de p1 quedan pendientes: %d", len(recs))
	}
	for _, r := range recs {
		if r.Attempts != 1 || r.LastError == "" || !r.NextAttemptAt.After(time.Now()) {
			t.Errorf("reintento mal programado: %+v", r)
		}
	}
	// Antes del backoff no se reintenta.
	if n, _ := relay.Flush(ctx); n != 0 {
		t.Fatalf("no debía reintentar antes de tiempo: %d", n)
	}
	// Pasado el backoff, y con Pub/Sub de vuelta, sale en orden.
	pub.fail = nil
	relay.Now = func() time.Time { return time.Now().Add(time.Hour) }
	if n, err := relay.Flush(ctx); err != nil || n != 2 {
		t.Fatalf("reintento: %d, %v", n, err)
	}
}

func TestRelayBackoff(t *testing.T) {
	r := NewRelay(nil, nil, nil)
	if r.backoff(1) != 5*time.Second || r.backoff(3) != 20*time.Second || r.backoff(50) != r.MaxBackoff {
		t.Errorf("backoff: %v %v %v", r.backoff(1), r.backoff(3), r.backoff(50))
	}
}

func TestRelayRunPublicaAlNotificar(t *testing.T) {
	client := newTestFirestore(t)
	pub := &fakePublisher{sent: make(chan struct{}, 1)}
	relay := NewRelay(client, pub, quiet())
	relay.Interval = time.Hour
	w := NewWriter(client, events.ProducerEducacion, relay.Notify)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { relay.Run(ctx); close(done) }()

	if err := w.Add(context.Background(), Event{Topic: "t", Type: events.CursoEliminado, Payload: events.CursoEliminadoPayload{CursoID: "c"}}); err != nil {
		t.Fatal(err)
	}
	relay.Notify() // un segundo aviso no bloquea
	select {
	case <-pub.sent:
	case <-time.After(10 * time.Second):
		t.Fatal("el relay no publicó después del aviso")
	}
	cancel()
	<-done
}

func TestFlushHandler(t *testing.T) {
	client := newTestFirestore(t)
	relay := NewRelay(client, &fakePublisher{}, quiet())
	h := relay.FlushHandler(events.NoVerifier{})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	relay.FlushHandler(denyAll{}).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("sin autorización: %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"publicados":0`) {
		t.Errorf("flush: %d %s", rec.Code, rec.Body.String())
	}
}

type denyAll struct{}

func (denyAll) Verify(*http.Request) error { return events.ErrUnauthorized }

func TestFirestoreProcessedStore(t *testing.T) {
	client := newTestFirestore(t)
	s := NewFirestoreProcessedStore(client)
	ctx := context.Background()
	if ok, err := s.IsProcessed(ctx, "c", "e1"); ok || err != nil {
		t.Fatalf("nuevo: %v %v", ok, err)
	}
	if err := s.MarkProcessed(ctx, "c", "e1", "t"); err != nil {
		t.Fatal(err)
	}
	if ok, _ := s.IsProcessed(ctx, "c", "e1"); !ok {
		t.Fatal("debe quedar marcado")
	}
	if ok, _ := s.IsProcessed(ctx, "otro", "e1"); ok {
		t.Fatal("cada consumidor lleva su cuenta")
	}
	snap, err := client.Collection(ProcessedCollection).Doc(processedID("c", "e1")).Get(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var doc processedDoc
	_ = snap.DataTo(&doc)
	if doc.ExpiresAt.Sub(doc.ProcessedAt) != ProcessedTTL {
		t.Errorf("TTL: %v", doc.ExpiresAt.Sub(doc.ProcessedAt))
	}
}

func TestInlineFlushEnabled(t *testing.T) {
	t.Run("disabled by default outside Cloud Run", func(t *testing.T) {
		// Arrange
		t.Setenv("K_SERVICE", "")
		t.Setenv("OUTBOX_INLINE_FLUSH", "")

		// Act & Assert
		if InlineFlushEnabled() {
			t.Fatal("expected inline flush disabled by default")
		}
	})

	t.Run("enabled automatically on Cloud Run", func(t *testing.T) {
		// Arrange
		t.Setenv("K_SERVICE", "education-testing")
		t.Setenv("OUTBOX_INLINE_FLUSH", "")

		// Act & Assert
		if !InlineFlushEnabled() {
			t.Fatal("expected inline flush enabled when K_SERVICE is set")
		}
	})

	t.Run("explicit false overrides K_SERVICE", func(t *testing.T) {
		// Arrange
		t.Setenv("K_SERVICE", "education-testing")
		t.Setenv("OUTBOX_INLINE_FLUSH", "false")

		// Act & Assert
		if InlineFlushEnabled() {
			t.Fatal("expected explicit false to disable inline flush")
		}
	})

	t.Run("explicit true enables locally", func(t *testing.T) {
		// Arrange
		t.Setenv("K_SERVICE", "")
		t.Setenv("OUTBOX_INLINE_FLUSH", "true")

		// Act & Assert
		if !InlineFlushEnabled() {
			t.Fatal("expected explicit true to enable inline flush")
		}
	})
}

func TestRelayFlushBestEffortAndNotifyInline_NilSafe(t *testing.T) {
	// Arrange
	var nilRelay *Relay
	zeroRelay := &Relay{}
	emptyRelay := NewRelay(nil, nil, quiet())

	// Act & Assert
	if n := nilRelay.FlushBestEffort(context.Background(), 0); n != 0 {
		t.Fatalf("nil relay expected 0, got %d", n)
	}
	if n := zeroRelay.FlushBestEffort(nil, 0); n != 0 {
		t.Fatalf("zero relay expected 0, got %d", n)
	}
	nilRelay.Notify()
	nilRelay.NotifyInline()
	nilRelay.Shutdown(context.Background())

	t.Setenv("K_SERVICE", "education-testing")
	t.Setenv("OUTBOX_INLINE_FLUSH", "")
	emptyRelay.NotifyInline()
	emptyRelay.Shutdown(context.Background())

	t.Setenv("K_SERVICE", "")
	emptyRelay.NotifyInline()
}
