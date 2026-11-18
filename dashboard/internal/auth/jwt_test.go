package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const secreto = "secreto-de-prueba-de-al-menos-32-caracteres"

func firmar(t *testing.T, metodo jwt.SigningMethod, clave any, claims jwt.MapClaims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(metodo, claims).SignedString(clave)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func acceso(extra map[string]any) jwt.MapClaims {
	c := jwt.MapClaims{"sub": "u-1", "role": "role_admin", "type": "access", "exp": time.Now().Add(time.Hour).Unix()}
	for k, v := range extra {
		c[k] = v
	}
	return c
}

func TestNewValidator_RechazaSecretosInseguros(t *testing.T) {
	for _, s := range []string{"", "fallback-access-secret", "super-secret-access-key"} {
		if _, err := NewValidator(s); err == nil {
			t.Errorf("%q debía rechazarse", s)
		}
	}
}

func TestValidar(t *testing.T) {
	v, err := NewValidator(secreto)
	if err != nil {
		t.Fatal(err)
	}
	id, err := v.Validar(firmar(t, jwt.SigningMethodHS256, []byte(secreto), acceso(nil)))
	if err != nil || id.UserID != "u-1" || id.Rol != "role_admin" {
		t.Fatalf("token válido: %+v, %v", id, err)
	}
	invalidos := map[string]string{
		"otro secreto": firmar(t, jwt.SigningMethodHS256, []byte("otro-secreto-cualquiera-de-32-caracteres"), acceso(nil)),
		"refresh":      firmar(t, jwt.SigningMethodHS256, []byte(secreto), acceso(map[string]any{"type": "refresh"})),
		"vencido":      firmar(t, jwt.SigningMethodHS256, []byte(secreto), acceso(map[string]any{"exp": time.Now().Add(-time.Minute).Unix()})),
		"sin firma":    firmar(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, acceso(nil)),
		"basura":       "no-es-un-jwt",
	}
	for nombre, tok := range invalidos {
		if _, err := v.Validar(tok); err == nil {
			t.Errorf("%s: debía rechazarse", nombre)
		}
	}
}

func TestTokenDelPedido(t *testing.T) {
	casos := []struct {
		nombre string
		armar  func(*http.Request)
		token  string
	}{
		{"cookie __Host-", func(r *http.Request) { r.AddCookie(&http.Cookie{Name: "__Host-access_token", Value: "a"}) }, "a"},
		{"cookie access_token", func(r *http.Request) { r.AddCookie(&http.Cookie{Name: "access_token", Value: "b"}) }, "b"},
		{"bearer", func(r *http.Request) { r.Header.Set("Authorization", "Bearer c") }, "c"},
		{"cookie gana", func(r *http.Request) {
			r.AddCookie(&http.Cookie{Name: "__Host-access_token", Value: "a"})
			r.Header.Set("Authorization", "Bearer c")
		}, "a"},
		{"basic no sirve", func(r *http.Request) { r.Header.Set("Authorization", "Basic c") }, ""},
		{"nada", func(*http.Request) {}, ""},
	}
	for _, tc := range casos {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		tc.armar(r)
		if got := TokenDelPedido(r); got != tc.token {
			t.Errorf("%s: %q, esperaba %q", tc.nombre, got, tc.token)
		}
	}
}
