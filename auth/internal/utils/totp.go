package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" // #nosec G505 -- RFC 6238 TOTP specification requires SHA-1
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"net/url"
	"strings"
	"time"
)

const (
	totpStepSeconds = 30
	totpDigits      = 6
)

// GenerateTOTPSecret creates a 20-byte random Base32 string (RFC 6238 / RFC 4226 standard).
func GenerateTOTPSecret() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes), nil
}

// BuildOTPAuthURI constructs an otpauth://totp URI for Authenticator apps (Authy, Google Authenticator).
func BuildOTPAuthURI(issuer, accountName, secret string) string {
	label := fmt.Sprintf("%s:%s", issuer, accountName)
	values := url.Values{}
	values.Set("secret", secret)
	values.Set("issuer", issuer)
	values.Set("algorithm", "SHA1")
	values.Set("digits", fmt.Sprintf("%d", totpDigits))
	values.Set("period", fmt.Sprintf("%d", totpStepSeconds))

	return fmt.Sprintf("otpauth://totp/%s?%s", url.PathEscape(label), values.Encode())
}

// GenerateTOTPCode generates the 6-digit TOTP code for a specific point in time.
func GenerateTOTPCode(secret string, t time.Time) (string, error) {
	cleanedSecret := strings.ToUpper(strings.ReplaceAll(secret, " ", ""))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(cleanedSecret)
	if err != nil {
		// Fallback trying with padding if input had it
		key, err = base32.StdEncoding.DecodeString(cleanedSecret)
		if err != nil {
			return "", fmt.Errorf("invalid base32 secret: %w", err)
		}
	}

	counter := uint64(t.Unix() / totpStepSeconds)
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, counter)

	// #nosec G401 -- HMAC-SHA1 required by RFC 6238
	h := hmac.New(sha1.New, key)
	_, _ = h.Write(counterBytes)
	hash := h.Sum(nil)

	offset := hash[len(hash)-1] & 0x0F
	truncatedHash := binary.BigEndian.Uint32(hash[offset : offset+4])
	truncatedHash &= 0x7FFFFFFF

	codeNum := truncatedHash % uint32(math.Pow10(totpDigits))
	return fmt.Sprintf("%0*d", totpDigits, codeNum), nil
}

// ValidateTOTPCode validates the 6-digit code with +/- 1 time step window for clock skew tolerance.
func ValidateTOTPCode(secret, code string, now time.Time) bool {
	code = strings.TrimSpace(code)
	if len(code) != totpDigits {
		return false
	}

	// Window: -1 step, current step, +1 step
	for stepOffset := -1; stepOffset <= 1; stepOffset++ {
		t := now.Add(time.Duration(stepOffset*totpStepSeconds) * time.Second)
		expectedCode, err := GenerateTOTPCode(secret, t)
		if err == nil && hmac.Equal([]byte(expectedCode), []byte(code)) {
			return true
		}
	}
	return false
}

// GenerateNumericOTP generates a cryptographically secure numeric OTP string of given digits.
func GenerateNumericOTP(digits int) (string, error) {
	if digits <= 0 {
		digits = 6
	}
	maxVal := int64(math.Pow10(digits))
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	val := binary.BigEndian.Uint64(bytes) % uint64(maxVal)
	return fmt.Sprintf("%0*d", digits, val), nil
}
