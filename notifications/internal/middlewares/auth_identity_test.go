package middlewares_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"buenvivir/notifications/internal/middlewares"
)

func firmarJWT(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()
	h := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	b, _ := json.Marshal(claims)
	p := base64.RawURLEncoding.EncodeToString(b)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(h + "." + p))
	return h + "." + p + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Con JWT_SECRET la identidad sale solo del token: un X-User-ID inventado
// (que el API Gateway de GCP reenvía) ya no alcanza (auditoría 17/11/2026).
func TestAuthHeaders_ConSecretoIgnoraHeadersFalsificados(t *testing.T) {
	const secret = "s3cr3t"
	t.Setenv("JWT_SECRET", secret)
	var got string
	h := middlewares.AuthHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = middlewares.GetUserID(r.Context())
	}))
	valido := firmarJWT(t, secret, map[string]any{"sub": "usr-1", "role": "role_participant", "type": "access", "exp": time.Now().Add(time.Hour).Unix()})

	// Solo header falsificado: 401.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	req.Header.Set("X-User-ID", "usr-victima")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("header falsificado: esperaba 401, fue %d", rec.Code)
	}

	// Token válido + header falsificado: manda el token.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: valido})
	req.Header.Set("X-User-ID", "usr-victima")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || got != "usr-1" {
		t.Fatalf("cookie válida: code=%d user=%q", rec.Code, got)
	}

	// Authorization original que el API Gateway mueve a X-Forwarded-Authorization.
	got = ""
	req = httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	req.Header.Set("Authorization", "Bearer token-de-cloud-run")
	req.Header.Set("X-Forwarded-Authorization", "Bearer "+valido)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || got != "usr-1" {
		t.Fatalf("X-Forwarded-Authorization: code=%d user=%q", rec.Code, got)
	}

	// Llamada directa S2S en Cloud Run con OIDC en Authorization y JWT interno en X-Service-Authorization.
	got = ""
	req = httptest.NewRequest(http.MethodGet, "/api/v1/notifications/preferences/init", nil)
	req.Header.Set("Authorization", "Bearer gcp-oidc-id-token")
	req.Header.Set("X-Service-Authorization", "Bearer "+valido)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || got != "usr-1" {
		t.Fatalf("X-Service-Authorization: code=%d user=%q", rec.Code, got)
	}

	// Firma inválida: 401.
	req = httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: firmarJWT(t, "otra", map[string]any{"sub": "usr-2", "type": "access"})})
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("firma inválida: esperaba 401, fue %d", rec.Code)
	}
}
