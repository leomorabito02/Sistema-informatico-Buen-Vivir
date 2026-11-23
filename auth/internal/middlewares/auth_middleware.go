package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string
const (
	UserIDKey     contextKey = "userID"
	RoleIDKey     contextKey = "roleID"
	IsExternalKey contextKey = "isExternal"
)

// JWTService interface local para evitar ciclo de importación
type JWTService interface {
	ValidateToken(tokenString string) (jwt.MapClaims, error)
}

// extractToken intenta obtener el JWT token primero desde cookies (__Host-access_token / access_token) y como fallback desde el encabezado Authorization.
func extractToken(r *http.Request) (string, int, string) {
	cookie, err := r.Cookie("__Host-access_token")
	if err != nil || cookie.Value == "" {
		cookie, err = r.Cookie("access_token")
	}

	if err == nil && cookie.Value != "" {
		return cookie.Value, 0, ""
	}

	for _, header := range []string{"X-Service-Authorization", "X-Forwarded-Authorization", "Authorization"} {
		authHeader := r.Header.Get(header)
		if authHeader == "" {
			continue
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return "", http.StatusUnauthorized, "Invalid token format"
		}
		return parts[1], 0, ""
	}

	return "", http.StatusUnauthorized, "Unauthorized: missing token"
}

// extractContextFromClaims llena el contexto HTTP con los claims decodificados del JWT.
func extractContextFromClaims(parentCtx context.Context, claims jwt.MapClaims) context.Context {
	userID, _ := claims["sub"].(string)
	roleID, _ := claims["role"].(string)
	isExternal, _ := claims["is_external"].(bool)

	ctx := context.WithValue(parentCtx, UserIDKey, userID)
	ctx = context.WithValue(ctx, RoleIDKey, roleID)
	return context.WithValue(ctx, IsExternalKey, isExternal)
}

// AuthMiddleware verifica el token JWT en el header o cookie de la solicitud
func AuthMiddleware(jwtService JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			tokenString, statusCode, errMsg := extractToken(r)
			if errMsg != "" {
				http.Error(w, errMsg, statusCode)
				return
			}

			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := extractContextFromClaims(r.Context(), claims)
			if h := auditActorFrom(r.Context()); h != nil {
				h.userID, _ = claims["sub"].(string)
				h.role, _ = claims["role"].(string)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermissionMiddleware verifica si el usuario en el contexto tiene permisos
func RequirePermissionMiddleware(permissionService interface{
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
