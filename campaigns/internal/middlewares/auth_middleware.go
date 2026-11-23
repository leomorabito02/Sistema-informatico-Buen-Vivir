package middlewares

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	RoleIDKey contextKey = "roleID"
)

// JWTService interface local para evitar ciclo de importación.
type JWTService interface {
	ValidateToken(tokenString string) (jwt.MapClaims, error)
}

// extractBearerToken reads the JWT from cookie, X-Forwarded-Authorization (set by API Gateway), or Authorization.
func extractBearerToken(r *http.Request) (string, string) {
	if cookie, err := r.Cookie("__Host-access_token"); err == nil && cookie.Value != "" {
		return cookie.Value, ""
	}
	for _, header := range []string{"X-Forwarded-Authorization", "Authorization"} {
		authHeader := r.Header.Get(header)
		if authHeader == "" {
			continue
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return "", "Invalid token format"
		}
		return parts[1], ""
	}
	return "", "Unauthorized: missing token"
}

// AuthMiddleware verifica el token JWT (cookie "__Host-access_token",
// X-Forwarded-Authorization o Authorization) y deja el usuario y su rol en el
// contexto, con el mismo criterio que auth y notifications.
func AuthMiddleware(jwtService JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			tokenString, errMsg := extractBearerToken(r)
			if errMsg != "" {
				http.Error(w, errMsg, http.StatusUnauthorized)
				return
			}

			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				// HU-03: toda validación fallida queda registrada.
				slog.Warn("token inválido",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("ip", r.RemoteAddr),
				)
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			userID, _ := claims["sub"].(string)
			roleID, _ := claims["role"].(string)
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, RoleIDKey, roleID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext devuelve el usuario autenticado que dejó AuthMiddleware.
func UserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}

// RoleFromContext devuelve el rol del usuario autenticado.
func RoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(RoleIDKey).(string)
	return role
}
