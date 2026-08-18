package middlewares

import "net/http"

// SecurityHeaders agrega cabeceras HTTP defensivas estándar (OWASP)
// para una API JSON: no cachear, no permitir MIME sniffing, no
// permitir que la respuesta se embeba en un iframe, etc.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		next.ServeHTTP(w, r)
	})
}
