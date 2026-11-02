package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const HeaderInternalSignature = "X-Internal-Signature"

func respondHMACError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// RequireInternalHMAC enforces HMAC-SHA256 signature verification over the request body
// for internal service-to-service communication.
func RequireInternalHMAC(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			trimmedSecret := strings.TrimSpace(secret)
			if trimmedSecret == "" {
				next.ServeHTTP(w, r)
				return
			}

			sigHex := strings.TrimSpace(r.Header.Get(HeaderInternalSignature))
			if sigHex == "" {
				respondHMACError(w, http.StatusUnauthorized, "missing X-Internal-Signature header")
				return
			}

			providedSig, err := hex.DecodeString(sigHex)
			if err != nil {
				respondHMACError(w, http.StatusUnauthorized, "invalid signature encoding")
				return
			}

			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				respondHMACError(w, http.StatusBadRequest, "failed to read request body")
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			mac := hmac.New(sha256.New, []byte(trimmedSecret))
			mac.Write(bodyBytes)
			expectedSig := mac.Sum(nil)

			if !hmac.Equal(expectedSig, providedSig) {
				respondHMACError(w, http.StatusUnauthorized, "invalid internal HMAC signature")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
