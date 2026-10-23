package middlewares

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserIDKey contextKey = "userID"
	RoleIDKey contextKey = "roleID"
)

var ErrUnauthorized = errors.New("unauthorized: missing user context")

// AuthHeaders inspects identity headers injected by the API Gateway or Bearer token fallback.
func AuthHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := strings.TrimSpace(r.Header.Get("X-User-ID"))
		roleID := strings.TrimSpace(r.Header.Get("X-User-Role"))

		// Development / fallback: allow query param user_id only in test/direct calls if explicitly passed,
		// or allow Authorization / Cookie if X-User-ID is not yet set by an external proxy.
		if userID == "" {
			if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				// If a raw user id is passed in local tests as token
				if token != "" {
					userID = token
				}
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
