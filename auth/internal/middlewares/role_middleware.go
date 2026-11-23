package middlewares

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
				http.Error(w, "Forbidden: insufficient role permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRoleOrSelf deja pasar a los roles indicados (más admin y dev) y a cualquier
// usuario autenticado cuando el parámetro `param` de la URL es su propio user_id.
func RequireRoleOrSelf(param string, allowedRoles ...string) func(http.Handler) http.Handler {
	byRole := RequireRole(allowedRoles...)
	return func(next http.Handler) http.Handler {
		roleChecked := byRole(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, _ := r.Context().Value(UserIDKey).(string)
			if userID != "" && chi.URLParam(r, param) == userID {
				next.ServeHTTP(w, r)
				return
			}
			roleChecked.ServeHTTP(w, r)
		})
	}
}
