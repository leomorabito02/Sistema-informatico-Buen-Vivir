package middlewares

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	RoleIDKey contextKey = "roleID"
)

var ErrUnauthorized = errors.New("unauthorized: missing user context")

// AuthHeaders identifica a quien hace el pedido.
//
// Con JWT_SECRET (Cloud Run) la identidad sale SOLO del JWT verificado
// (cookie de sesión, X-Forwarded-Authorization o Authorization): los
// X-User-ID / X-User-Role que mande el cliente se ignoran, porque el API
// Gateway de GCP los reenvía tal cual y cualquiera podía leer las
// notificaciones de otro usuario (auditoría 17/11/2026). Sin JWT_SECRET
// (Docker local) se confía en los headers que pone el gateway nginx después
// de validar el token, como antes.
func AuthHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwtSecret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
		rawToken := extractRequestToken(r)

		var userID, roleID string
		if jwtSecret != "" {
			if rawToken != "" {
				if sub, role, ok := parseHS256AccessToken(rawToken, jwtSecret); ok {
					userID, roleID = sub, role
				}
			}
		} else {
			userID = strings.TrimSpace(r.Header.Get("X-User-ID"))
			roleID = strings.TrimSpace(r.Header.Get("X-User-Role"))
			// Tests y desarrollo sin gateway: token crudo como user id.
			if userID == "" && rawToken != "" {
				userID = rawToken
			}
		}

		if userID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "Unauthorized: missing user context",
			})
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, userID)
		if roleID != "" {
			ctx = context.WithValue(ctx, RoleIDKey, roleID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractRequestToken(r *http.Request) string {
	for _, cookieName := range []string{"__Host-access_token", "access_token"} {
		if c, err := r.Cookie(cookieName); err == nil && strings.TrimSpace(c.Value) != "" {
			return strings.TrimSpace(c.Value)
		}
	}
	// El API Gateway de GCP pone su token de Cloud Run en Authorization y
	// mueve el original a X-Forwarded-Authorization.
	for _, header := range []string{"X-Forwarded-Authorization", "Authorization"} {
		if v := r.Header.Get(header); strings.HasPrefix(v, "Bearer ") {
			return strings.TrimSpace(strings.TrimPrefix(v, "Bearer "))
		}
	}
	return ""
}

func parseHS256AccessToken(token, secret string) (string, string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", "", false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(sig, mac.Sum(nil)) {
		return "", "", false
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", false
	}
	var claims struct {
		Sub  string  `json:"sub"`
		Role string  `json:"role"`
		Type string  `json:"type"`
		Exp  float64 `json:"exp"`
	}
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return "", "", false
	}
	if claims.Type != "access" || claims.Sub == "" {
		return "", "", false
	}
	if claims.Exp > 0 && time.Now().Unix() > int64(claims.Exp) {
		return "", "", false
	}
	return claims.Sub, claims.Role, true
}

// GetUserID retrieves the authenticated user ID from context.
func GetUserID(ctx context.Context) (string, error) {
	val, ok := ctx.Value(UserIDKey).(string)
	if !ok || strings.TrimSpace(val) == "" {
		return "", ErrUnauthorized
	}
	return val, nil
}

// GetUserRole retrieves the authenticated user role from context if present.
func GetUserRole(ctx context.Context) string {
	val, ok := ctx.Value(RoleIDKey).(string)
	if !ok {
		return ""
	}
	return val
}
