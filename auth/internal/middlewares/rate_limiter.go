package middlewares

import (
	"time"

	"github.com/go-chi/httprate"
	"net/http"
)

// RateLimiter returns a middleware that limits requests by IP.
// Useful as a fallback for DDoS protection if the API Gateway lets some traffic through.
func RateLimiter() func(next http.Handler) http.Handler {
	return httprate.LimitByIP( //nolint:staticcheck // LimitByIP is valid for basic IP rate limiting
		100,             // max requests
		1*time.Minute,   // per window
	)
}

// StrictRateLimiter is for sensitive endpoints like login
func StrictRateLimiter() func(next http.Handler) http.Handler {
	return httprate.LimitByIP( //nolint:staticcheck // LimitByIP is valid for login rate limiting
		10,              // 10 attempts
		1*time.Minute,   // per minute
	)
}
