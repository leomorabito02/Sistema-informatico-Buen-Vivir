package middlewares

import (
	"net/http"
	"strings"
)

// NewCORS arma un middleware de CORS manual (sin depender de una
// librería externa como rs/cors) a partir de una lista de orígenes
// permitidos separados por coma (CORS_ALLOWED_ORIGINS). Refleja el
// origen de la request si está en la lista, para poder permitir
// credenciales (cookies) sin usar el comodín "*".
func NewCORS(allowedOriginsCSV string) Middleware {
	allowed := make(map[string]bool)
	for _, origin := range strings.Split(allowedOriginsCSV, ",") {
		if trimmed := strings.TrimSpace(origin); trimmed != "" {
			allowed[trimmed] = true
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Vary", "Origin")
			}

			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Requested-With")
				w.Header().Set("Access-Control-Max-Age", "300")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
