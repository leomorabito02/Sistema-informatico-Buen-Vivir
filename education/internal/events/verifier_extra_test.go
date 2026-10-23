package events

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFetchKeys_EquivalenceClasses tests OIDCVerifier.fetchKeys handling of HTTP status, malformed JSON, and non-RSA keys.
func TestFetchKeys_EquivalenceClasses(t *testing.T) {
	t.Run("HTTP error returns error", func(t *testing.T) {
		// Arrange
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		v := &OIDCVerifier{CertsURL: srv.URL, HTTPClient: srv.Client()}

		// Act
		keys, err := v.fetchKeys(context.Background())

		// Assert
		assert.Error(t, err)
		assert.Nil(t, keys)
	})

	t.Run("Malformed JSON returns error", func(t *testing.T) {
		// Arrange
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{invalid-json`))
		}))
		defer srv.Close()

		v := &OIDCVerifier{CertsURL: srv.URL, HTTPClient: srv.Client()}

		// Act
		keys, err := v.fetchKeys(context.Background())

		// Assert
		assert.Error(t, err)
		assert.Nil(t, keys)
	})

	t.Run("Non-RSA and invalid base64 keys are skipped", func(t *testing.T) {
		// Arrange
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"keys":[{"kid":"k1","kty":"EC"},{"kid":"k2","kty":"RSA","n":"invalid!base64","e":"AQAB"}]}`))
		}))
		defer srv.Close()

		v := &OIDCVerifier{CertsURL: srv.URL, HTTPClient: srv.Client()}

		// Act
		keys, err := v.fetchKeys(context.Background())

		// Assert
		require.NoError(t, err)
		assert.Empty(t, keys)
	})
}

// TestNewWithID_UnserializablePayload tests payload serialization error boundary.
func TestNewWithID_UnserializablePayload(t *testing.T) {
	// Arrange
	ctx := context.Background()

	// Act
	_, err := NewWithID(ctx, NewUUID(), ProducerEducacion, CursoCreado, make(chan int))

	// Assert
	assert.Error(t, err)
}

// TestIsUUID_EquivalenceClassesAndBoundaries tests uppercase hex and hyphen boundary checks.
func TestIsUUID_EquivalenceClassesAndBoundaries(t *testing.T) {
	// Arrange & Act & Assert
	assert.True(t, IsUUID("123E4567-E89B-12D3-A456-426614174000"), "uppercase hex UUID is valid")
	assert.False(t, IsUUID("12345678-1234a1234-1234-123456789012"), "missing hyphen at pos 13")
	assert.False(t, IsUUID("12345678-1234-1234a1234-123456789012"), "missing hyphen at pos 18")
	assert.False(t, IsUUID("12345678-1234-1234-1234a123456789012"), "missing hyphen at pos 23")
}

// TestEnvelope_Validate_EquivalenceClasses tests each validation rule in Envelope.Validate.
func TestEnvelope_Validate_EquivalenceClasses(t *testing.T) {
	validEnv := func() Envelope {
		return Envelope{
			EventID:       NewUUID(),
			EventType:     CursoCreado,
			Version:       1,
			OccurredAt:    time.Now().UTC(),
			Producer:      ProducerEducacion,
			CorrelationID: NewUUID(),
			Payload:       []byte(`{"cursoId":"c1"}`),
		}
	}

	t.Run("Valid envelope passes validation", func(t *testing.T) {
		// Arrange
		env := validEnv()
		// Act & Assert
		assert.NoError(t, env.Validate())
	})

	t.Run("Empty event type fails", func(t *testing.T) {
		// Arrange
		env := validEnv()
		env.EventType = "   "
		// Act & Assert
		assert.Error(t, env.Validate())
	})

	t.Run("Version less than 1 fails", func(t *testing.T) {
		// Arrange
		env := validEnv()
		env.Version = 0
		// Act & Assert
		assert.Error(t, env.Validate())
	})

	t.Run("Zero OccurredAt fails", func(t *testing.T) {
		// Arrange
		env := validEnv()
		env.OccurredAt = time.Time{}
		// Act & Assert
		assert.Error(t, env.Validate())
	})

	t.Run("Empty producer fails", func(t *testing.T) {
		// Arrange
		env := validEnv()
		env.Producer = ""
		// Act & Assert
		assert.Error(t, env.Validate())
	})

	t.Run("Invalid event ID fails", func(t *testing.T) {
		// Arrange
		env := validEnv()
		env.EventID = "invalid-uuid"
		// Act & Assert
		assert.Error(t, env.Validate())
	})

	t.Run("Empty correlation ID fails", func(t *testing.T) {
		// Arrange
		env := validEnv()
		env.CorrelationID = "   "
		// Act & Assert
		assert.Error(t, env.Validate())
	})

	t.Run("Empty or null payload fails", func(t *testing.T) {
		// Arrange
		env1 := validEnv()
		env1.Payload = nil
		env2 := validEnv()
		env2.Payload = []byte("null")

		// Act & Assert
		assert.Error(t, env1.Validate())
		assert.Error(t, env2.Validate())
	})
}
