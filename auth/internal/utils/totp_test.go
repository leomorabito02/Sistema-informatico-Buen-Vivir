package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTOTPGenerationAndValidation(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	require.NoError(t, err)
	assert.NotEmpty(t, secret)

	now := time.Now()
	code, err := GenerateTOTPCode(secret, now)
	require.NoError(t, err)
	assert.Len(t, code, 6)

	assert.True(t, ValidateTOTPCode(secret, code, now))
	// Test time skew: 25 seconds earlier (same or adjacent step)
	assert.True(t, ValidateTOTPCode(secret, code, now.Add(25*time.Second)))
	// Test time skew: 30 seconds later (adjacent step)
	assert.True(t, ValidateTOTPCode(secret, code, now.Add(30*time.Second)))
	// Test invalid code
	assert.False(t, ValidateTOTPCode(secret, "000000", now.Add(2*time.Minute)))

	uri := BuildOTPAuthURI("BuenVivir", "user@test.org", secret)
	assert.Contains(t, uri, "otpauth://totp/BuenVivir:user@test.org")
	assert.Contains(t, uri, "secret="+secret)
	assert.Contains(t, uri, "issuer=BuenVivir")
}

func TestNumericOTP(t *testing.T) {
	otp, err := GenerateNumericOTP(6)
	require.NoError(t, err)
	assert.Len(t, otp, 6)
	for _, c := range otp {
		assert.True(t, c >= '0' && c <= '9')
	}
}
