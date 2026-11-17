package routes

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Seguridad de borde del servicio (auditoría 17/11/2026).
//
// En Docker el gateway nginx valida el token y pone X-User-ID /
// X-User-Role. En GCP el API Gateway no hace eso: reenvía los headers que
// mande el cliente. Con JWT_SECRET configurado, la identidad sale SOLO del
// JWT verificado (cookie de sesión o Authorization) y se descartan los
// X-User-* recibidos, igual que en labor-integration.
//
// CORS: el API Gateway de GCP no agrega los headers, así que los pone el
// servicio (mismo criterio y lista fija que campaigns y auth). nginx en
// local oculta los del servicio y pone los suyos, así que no se duplican.

const (
	corsMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
	corsHeaders = "Accept, Authorization, Content-Type, X-CSRF-Token, X-Requested-With, X-Skip-Global-Error, X-Redirect-On-404, X-Skip-500-Redirect"
	corsMaxAge  = "300"
)

// Security envuelve el mux con CORS e identidad por JWT.
func Security(next http.Handler, jwtSecret, corsOrigins string) http.Handler {
	origins := parseOrigins(corsOrigins)
	secret := strings.TrimSpace(jwtSecret)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Vary", "Origin")
		if origin := r.Header.Get("Origin"); isAllowedOrigin(origin, origins) {
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Access-Control-Allow-Methods", corsMethods)
			h.Set("Access-Control-Allow-Headers", corsHeaders)
			h.Set("Access-Control-Max-Age", corsMaxAge)
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if secret != "" {
			r.Header.Del("X-User-ID")
			r.Header.Del("X-User-Role")
			if sub, role, ok := parseAccessToken(extractToken(r), secret); ok {
				r.Header.Set("X-User-ID", sub)
				r.Header.Set("X-User-Role", role)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func parseOrigins(raw string) []string {
	var out []string
	for _, o := range strings.Split(raw, ",") {
		if o = strings.TrimSpace(o); o != "" {
			out = append(out, o)
		}
	}
	return out
}

func isAllowedOrigin(origin string, envOrigins []string) bool {
	if origin == "" {
		return false
	}
	for _, o := range envOrigins {
		if o == origin {
			return true
		}
	}
	return origin == "http://localhost:4200" ||
		origin == "http://localhost:8080" ||
		origin == "https://ceibuenvivir.org" ||
		strings.HasSuffix(origin, ".ceibuenvivir.pages.dev") ||
		strings.HasSuffix(origin, ".ceibuenvivir.org")
}

// extractToken busca el JWT de sesión: cookie, después el Authorization
// original que el API Gateway mueve a X-Forwarded-Authorization (porque
// pone su propio token de Cloud Run en Authorization) y por último
// Authorization.
func extractToken(r *http.Request) string {
	for _, name := range []string{"__Host-access_token", "access_token"} {
		if c, err := r.Cookie(name); err == nil && strings.TrimSpace(c.Value) != "" {
			return strings.TrimSpace(c.Value)
		}
	}
	for _, header := range []string{"X-Forwarded-Authorization", "Authorization"} {
		if v := r.Header.Get(header); strings.HasPrefix(v, "Bearer ") {
			return strings.TrimSpace(strings.TrimPrefix(v, "Bearer "))
		}
	}
	return ""
}

// parseAccessToken valida un access token HS256 de auth y devuelve sub y rol.
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
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", false
	}
	var claims struct {
		Sub  string  `json:"sub"`
		Role string  `json:"role"`
		Type string  `json:"type"`
		Exp  float64 `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
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
