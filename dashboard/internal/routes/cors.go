package routes

import (
	"net/http"
	"strings"
)

// CORS hecho a mano, con el mismo criterio que auth
// (auth/internal/middlewares/security_middleware.go).
//
// No se usa rs/cors porque el API Gateway de Google le saca al preflight el
// header Access-Control-Request-Method antes de mandarlo a Cloud Run: rs/cors
// no lo reconoce como preflight, lo pasa a las rutas y chi contesta 405. Por
// eso todo OPTIONS se contesta acá, sin llegar a las rutas.
//
// Orígenes permitidos: los de CORS_ALLOWED_ORIGINS más una lista fija (la
// misma de auth), así testing (*.ceibuenvivir.pages.dev) y producción
// funcionan aunque el secret no los tenga. Un origen que no está en ninguna
// de las dos no recibe headers de CORS y el navegador bloquea la respuesta.
const (
	corsMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsHeaders = "Accept, Authorization, Content-Type, X-CSRF-Token, X-Requested-With, X-Skip-Global-Error, X-Redirect-On-404, X-Skip-500-Redirect"
	corsExpose  = "Link, Set-Cookie"
	corsMaxAge  = "300"
)

// ParseOrigins arma la lista de CORS_ALLOWED_ORIGINS (separada por comas).
func ParseOrigins(raw string) []string {
	var out []string
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			out = append(out, o)
		}
	}
	return out
}

// IsAllowedOrigin dice si el origen está en la lista del entorno o en la
// lista fija del proyecto.
func IsAllowedOrigin(origin string, envOrigins []string) bool {
	if origin == "" {
		return false
	}
	for _, o := range envOrigins {
		if o == origin {
			return true
		}
	}
	return origin == "http://localhost:4200" ||
		origin == "http://localhost:8080" ||
		origin == "https://ceibuenvivir.org" ||
		strings.HasSuffix(origin, ".ceibuenvivir.pages.dev") ||
		strings.HasSuffix(origin, ".ceibuenvivir.org")
}

// CORS agrega los headers para los orígenes permitidos y contesta 200 a
// todo OPTIONS sin pasar por las rutas.
func CORS(envOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Add("Vary", "Origin")
			if origin := r.Header.Get("Origin"); IsAllowedOrigin(origin, envOrigins) {
				h.Set("Access-Control-Allow-Origin", origin)
				h.Set("Access-Control-Allow-Credentials", "true")
				h.Set("Access-Control-Allow-Methods", corsMethods)
				h.Set("Access-Control-Allow-Headers", corsHeaders)
				h.Set("Access-Control-Max-Age", corsMaxAge)
				if corsExpose != "" {
					h.Set("Access-Control-Expose-Headers", corsExpose)
				}
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// WithCORS envuelve el mux del servidor con CORS. En Docker el CORS lo pone
// también nginx; en Google Cloud no hay nginx y lo pone el servicio.
func WithCORS(next http.Handler, origins []string) http.Handler {
	return CORS(origins)(next)
}
