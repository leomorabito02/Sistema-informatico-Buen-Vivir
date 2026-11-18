package middlewares

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SetupMiddlewares configures global middlewares for the application
func SetupMiddlewares(r *chi.Mux) {
	// CORS primero: contesta el preflight (OPTIONS) antes que el resto.
	r.Use(CORS(ParseOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))))

	// Standard chi middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP) //nolint:staticcheck // RealIP se usa solo para logging interno; el gateway ya sanea X-Forwarded-For
	r.Use(middleware.Recoverer)

	// Logging middleware adapted for JSON / Google Cloud Logging
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			slog.Info("Incoming Request",
				slog.String("method", req.Method),
				slog.String("url", req.URL.Path),
				slog.String("ip", req.RemoteAddr),
			)
			next.ServeHTTP(w, req)
		})
	})

	// Security Headers (OWASP)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Prevent MIME sniffing (A05)
			w.Header().Set("X-Content-Type-Options", "nosniff")
			// Prevent clickjacking (A05)
			w.Header().Set("X-Frame-Options", "DENY")
			// HSTS: enforce HTTPS for 1 year including subdomains (A02/A05)
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			// CSP: restrict resource origins to known domains (A03/A05)
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"connect-src 'self' http://localhost:4200 https://ceibuenvivir.org https://testing.ceibuenvivir.pages.dev; "+
					"frame-ancestors 'none';")
			// Prevent referrer leakage (A01)
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			// Disable unnecessary browser features (A05)
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			next.ServeHTTP(w, req)
		})
	})
}
