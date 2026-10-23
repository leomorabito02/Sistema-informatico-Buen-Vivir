package middlewares

import (
	"log/slog"
	"net/http"
)

// RequireRole verifica que el usuario autenticado tenga uno de los roles
// permitidos. role_admin y role_dev pasan siempre (mismo bypass global
// que auth-service, education y human-promotion).
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(RoleIDKey).(string)
			if !ok {
				http.Error(w, "Unauthorized: missing role in context", http.StatusUnauthorized)
				return
			}

			allowed := userRole == "role_admin" || userRole == "role_dev"
			for _, role := range allowedRoles {
				if userRole == role {
					allowed = true
					break
				}
			}

			if !allowed {
				// HU-06: todo intento de acceso no autorizado queda registrado.
				slog.Warn("acceso denegado por rol",
					slog.String("user_id", UserIDFromContext(r.Context())),
					slog.String("role", userRole),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("ip", r.RemoteAddr),
				)
				http.Error(w, "Forbidden: insufficient role permissions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
