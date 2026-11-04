package routes

import (
	"net/http"
	"strings"
)

// En Docker el CORS lo pone el gateway (nginx). En Google Cloud no hay
// nginx: el API Gateway pasa la respuesta tal cual, así que el servicio
// tiene que poner los headers. Mismos valores que education,
// human-promotion y campaigns (rs/cors).
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

// WithCORS agrega los headers de CORS para los orígenes permitidos y
// responde el preflight (OPTIONS) sin pasar por las rutas. Sin orígenes
// no hace nada.
func WithCORS(next http.Handler, origins []string) http.Handler {
	if len(origins) == 0 {
		return next
	}
	allowed := make(map[string]bool, len(origins))
	for _, o := range origins {
		allowed[o] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		w.Header().Add("Vary", "Origin")
		if origin == "" || !allowed[origin] {
			next.ServeHTTP(w, r)
			return
		}
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Allow-Credentials", "true")
		h.Set("Access-Control-Expose-Headers", corsExpose)
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			h.Set("Access-Control-Allow-Methods", corsMethods)
			h.Set("Access-Control-Allow-Headers", corsHeaders)
			h.Set("Access-Control-Max-Age", corsMaxAge)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
