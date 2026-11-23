package middlewares

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// isAllowedOrigin validates an origin against a static allowlist and the env-based list.
func isAllowedOrigin(origin string, envOrigins []string) bool {
	if origin == "" {
		return false
	}
	for _, o := range envOrigins {
		if o == origin || o == "*" {
			return true
		}
	}
	return origin == "http://localhost:4200" ||
		origin == "http://localhost:8080" ||
		origin == "https://ceibuenvivir.org" ||
		strings.HasSuffix(origin, ".ceibuenvivir.pages.dev") ||
		strings.HasSuffix(origin, ".ceibuenvivir.org")
}

// corsMiddleware contesta todo OPTIONS antes del ruteo y agrega los headers de CORS.
// Reemplaza rs/cors porque el API Gateway le saca al preflight Access-Control-Request-Method.
func corsMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			origin := req.Header.Get("Origin")

			if origin != "" && isAllowedOrigin(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Requested-With, X-Skip-Global-Error, X-Redirect-On-404, X-Skip-500-Redirect")
				w.Header().Set("Access-Control-Expose-Headers", "Link, Set-Cookie")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "300")
			}

			// Short-circuit OPTIONS — return 200 immediately without calling next.
			// This prevents chi's MethodNotAllowed handler from ever seeing the request.
			if req.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

// OptionsInterceptor wraps an http.Handler and intercepts ALL OPTIONS requests
// at the net/http level, before chi's routing tree is evaluated.
func OptionsInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodOptions {
			slog.Info("OptionsInterceptor: intercepted preflight",
				slog.String("path", req.URL.Path),
				slog.String("origin", req.Header.Get("Origin")),
			)
			origin := req.Header.Get("Origin")
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Requested-With, X-Skip-Global-Error, X-Redirect-On-404, X-Skip-500-Redirect")
				w.Header().Set("Access-Control-Expose-Headers", "Link, Set-Cookie")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Max-Age", "300")
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, req)
	})
}

// SetupMiddlewares configura los middlewares globales. Los OPTIONS los contesta antes
// OptionsInterceptor, a nivel de http.Server.
func SetupMiddlewares(r *chi.Mux) {
	allowedOriginsEnv := os.Getenv("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsEnv != "" {
		for _, o := range strings.Split(allowedOriginsEnv, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP) //nolint:staticcheck // RealIP used for internal logging
	r.Use(middleware.Recoverer)

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
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
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

	// corsMiddleware agrega los headers de CORS a las respuestas normales.
	r.Use(corsMiddleware(allowedOrigins))
}
