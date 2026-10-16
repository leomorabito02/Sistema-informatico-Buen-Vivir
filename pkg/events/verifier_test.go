package events

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	testAudience = "https://education.run.app/eventos/usuarios"
	testSA       = "push-educacion@proyecto.iam.gserviceaccount.com"
)

type fakeGoogle struct {
	key      *rsa.PrivateKey
	server   *httptest.Server
	requests int
}

func newFakeGoogle(t *testing.T) *fakeGoogle {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	g := &fakeGoogle{key: key}
	g.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		g.requests++
		e := big.NewInt(int64(key.E)).Bytes()
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{
			{"kid": "k1", "kty": "RSA", "alg": "RS256", "n": base64.RawURLEncoding.EncodeToString(key.N.Bytes()), "e": base64.RawURLEncoding.EncodeToString(e)},
			{"kid": "ec", "kty": "EC"},
		}})
	}))
	t.Cleanup(g.server.Close)
	return g
}

func (g *fakeGoogle) token(t *testing.T, alg, kid string, claims map[string]any) string {
	t.Helper()
	h, _ := json.Marshal(map[string]string{"alg": alg, "kid": kid, "typ": "JWT"})
	c, _ := json.Marshal(claims)
	signing := base64.RawURLEncoding.EncodeToString(h) + "." + base64.RawURLEncoding.EncodeToString(c)
	digest := sha256.Sum256([]byte(signing))
	sig, err := rsa.SignPKCS1v15(rand.Reader, g.key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return signing + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func validClaims(now time.Time) map[string]any {
	return map[string]any{
		"iss": "https://accounts.google.com", "aud": testAudience,
		"iat": now.Unix(), "exp": now.Add(time.Hour).Unix(),
		"email": testSA, "email_verified": true,
	}
}

func (g *fakeGoogle) verifier(now time.Time) *OIDCVerifier {
	v := NewOIDCVerifier(testAudience, testSA)
	v.CertsURL = g.server.URL
	v.Now = func() time.Time { return now }
	return v
}

func requestWith(token string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	return r
}

func TestOIDCVerifierAceptaTokenValido(t *testing.T) {
	g := newFakeGoogle(t)
	now := time.Now()
	v := g.verifier(now)
	tok := g.token(t, "RS256", "k1", validClaims(now))
	if err := v.Verify(requestWith(tok)); err != nil {
		t.Fatalf("token válido rechazado: %v", err)
	}
	// La segunda vez usa la clave cacheada.
	if err := v.Verify(requestWith(tok)); err != nil || g.requests != 1 {
		t.Fatalf("debía usar la caché de claves: %v, pedidos %d", err, g.requests)
	}
}

func TestOIDCVerifierRechazos(t *testing.T) {
	g := newFakeGoogle(t)
	now := time.Now()
	mod := func(k string, val any) map[string]any {
		c := validClaims(now)
		c[k] = val
		return c
	}
	cases := map[string]string{
		"sin token":           "",
		"malformado":          "abc",
		"header ilegible":     "!!.e30.e30",
		"otro algoritmo":      g.token(t, "HS256", "k1", validClaims(now)),
		"kid desconocido":     g.token(t, "RS256", "k9", validClaims(now)),
		"otro issuer":         g.token(t, "RS256", "k1", mod("iss", "https://evil.example")),
		"otra audience":       g.token(t, "RS256", "k1", mod("aud", "https://otro")),
		"vencido":             g.token(t, "RS256", "k1", mod("exp", now.Add(-time.Hour).Unix())),
		"emitido a futuro":    g.token(t, "RS256", "k1", mod("iat", now.Add(time.Hour).Unix())),
		"otra cuenta":         g.token(t, "RS256", "k1", mod("email", "otra@x.iam.gserviceaccount.com")),
		"email no verificado": g.token(t, "RS256", "k1", mod("email_verified", false)),
	}
	tampered := g.token(t, "RS256", "k1", validClaims(now))
	cases["firma alterada"] = tampered[:len(tampered)-4] + "AAAA"
	cases["firma ilegible"] = tampered[:len(tampered)-4] + "!!!!"

	for name, tok := range cases {
		v := g.verifier(now)
		if err := v.Verify(requestWith(tok)); !errors.Is(err, ErrUnauthorized) {
			t.Errorf("%s: se esperaba ErrUnauthorized, llegó %v", name, err)
		}
	}
}

func TestOIDCVerifierClavesNoDisponibles(t *testing.T) {
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer down.Close()
	g := newFakeGoogle(t)
	now := time.Now()
	v := g.verifier(now)
	v.CertsURL = down.URL
	if err := v.Verify(requestWith(g.token(t, "RS256", "k1", validClaims(now)))); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("sin claves no se puede autorizar: %v", err)
	}
	v.CertsURL = "http://127.0.0.1:1"
	if err := v.Verify(requestWith(g.token(t, "RS256", "k1", validClaims(now)))); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("sin conexión no se puede autorizar: %v", err)
	}
}

func TestNewPushVerifier(t *testing.T) {
	if _, err := NewPushVerifier(PushAuthDisabled, "", nil, true); err == nil {
		t.Error("en producción no se puede desactivar")
	}
	if v, err := NewPushVerifier(PushAuthDisabled, "", nil, false); err != nil || v.Verify(nil) != nil {
		t.Error("desactivado en local debe aceptar todo")
	}
	if _, err := NewPushVerifier("", "aud", []string{" "}, false); err == nil {
		t.Error("sin service accounts no debe arrancar")
	}
	v, err := NewPushVerifier("", "aud", []string{" sa@x ", ""}, true)
	if err != nil {
		t.Fatal(err)
	}
	if o := v.(*OIDCVerifier); o.Audience != "aud" || len(o.AllowedEmails) != 1 || o.AllowedEmails[0] != "sa@x" {
		t.Errorf("verificador mal armado: %+v", o)
	}
}
