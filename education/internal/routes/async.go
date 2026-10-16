package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"

	"buenvivir/education/internal/events"
)

// AsyncEndpoints son los endpoints de mensajería asíncrona. Los que
// quedan en nil no se exponen (por ejemplo, sin configuración de
// Pub/Sub en desarrollo local).
type AsyncEndpoints struct {
	// ContentfulWebhook recibe los webhooks de Contentful (se autentica
	// con el secreto del header X-Webhook-Secret).
	ContentfulWebhook http.Handler
	// Usuarios y Contenido son las suscripciones push de Pub/Sub; y
	// OutboxFlush, el barrido del outbox que dispara Cloud Scheduler.
	// Los tres se autentican con el token OIDC de Google.
	Usuarios    http.Handler
	Contenido   http.Handler
	OutboxFlush http.Handler
}

// WithAsync arma el handler raíz del servicio. Los endpoints asíncronos
// van POR FUERA del router de la API: no les corresponden ni el CORS, ni
// el JWT de usuario, ni el rate limit por IP (Pub/Sub manda los push
// desde IPs compartidas de Google y el límite de 100/min los frenaría).
// El correlation id (header X-Correlation-ID) se aplica a todo.
func WithAsync(api http.Handler, a AsyncEndpoints) http.Handler {
	mux := http.NewServeMux()
	mount := func(pattern string, h http.Handler) {
		if h != nil {
			mux.Handle(pattern, middleware.Recoverer(h))
		}
	}
	mount("POST /webhooks/contentful", a.ContentfulWebhook)
	mount("POST /eventos/usuarios", a.Usuarios)
	mount("POST /eventos/contenido", a.Contenido)
	mount("POST /internal/outbox/flush", a.OutboxFlush)
	mux.Handle("/", api)
	return events.CorrelationMiddleware(mux)
}
