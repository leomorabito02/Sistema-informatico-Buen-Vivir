package middlewares

import (
	"context"
	"net/http"
)

// Middleware envuelve un http.Handler con comportamiento adicional
type Middleware func(http.Handler) http.Handler

// Chain aplica los middlewares en el orden en que se los pasa
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// AuthHeaders lee los headers inyectados por el API Gateway (NGINX)
// y los coloca en el contexto de la request para que la capa de
// controladores/servicios pueda identificar al usuario autenticado.
func AuthHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		roleID := r.Header.Get("X-User-Role")

		ctx := r.Context()
		if userID != "" {
			ctx = context.WithValue(ctx, UserIDKey, userID)
		}
		if roleID != "" {
			ctx = context.WithValue(ctx, RoleIDKey, roleID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
