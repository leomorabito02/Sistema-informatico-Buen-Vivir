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

// AuthMiddleware verifica el token JWT (cookie "__Host-access_token" o
// header Authorization) y deja el usuario y su rol en el contexto, con
// el mismo criterio que human-promotion.
func AuthMiddleware(jwtService JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			var tokenString string
			if cookie, err := r.Cookie("__Host-access_token"); err == nil && cookie.Value != "" {
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
