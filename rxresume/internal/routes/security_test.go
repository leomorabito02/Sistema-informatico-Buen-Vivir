package routes

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func firmar(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()
	h := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	b, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	p := base64.RawURLEncoding.EncodeToString(b)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(h + "." + p))
	return h + "." + p + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// eco devuelve en headers la identidad que llegó al handler.
var eco = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Got-User", r.Header.Get("X-User-ID"))
	w.Header().Set("Got-Role", r.Header.Get("X-User-Role"))
})

func TestSecurity_IdentidadSaleDelJWT(t *testing.T) {
	const secret = "s3cr3t"
	h := Security(eco, secret, "")
	valido := firmar(t, secret, map[string]any{"sub": "usr-1", "role": "role_participant", "type": "access", "exp": time.Now().Add(time.Hour).Unix()})

	casos := []struct {
		nombre   string
		preparar func(r *http.Request)
		user     string
		role     string
	}{
		{"cookie válida", func(r *http.Request) {
			r.AddCookie(&http.Cookie{Name: "access_token", Value: valido})
		}, "usr-1", "role_participant"},
		{"Authorization original movido por el API Gateway", func(r *http.Request) {
			r.Header.Set("X-Forwarded-Authorization", "Bearer "+valido)
			r.Header.Set("Authorization", "Bearer token-de-cloud-run")
		}, "usr-1", "role_participant"},
		{"headers falsificados sin token se descartan", func(r *http.Request) {
			r.Header.Set("X-User-ID", "usr-otro")
			r.Header.Set("X-User-Role", "role_admin")
		}, "", ""},
		{"token con otra firma no sirve", func(r *http.Request) {
			r.AddCookie(&http.Cookie{Name: "access_token", Value: firmar(t, "otra", map[string]any{"sub": "usr-2", "type": "access"})})
			r.Header.Set("X-User-ID", "usr-otro")
		}, "", ""},
		{"refresh token no sirve", func(r *http.Request) {
			r.AddCookie(&http.Cookie{Name: "access_token", Value: firmar(t, secret, map[string]any{"sub": "usr-3", "type": "refresh"})})
		}, "", ""},
		{"token vencido no sirve", func(r *http.Request) {
			r.AddCookie(&http.Cookie{Name: "access_token", Value: firmar(t, secret, map[string]any{"sub": "usr-4", "type": "access", "exp": time.Now().Add(-time.Hour).Unix()})})
		}, "", ""},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/rxresume/me", nil)
			c.preparar(req)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)
			if got := w.Header().Get("Got-User"); got != c.user {
				t.Fatalf("X-User-ID = %q, quería %q", got, c.user)
			}
			if got := w.Header().Get("Got-Role"); got != c.role {
				t.Fatalf("X-User-Role = %q, quería %q", got, c.role)
			}
		})
	}
}

// Sin JWT_SECRET (Docker local) se respetan los headers del gateway nginx.
func TestSecurity_SinSecretoUsaHeadersDelGateway(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/rxresume/me", nil)
	req.Header.Set("X-User-ID", "usr-9")
	w := httptest.NewRecorder()
	Security(eco, "", "").ServeHTTP(w, req)
	if got := w.Header().Get("Got-User"); got != "usr-9" {
		t.Fatalf("X-User-ID = %q", got)
	}
}

func TestSecurity_CORS(t *testing.T) {
	h := Security(eco, "", "https://otro.example")

	pre := httptest.NewRequest(http.MethodOptions, "/api/v1/rxresume/me", nil)
	pre.Header.Set("Origin", "https://testing.ceibuenvivir.pages.dev")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, pre)
	if w.Code != http.StatusOK || w.Header().Get("Access-Control-Allow-Origin") != "https://testing.ceibuenvivir.pages.dev" {
		t.Fatalf("preflight: code=%d ACAO=%q", w.Code, w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Got-User") != "" {
		t.Fatal("el preflight no debe llegar a las rutas")
	}

	for origin, permitido := range map[string]bool{"https://otro.example": true, "https://malicioso.com": false} {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/rxresume/me", nil)
		req.Header.Set("Origin", origin)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
		if got := w.Header().Get("Access-Control-Allow-Origin") != ""; got != permitido {
			t.Fatalf("origen %s: CORS=%v, quería %v", origin, got, permitido)
		}
	}
}
