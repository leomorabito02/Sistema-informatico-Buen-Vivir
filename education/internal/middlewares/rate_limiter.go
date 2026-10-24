package middlewares

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/httprate"
)

// RateLimiter returns a middleware that limits requests by IP.
// Useful as a fallback for DDoS protection if the API Gateway lets some traffic through.
func RateLimiter() func(next http.Handler) http.Handler {
	if os.Getenv("RATE_LIMIT_DISABLED") == "true" {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	return httprate.LimitByIP( //nolint:staticcheck // LimitByIP alcanza para rate limiting básico por IP
		100,           // max requests
		1*time.Minute, // per window
	)
}

// StrictRateLimiter is for sensitive endpoints like login
func StrictRateLimiter() func(next http.Handler) http.Handler {
	return httprate.LimitByIP( //nolint:staticcheck // LimitByIP alcanza para rate limiting de login
		10,            // 10 attempts
		1*time.Minute, // per minute
	)
}
