package middlewares

import (
	"log/slog"
	"net/http"
)

// RequireRole verifica que el usuario autenticado (cuyo RoleID está en el contexto)
// tenga uno de los roles permitidos.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(allowedRoles) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			roleIDVal := r.Context().Value(RoleIDKey)
			if roleIDVal == nil {
				http.Error(w, "Unauthorized: missing role in context", http.StatusUnauthorized)
				return
			}

			userRole := roleIDVal.(string)

			allowed := false
			for _, role := range allowedRoles {
				if userRole == role {
					allowed = true
					break
				}
			}

			// Admin y dev tienen bypass global de roles
			if userRole == "role_admin" || userRole == "role_dev" {
				allowed = true
			}

			if !allowed {
				// Las fichas son datos sensibles (Ley 25.326): todo intento
				// denegado queda registrado. Cuando exista el log de
				// auditoría (HU-35), este es el punto donde enviarlo.
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
