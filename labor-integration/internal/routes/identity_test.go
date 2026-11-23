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

	"github.com/go-chi/chi/v5"
)

const secretoPrueba = "secreto-de-prueba-de-al-menos-32-caracteres"

func firmar(t *testing.T, claims map[string]any) string {
	t.Helper()
	enc := base64.RawURLEncoding
	h := enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	p, err := json.Marshal(claims)
	if err != nil {
		t.Fatal(err)
	}
	body := h + "." + enc.EncodeToString(p)
	mac := hmac.New(sha256.New, []byte(secretoPrueba))
	mac.Write([]byte(body))
	return body + "." + enc.EncodeToString(mac.Sum(nil))
}

// identidad corre el middleware y devuelve lo que verían los controllers.
func identidad(r *http.Request) (string, string) {
	var id, rol string
	h := corsAndIdentityMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		id, rol = r.Header.Get("X-User-ID"), r.Header.Get("X-User-Role")
	}))
	h.ServeHTTP(httptest.NewRecorder(), r)
	return id, rol
}

// Con JWT_SECRET, los headers de identidad que manda el cliente no valen; solo el JWT
// verificado.
func TestIdentidad_NoSeConfiaEnHeadersDelCliente(t *testing.T) {
	t.Setenv("JWT_SECRET", secretoPrueba)

	t.Run("sin token, X-User-Role falsificado se descarta", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, "/api/v1/labor-integration/offers/o1/approve", nil)
		r.Header.Set("X-User-ID", "usr-atacante")
		r.Header.Set("X-User-Role", "role_admin")
		id, rol := identidad(r)
		if id != "" || rol != "" {
			t.Fatalf("se aceptó una identidad falsificada: id=%q rol=%q", id, rol)
		}
	})

	t.Run("token inválido + headers falsificados se descartan", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/labor-integration/offers", nil)
		r.Header.Set("Authorization", "Bearer a.b.c")
		r.Header.Set("X-User-Role", "role_admin")
		if _, rol := identidad(r); rol != "" {
			t.Fatalf("rol aceptado con token inválido: %q", rol)
		}
	})

	t.Run("token válido de empresa: no se puede subir a admin por header", func(t *testing.T) {
		tok := firmar(t, map[string]any{"sub": "usr-empresa", "role": "role_company", "type": "access", "exp": time.Now().Add(time.Hour).Unix()})
		r := httptest.NewRequest(http.MethodGet, "/api/v1/labor-integration/offers", nil)
		r.Header.Set("Authorization", "Bearer "+tok)
		r.Header.Set("X-User-Role", "role_admin")
		id, rol := identidad(r)
		if id != "usr-empresa" || rol != "role_company" {
			t.Fatalf("identidad: id=%q rol=%q", id, rol)
		}
	})

	t.Run("token vencido o de refresh no da identidad", func(t *testing.T) {
		for _, c := range []map[string]any{
			{"sub": "u", "role": "role_admin", "type": "access", "exp": time.Now().Add(-time.Hour).Unix()},
			{"sub": "u", "role": "role_admin", "type": "refresh", "exp": time.Now().Add(time.Hour).Unix()},
		} {
			r := httptest.NewRequest(http.MethodGet, "/api/v1/labor-integration/offers", nil)
			r.Header.Set("Authorization", "Bearer "+firmar(t, c))
			if _, rol := identidad(r); rol != "" {
				t.Fatalf("token %v aceptado", c)
			}
		}
	})
}

func TestRegisterRoutes_AllowsPreRegisteredRoutesWithoutPanic(t *testing.T) {
	// Arrange
	r := chi.NewRouter()
	r.Handle("/metrics", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Act: RegisterRoutes must not panic even when /metrics was registered beforehand.
	RegisterRoutes(r, Controllers{})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on /health, got %d", w.Code)
	}
}
