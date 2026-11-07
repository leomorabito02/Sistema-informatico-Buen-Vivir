package routes

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"buenvivir/labor-integration/internal/controllers"
)

type Controllers struct {
	JobOffer       *controllers.JobOfferController
	JobApplication *controllers.JobApplicationController
}

func RegisterRoutes(r chi.Router, ctrls Controllers) {
	r.Use(corsAndIdentityMiddleware)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"UP","service":"labor-integration"}`))
	})

	r.Route("/api/v1/labor-integration", func(r chi.Router) {
		r.Get("/offers", ctrls.JobOffer.List)
		r.Post("/offers", ctrls.JobOffer.Create)
		r.Put("/offers/{id}", ctrls.JobOffer.Update)
		r.Delete("/offers/{id}", ctrls.JobOffer.Delete)
		r.Put("/offers/{id}/approve", ctrls.JobOffer.Approve)
		r.Put("/offers/{id}/reject", ctrls.JobOffer.Reject)
		r.Get("/offers/{id}/applications", ctrls.JobApplication.ListByOffer)
		r.Get("/applications/me", ctrls.JobApplication.ListMine)
		r.Post("/applications", ctrls.JobApplication.Apply)
		r.Delete("/applications/{id}", ctrls.JobApplication.Delete)
		r.Put("/applications/{id}/status", ctrls.JobApplication.UpdateStatus)
	})
}

func corsAndIdentityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if applyCORS(w, r) {
			return
		}
		hydrateJWTHeaders(r)
		next.ServeHTTP(w, r)
	})
}

func applyCORS(w http.ResponseWriter, r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	allowedRaw := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS"))
	if origin != "" && allowedRaw != "" {
		for _, allowed := range strings.Split(allowedRaw, ",") {
			if strings.TrimSpace(allowed) == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Requested-With, X-User-ID, X-User-Role, X-Skip-Global-Error, X-Redirect-On-404, X-Skip-500-Redirect")
				w.Header().Set("Access-Control-Max-Age", "300")
				break
			}
		}
	}
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

func hydrateJWTHeaders(r *http.Request) {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return
	}
	tok := extractToken(r)
	if tok == "" {
		return
	}
	if sub, role, ok := parseAccessToken(tok, secret); ok {
		r.Header.Set("X-User-ID", sub)
		r.Header.Set("X-User-Role", role)
	}
}

func extractToken(r *http.Request) string {
	for _, cookieName := range []string{"__Host-access_token", "access_token"} {
		if c, err := r.Cookie(cookieName); err == nil && strings.TrimSpace(c.Value) != "" {
			return strings.TrimSpace(c.Value)
		}
	}
	if authHeader := r.Header.Get("Authorization"); strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}
	return ""
}

func parseAccessToken(token, secret string) (string, string, bool) {
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
