package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const secretoPrueba = "secreto-de-prueba-de-al-menos-32-caracteres"

func firmar(t *testing.T, metodo jwt.SigningMethod, clave any, claims jwt.MapClaims) string {
	t.Helper()
	s, err := jwt.NewWithClaims(metodo, claims).SignedString(clave)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func claimsAcceso() jwt.MapClaims {
	return jwt.MapClaims{"sub": "usuario-1", "role": "role_volunteer", "type": "access", "exp": time.Now().Add(time.Hour).Unix()}
}

func TestValidateToken(t *testing.T) {
	svc := NewJWTServiceWithSecret(secretoPrueba)

	if _, err := svc.ValidateToken(firmar(t, jwt.SigningMethodHS256, []byte(secretoPrueba), claimsAcceso())); err != nil {
		t.Errorf("token válido rechazado: %v", err)
	}

	refresh := claimsAcceso()
	refresh["type"] = "refresh"
	vencido := claimsAcceso()
	vencido["exp"] = time.Now().Add(-time.Minute).Unix()

	invalidos := map[string]string{
		"otro secreto":     firmar(t, jwt.SigningMethodHS256, []byte("otro-secreto-cualquiera-de-32-caracteres"), claimsAcceso()),
		"tipo refresh":     firmar(t, jwt.SigningMethodHS256, []byte(secretoPrueba), refresh),
		"vencido":          firmar(t, jwt.SigningMethodHS256, []byte(secretoPrueba), vencido),
		"sin firma (none)": firmar(t, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, claimsAcceso()),
		"basura":           "no-es-un-jwt",
	}
	for nombre, token := range invalidos {
		if _, err := svc.ValidateToken(token); err == nil {
			t.Errorf("%s: el token debía rechazarse", nombre)
		}
	}
}
