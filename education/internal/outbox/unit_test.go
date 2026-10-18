package outbox

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockVerifier struct {
	mock.Mock
}

func (m *mockVerifier) Verify(r *http.Request) error {
	args := m.Called(r)
	return args.Error(0)
}

// TestRelay_Backoff_BoundaryValues tests exponential backoff calculations and max backoff capping.
func TestRelay_Backoff_BoundaryValues(t *testing.T) {
	r := &Relay{
		MaxBackoff: 10 * time.Minute,
	}

	tests := []struct {
		name     string
		attempts int
		expected time.Duration
	}{
		{"Boundary: attempt 0 or 1", 1, 5 * time.Second},
		{"Nominal: attempt 2", 2, 10 * time.Second},
		{"Nominal: attempt 3", 3, 20 * time.Second},
		{"Boundary: high attempt capped at MaxBackoff", 20, 10 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := r.backoff(tt.attempts)
			assert.Equal(t, tt.expected, delay)
		})
	}
}

// TestRelay_Notify_NonBlocking verifies non-blocking behavior of Notify channel.
func TestRelay_Notify_NonBlocking(t *testing.T) {
	r := &Relay{
		notify: make(chan struct{}, 1),
	}

	r.Notify()
	r.Notify() // Second call must not block even if buffer is full

	select {
	case <-r.notify:
	default:
		t.Fatal("expected message on notify channel")
	}
}

// TestFlushHandler_DecisionTable tests method and verifier validation.
func TestFlushHandler_DecisionTable(t *testing.T) {
	relay := &Relay{}

	decisionMatrix := []struct {
		name         string
		method       string
		verifierErr  error
		expectedCode int
	}{
		{
			name:         "GET method returns 405 MethodNotAllowed",
			method:       http.MethodGet,
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "POST with invalid auth verification returns 401 Unauthorized",
			method:       http.MethodPost,
			verifierErr:  errors.New("invalid signature"),
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range decisionMatrix {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			verifier := new(mockVerifier)
			if tt.method == http.MethodPost {
				verifier.On("Verify", mock.Anything).Return(tt.verifierErr).Once()
			}
			handler := relay.FlushHandler(verifier)
			req := httptest.NewRequest(tt.method, "/internal/outbox/flush", nil)
			rec := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tt.expectedCode, rec.Code)
		})
	}
}

// TestWriter_NilGuardAndNotify tests safe handling when writer is nil or has notification callback.
func TestWriter_NilGuardAndNotify(t *testing.T) {
	t.Run("Nil Writer methods are safe no-ops", func(t *testing.T) {
		var w *Writer
		assert.NoError(t, w.AddTx(context.Background(), nil, Event{}))
		assert.NoError(t, w.Add(context.Background(), Event{}))
		w.Notify() // Must not panic
	})

	t.Run("Writer Notify triggers callback when non-nil", func(t *testing.T) {
		called := false
		w := NewWriter(nil, "producer-test", func() {
			called = true
		})

		w.Notify()
		assert.True(t, called)
	})

	t.Run("Writer Add and AddTx return error on invalid event record", func(t *testing.T) {
		// Arrange
		w := NewWriter(nil, "producer-test", nil)
		badEvent := Event{Topic: "", Type: "test.event"}

		// Act
		errAdd := w.Add(context.Background(), badEvent)
		errAddTx := w.AddTx(context.Background(), nil, badEvent)

		// Assert
		assert.Error(t, errAdd)
		assert.Error(t, errAddTx)
	})

	t.Run("Writer Add and AddTx with empty events succeed safely", func(t *testing.T) {
		// Arrange
		called := false
		w := NewWriter(nil, "producer-test", func() { called = true })

		// Act
		errAdd := w.Add(context.Background())
		errAddTx := w.AddTx(context.Background(), nil)

		// Assert
		assert.NoError(t, errAdd)
		assert.NoError(t, errAddTx)
		assert.True(t, called)
	})

	t.Run("Record validation: Event without topic returns error", func(t *testing.T) {
		w := NewWriter(nil, "producer-test", nil)
		_, err := w.record(context.Background(), Event{
			Topic: "",
			Type:  "test.event",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sin topic configurado")
	})

	t.Run("Record with custom ID and valid payload", func(t *testing.T) {
		w := NewWriter(nil, "producer-test", nil)
		uuid := "a0000000-0000-0000-0000-000000000001"
		rec, err := w.record(context.Background(), Event{
			ID:          uuid,
			Topic:       "topic-a",
			Type:        "test.event",
			Payload:     map[string]string{"foo": "bar"},
			OrderingKey: "key-1",
		})
		require.NoError(t, err)
		assert.Equal(t, uuid, rec.EventID)
		assert.Equal(t, "topic-a", rec.Topic)
		assert.Equal(t, "test.event", rec.EventType)
		assert.Equal(t, "key-1", rec.OrderingKey)
		assert.Contains(t, rec.Data, "bar")
	})

	t.Run("Record with empty ID generates UUID", func(t *testing.T) {
		w := NewWriter(nil, "producer-test", nil)
		rec, err := w.record(context.Background(), Event{
			Topic:   "topic-a",
			Type:    "test.event",
			Payload: map[string]string{"k": "v"},
		})
		require.NoError(t, err)
		assert.NotEmpty(t, rec.EventID)
	})

	t.Run("Record with unmarshalable payload returns error", func(t *testing.T) {
		w := NewWriter(nil, "producer-test", nil)
		_, err := w.record(context.Background(), Event{
			Topic:   "topic-a",
			Type:    "test.event",
			Payload: make(chan int),
		})
		require.Error(t, err)
	})
}

// TestRelay_Log tests logger execution with both nil and non-nil error.
func TestRelay_Log(t *testing.T) {
	r := &Relay{
		logger: slog.Default(),
	}
	rec := Record{
		EventID:   "e1",
		EventType: "type1",
		Topic:     "top1",
		Attempts:  1,
	}
	r.log(slog.LevelInfo, rec, "exito", nil)
	r.log(slog.LevelError, rec, "fallo", assert.AnError)
}

// TestProcessedID_DeterministicID verifies formatted ID generation.
func TestProcessedID_DeterministicID(t *testing.T) {
	id := processedID("consumer-a", "event-123")
	assert.Equal(t, "consumer-a_event-123", id)
}
