package middlewares

import "net/http"

// RequireInternalTeacher bloquea a los docentes externos (is_external) en
// /api/docente. Va después de AuthMiddleware y RequireRole("role_teacher").
func RequireInternalTeacher(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isExternal, _ := r.Context().Value(IsExternalKey).(bool)
		if isExternal {
			http.Error(w, "Forbidden: docentes externos no tienen acceso a este recurso", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
