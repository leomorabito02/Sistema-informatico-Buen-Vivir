package utils_test

import (
	"testing"
	"time"

	"auth/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordComplexity(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"Short", "Pass1!", false},
		{"No uppercase", "password123!", false},
		{"No special", "Password123", false},
		{"Exact8Chars", "Pass!123", true},
		{"Valid", "Password123!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := utils.ValidatePasswordComplexity(tt.password)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestPasswordHashing(t *testing.T) {
	password := "SecurePass123!"
	
	hash, err := utils.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	
	assert.True(t, utils.CheckPasswordHash(password, hash))
	assert.False(t, utils.CheckPasswordHash("wrongpass", hash))
}

func TestHashToken(t *testing.T) {
	token := "some-random-token"
	hash1 := utils.HashToken(token)
	hash2 := utils.HashToken(token)
	
	assert.NotEmpty(t, hash1)
	assert.Equal(t, hash1, hash2, "Hashing should be deterministic")
}

func TestGenerateRandomString(t *testing.T) {
	str := utils.GenerateRandomString(10)
	assert.Len(t, str, 10)
	
	str2 := utils.GenerateRandomString(10)
	assert.NotEqual(t, str, str2, "Random strings should not match")
}

func TestNonceStore(t *testing.T) {
	store := utils.GetNonceStore()
	require.NotNil(t, store)

	nonce, err := store.Generate([]string{"usr-1"}, 1*time.Minute)
	require.NoError(t, err)
	assert.NotEmpty(t, nonce)

	assert.False(t, store.Consume(nonce, "usr-2"), "Should fail for wrong user")
	assert.True(t, store.Consume(nonce, "usr-1"), "Should succeed for correct user")
	assert.False(t, store.Consume(nonce, "usr-1"), "Should fail on second consume (single use)")
}
