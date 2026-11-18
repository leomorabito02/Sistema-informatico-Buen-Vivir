package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"buenvivir/human-promotion/internal/events"
)

// AsyncEndpoints son las suscripciones push de Pub/Sub y el barrido del
// outbox (Cloud Scheduler). Los que quedan en nil no se exponen.
type AsyncEndpoints struct {
	Educacion   http.Handler
	Usuarios    http.Handler
	OutboxFlush http.Handler
}

// WithAsync monta los endpoints asíncronos POR FUERA del router de la
// API (sin CORS, JWT de usuario ni rate limit por IP: se autentican con
// el token OIDC de Google). El correlation id se aplica a todo.
func WithAsync(api http.Handler, a AsyncEndpoints) http.Handler {
	mux := http.NewServeMux()
	mount := func(pattern string, h http.Handler) {
		if h != nil {
			mux.Handle(pattern, middleware.Recoverer(h))
		}
	}
	mount("POST /eventos/educacion", a.Educacion)
	mount("POST /eventos/usuarios", a.Usuarios)
	mount("POST /internal/outbox/flush", a.OutboxFlush)
	mux.Handle("/", api)
	return events.CorrelationMiddleware(mux)
}
