package clients

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func servidor(t *testing.T, h http.HandlerFunc) *Upstream {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return NewUpstream("prueba", srv.URL, 2*time.Second)
}

func TestGetJSON_ReenviaCredencialesYTokenDeGoogle(t *testing.T) {
	var visto http.Header
	u := servidor(t, func(w http.ResponseWriter, r *http.Request) {
		visto = r.Header.Clone()
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	u.IDToken = func(context.Context) (string, error) { return "token-google", nil }

	var out struct{ OK bool }
	cred := Credentials{Cookie: "__Host-access_token=x", Authorization: "Bearer usuario", UserID: "u1", UserRole: "role_admin"}
	if err := u.GetJSON(context.Background(), "/a", cred, &out); err != nil || !out.OK {
		t.Fatalf("%+v %v", out, err)
	}
	if visto.Get("X-Serverless-Authorization") != "Bearer token-google" {
		t.Errorf("falta el token de Google: %v", visto)
	}
	// La sesión del usuario viaja aparte, sin pisarse con el token de Google.
	if visto.Get("Authorization") != "Bearer usuario" || visto.Get("Cookie") != "__Host-access_token=x" ||
		visto.Get("X-User-ID") != "u1" || visto.Get("X-User-Role") != "role_admin" {
		t.Errorf("credenciales del usuario: %v", visto)
	}
}

func TestGetJSON_SinTokenDeGoogleNoMandaElHeader(t *testing.T) {
	u := servidor(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Serverless-Authorization") != "" {
			t.Error("no debía mandar X-Serverless-Authorization")
		}
		_, _ = w.Write([]byte(`[]`))
	})
	var out []any
	if err := u.GetJSON(context.Background(), "/", Credentials{}, &out); err != nil {
		t.Fatal(err)
	}
}

func TestGetJSON_Errores(t *testing.T) {
	var up *UpstreamError

	// Sin dirección: "no configurado" (por ejemplo, Inserción Laboral sin desplegar).
	err := (&Upstream{Name: "laboral"}).GetJSON(context.Background(), "/", Credentials{}, nil)
	if !errors.As(err, &up) || !errors.Is(err, ErrNoConfigurado) {
		t.Errorf("sin URL: %v", err)
	}

	// El servicio responde 403: se informa el código.
	u := servidor(t, func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusForbidden) })
	err = u.GetJSON(context.Background(), "/", Credentials{}, nil)
	if !errors.As(err, &up) || up.Status != http.StatusForbidden || up.Error() != "prueba respondió 403" {
		t.Errorf("403: %v", err)
	}

	// Respuesta que no es JSON.
	u = servidor(t, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("<html>")) })
	var out []any
	if err := u.GetJSON(context.Background(), "/", Credentials{}, &out); !errors.As(err, &up) || up.Status != http.StatusOK {
		t.Errorf("JSON inválido: %v", err)
	}

	// Falla el token de Google: no llega a llamar.
	u.IDToken = func(context.Context) (string, error) { return "", errors.New("sin metadata server") }
	if err := u.GetJSON(context.Background(), "/", Credentials{}, &out); !errors.As(err, &up) || up.Status != 0 {
		t.Errorf("token de Google: %v", err)
	}

	// Servicio caído.
	caido := NewUpstream("caido", "http://127.0.0.1:1", time.Second)
	if err := caido.GetJSON(context.Background(), "/", Credentials{}, &out); !errors.As(err, &up) || up.Status != 0 || up.Unwrap() == nil {
		t.Errorf("caído: %v", err)
	}
}

func TestFromRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Cookie", "c=1")
	r.Header.Set("Authorization", "Bearer t")
	r.Header.Set("X-User-ID", "u")
	r.Header.Set("X-User-Role", "role_admin")
	if c := FromRequest(r); c.Cookie != "c=1" || c.Authorization != "Bearer t" || c.UserID != "u" || c.UserRole != "role_admin" {
		t.Fatalf("%+v", c)
	}
}
