package services

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService valida JWTs emitidos por auth-service. campaigns-service
// NUNCA genera tokens: solo los valida, con el mismo JWT_SECRET (HS256)
// que auth-service (mismo criterio que human-promotion y education).
type JWTService interface {
	ValidateToken(tokenString string) (jwt.MapClaims, error)
}

type jwtService struct {
	accessSecret []byte
}

// insecureDefaults son los secretos por defecto conocidos del código de
// auth-service: con cualquiera de ellos el servicio no arranca.
var insecureDefaults = map[string]bool{
	"":                         true,
	"fallback-access-secret":   true,
	"fallback-refresh-secret":  true,
	"super-secret-access-key":  true,
	"super-secret-refresh-key": true,
}

// NewJWTService lee JWT_SECRET del entorno y rechaza arrancar si está
// vacío o es un default inseguro.
func NewJWTService() JWTService {
	accessSecret := os.Getenv("JWT_SECRET")
	if insecureDefaults[accessSecret] {
		fmt.Fprintln(os.Stderr, "FATAL: JWT_SECRET is not set or uses an insecure default. Set a strong secret in environment variables.")
		os.Exit(1)
	}
	return NewJWTServiceWithSecret(accessSecret)
}

// NewJWTServiceWithSecret arma el validador con un secreto dado (tests).
func NewJWTServiceWithSecret(secret string) JWTService {
	return &jwtService{accessSecret: []byte(secret)}
}

func (s *jwtService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.accessSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims["type"] != "access" {
		return nil, errors.New("invalid token type")
	}
	return claims, nil
}
