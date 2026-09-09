package middlewares

import (
	"log/slog"
	"net/http"
)

// Recoverer evita que un panic en un handler tumbe todo el proceso: lo
// convierte en un 500 y sigue sirviendo el resto de las requests.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("panic recuperado", slog.Any("error", rec), slog.String("path", r.URL.Path))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"error interno del servidor"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
