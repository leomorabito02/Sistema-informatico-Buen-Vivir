package middlewares

import "net/http"

// RequireInternalTeacher bloquea a los docentes externos
// (claims["is_external"] == true) del acceso a las rutas /api/docente,
// que son exclusivas del docente responsable "interno" de un curso
// (ver historia de usuario del docente responsable). No existe en
// auth-service porque es una regla propia del dominio de Educación.
//
// Se aplica siempre después de AuthMiddleware (que puebla
// IsExternalKey en el contexto) y de RequireRole("role_teacher").
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
