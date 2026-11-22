package middlewares

import (
	"time"

	"github.com/go-chi/httprate"
	"net/http"
)

// RateLimiter returns a middleware that limits requests by IP.
// Useful as a fallback for DDoS protection if the API Gateway lets some traffic through.
func RateLimiter() func(next http.Handler) http.Handler {
	// 300 por minuto y por IP (antes 100). Una sola pantalla del panel
	// dispara muchos pedidos (Seguimiento: 2 por curso) y en el centro
	// varias personas salen a internet con la misma IP: con 100 se
	// cortaba todo con 429, incluida la sugerencia de habilidades con IA
	// (visto en testing el 22/11/2026).
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
