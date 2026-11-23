package middlewares

import (
	"time"

	"github.com/go-chi/httprate"
	"net/http"
)

// RateLimiter returns a middleware that limits requests by IP.
// Useful as a fallback for DDoS protection if the API Gateway lets some traffic through.
func RateLimiter() func(next http.Handler) http.Handler {
	// 300 por minuto y por IP: una pantalla del panel dispara muchos pedidos y en el
	// centro varias personas comparten la misma IP.
	return httprate.LimitByIP( //nolint:staticcheck // LimitByIP alcanza para rate limiting básico por IP
		300,           // max requests
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
