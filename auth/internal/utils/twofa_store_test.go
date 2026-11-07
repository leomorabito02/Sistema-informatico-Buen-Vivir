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

func TestTwofaStore_SingletonAndEdgeCases(t *testing.T) {
	// Singleton
	singleton := GetTwofaStore()
	assert.NotNil(t, singleton)

	store := newTwofaStore()

	// Nonce replacement on same email
	n1, err := store.CreateChallenge("user@example.com", []string{"u1"}, "email", "111111", "", time.Minute)
	require.NoError(t, err)
	n2, err := store.CreateChallenge("user@example.com", []string{"u1"}, "email", "222222", "", time.Minute)
	require.NoError(t, err)
	assert.NotEqual(t, n1, n2)

	// Old nonce should be gone
	_, ok := store.GetChallenge(n1)
	assert.False(t, ok)

	// GetChallengeByEmail
	ch, ok := store.GetChallengeByEmail("user@example.com")
	assert.True(t, ok)
	assert.Equal(t, "222222", ch.Code)

	// Nonexistent email
	_, ok = store.GetChallengeByEmail("nobody@example.com")
	assert.False(t, ok)
	assert.Empty(t, store.GetCodeByEmail("nobody@example.com"))

	// Expired challenge
	expNonce, err := store.CreateChallenge("expired@example.com", []string{"u2"}, "totp", "333333", "secret", -time.Second)
	require.NoError(t, err)
	_, ok = store.GetChallenge(expNonce)
	assert.False(t, ok)
	_, ok = store.GetChallengeByEmail("expired@example.com")
	assert.False(t, ok)
	assert.Empty(t, store.GetCodeByEmail("expired@example.com"))
	assert.False(t, store.UpdateCode(expNonce, "444444", time.Minute))
	_, ok = store.ConsumeChallenge(expNonce)
	assert.False(t, ok)

	// Failed attempt on non-existent nonce
	left, locked := store.RecordFailedAttempt("missing-nonce")
	assert.True(t, locked)
	assert.Equal(t, 0, left)

	// Exceeded failed attempts (lockout at 5)
	lockNonce, err := store.CreateChallenge("lock@example.com", []string{"u3"}, "email", "999999", "", time.Minute)
	require.NoError(t, err)
	for i := 1; i <= 4; i++ {
		left, locked = store.RecordFailedAttempt(lockNonce)
		assert.False(t, locked)
		assert.Equal(t, 5-i, left)
	}
	left, locked = store.RecordFailedAttempt(lockNonce)
	assert.True(t, locked)
	assert.Equal(t, 0, left)

	// Challenge deleted after 5 attempts
	_, ok = store.GetChallenge(lockNonce)
	assert.False(t, ok)
	_, ok = store.GetChallengeByEmail("lock@example.com")
	assert.False(t, ok)
}
