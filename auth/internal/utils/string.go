package utils

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"strings"
)

const hexCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRandomString returns a cryptographically secure random alphanumeric string of length n.
func GenerateRandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("utils: failed to generate random string: " + err.Error())
	}
	// Map each byte to a character in the charset using modulo
	result := make([]byte, n)
	for i, v := range b {
		result[i] = hexCharset[int(v)%len(hexCharset)]
	}
	return string(result)
}

// GenerateHexID generates a cryptographically secure random hex string (suitable for IDs).
func GenerateHexID(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic("utils: failed to generate hex ID: " + err.Error())
	}
	return hex.EncodeToString(b)[:n]
}

// ResolveFrontendURL returns the public frontend base URL for email action links,
// avoiding empty or localhost values and selecting production or testing from APP_ENV.
func ResolveFrontendURL() string {
	raw := strings.TrimRight(strings.TrimSpace(os.Getenv("FRONTEND_URL")), "/")
	if raw != "" && !strings.Contains(raw, "localhost") && !strings.Contains(raw, "127.0.0.1") {
		return raw
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV"))) {
	case "production", "prod":
		return "https://ceibuenvivir.org"
	default:
		return "https://testing.ceibuenvivir.pages.dev"
	}
}
