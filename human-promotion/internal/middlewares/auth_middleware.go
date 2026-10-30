package middlewares

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"buenvivir/human-promotion/internal/audit"
	"buenvivir/human-promotion/internal/events"
	"buenvivir/human-promotion/internal/repositories"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	RoleIDKey contextKey = "roleID"
)

// JWTService interface local para evitar ciclo de importación
type JWTService interface {
	ValidateToken(tokenString string) (jwt.MapClaims, error)
}

// AuthMiddleware verifica el token JWT (cookie "__Host-access_token" o
// header Authorization) y deja el usuario y su rol en el contexto, con
// el mismo criterio que auth-service y education-service.
func AuthMiddleware(jwtService JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

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

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, RoleIDKey, roleID)
			// El token se guarda para reenviarlo a education-service al
			// pedir las habilidades del participante (ver SkillDirectory).
			ctx = repositories.ContextWithToken(ctx, tokenString)
			// Actor e IP de los eventos y del registro de auditoría.
			ctx = events.WithActor(ctx, userID)
			ctx = audit.WithClientIP(ctx, clientIP(r.RemoteAddr))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromContext devuelve el usuario autenticado que dejó AuthMiddleware.
func UserIDFromContext(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}

// clientIP saca el puerto de RemoteAddr (que middleware.RealIP ya
// reemplazó por la IP del cliente si vino por el gateway).
func clientIP(remoteAddr string) string {
	if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
		return host
	}
	return remoteAddr
}
