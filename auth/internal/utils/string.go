package utils

import (
	"crypto/rand"
	"encoding/hex"
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
