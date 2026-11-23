package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"buenvivir/education/internal/audit"
	"buenvivir/education/internal/events"
)

type contextKey string

const (
	UserIDKey     contextKey = "userID"
	RoleIDKey     contextKey = "roleID"
	IsExternalKey contextKey = "isExternal"
	// TokenKey guarda el JWT crudo del usuario para reenviarlo a auth-service, que
	// aplica su propio RBAC.
	TokenKey contextKey = "rawToken"
)

// JWTService interface local para evitar ciclo de importación
type JWTService interface {
	ValidateToken(tokenString string) (jwt.MapClaims, error)
}

// AuthMiddleware verifica el token JWT en el header de la solicitud
func AuthMiddleware(jwtService JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			cookie, err := r.Cookie("__Host-access_token")
			if err == nil && cookie.Value != "" {
				tokenString = cookie.Value
			} else {
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
					return
				}
				parts := strings.Split(authHeader, " ")
				if len(parts) != 2 || parts[0] != "Bearer" {
					http.Error(w, "Invalid token format", http.StatusUnauthorized)
					return
				}
				tokenString = parts[1]
			}

			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			userID, _ := claims["sub"].(string)
			roleID, _ := claims["role"].(string)
			isExternal, _ := claims["is_external"].(bool)

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, RoleIDKey, roleID)
			ctx = context.WithValue(ctx, IsExternalKey, isExternal)
			ctx = context.WithValue(ctx, TokenKey, tokenString)
			// Actor de los eventos que se publiquen en este request.
			ctx = events.WithActor(ctx, userID)
			// IP del registro de auditoría (baja de cursos, asistencia general).
			ctx = audit.WithClientIP(ctx, r.RemoteAddr)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermissionMiddleware verifica si el usuario en el contexto tiene permisos
func RequirePermissionMiddleware(permissionService interface {
	CheckPermission(ctx context.Context, roleID, action, resource string) (bool, error)
}, requiredAction, requiredResource string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			userIDVal := r.Context().Value(UserIDKey)
			if userIDVal == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			roleIDVal := r.Context().Value(RoleIDKey)
			if roleIDVal == nil {
				http.Error(w, "Unauthorized: missing role", http.StatusUnauthorized)
				return
			}
			roleID := roleIDVal.(string)

			hasPermission, err := permissionService.CheckPermission(r.Context(), roleID, requiredAction, requiredResource)

			if err != nil || !hasPermission {
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
