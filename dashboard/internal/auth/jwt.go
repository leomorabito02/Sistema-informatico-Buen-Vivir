// Package auth valida el JWT de auth-service. En Docker el rol también
// llega por el API Gateway (X-User-Role), pero en Google Cloud el gateway
// no lo pone: por eso el Dashboard valida el token por su cuenta, igual
// que human-promotion y campaigns.
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Identidad es quién hizo el pedido, según el token.
type Identidad struct {
	UserID string
	Rol    string
}

// Validator valida tokens HS256 de tipo "access" con JWT_SECRET.
type Validator struct {
	secret []byte
}

// insecureDefaults son los secretos por defecto conocidos de auth-service.
var insecureDefaults = map[string]bool{
	"":                         true,
	"fallback-access-secret":   true,
	"fallback-refresh-secret":  true,
	"super-secret-access-key":  true,
	"super-secret-refresh-key": true,
}

// NewValidator devuelve error si el secreto está vacío o es inseguro.
func NewValidator(secret string) (*Validator, error) {
	if insecureDefaults[secret] {
		return nil, errors.New("JWT_SECRET vacío o con un valor inseguro conocido")
	}
	return &Validator{secret: []byte(secret)}, nil
}

// Validar revisa la firma, el vencimiento y el tipo del token.
func (v *Validator) Validar(token string) (Identidad, error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", t.Header["alg"])
		}
		return v.secret, nil
	})
	if err != nil {
		return Identidad{}, err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || !t.Valid || claims["type"] != "access" {
		return Identidad{}, errors.New("token inválido")
	}
	sub, _ := claims["sub"].(string)
	rol, _ := claims["role"].(string)
	return Identidad{UserID: sub, Rol: rol}, nil
}

// TokenDelPedido busca el token donde lo dejan auth-service y el
// frontend: cookie __Host-access_token (o access_token) o
// Authorization: Bearer.
func TokenDelPedido(r *http.Request) string {
	for _, nombre := range []string{"__Host-access_token", "access_token"} {
		if c, err := r.Cookie(nombre); err == nil && c.Value != "" {
			return c.Value
		}
	}
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}
