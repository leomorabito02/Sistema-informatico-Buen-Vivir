package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTwofaStore(t *testing.T) {
	store := newTwofaStore()

	nonce, err := store.CreateChallenge("test@example.com", []string{"u1", "u2"}, "email", "123456", "", 5*time.Minute)
	require.NoError(t, err)
	assert.NotEmpty(t, nonce)

	ch, ok := store.GetChallenge(nonce)
	assert.True(t, ok)
	assert.Equal(t, "test@example.com", ch.Email)
	assert.Equal(t, "123456", ch.Code)

	assert.Equal(t, "123456", store.GetCodeByEmail("test@example.com"))

	// Test failed attempts
	left, locked := store.RecordFailedAttempt(nonce)
	assert.False(t, locked)
	assert.Equal(t, 4, left)

	// Update code
	assert.True(t, store.UpdateCode(nonce, "654321", 5*time.Minute))
	assert.Equal(t, "654321", store.GetCodeByEmail("test@example.com"))

	// Consume
	consumed, ok := store.ConsumeChallenge(nonce)
	assert.True(t, ok)
	assert.Equal(t, "654321", consumed.Code)

	// Should not exist after consume
	_, ok = store.GetChallenge(nonce)
	assert.False(t, ok)
}
