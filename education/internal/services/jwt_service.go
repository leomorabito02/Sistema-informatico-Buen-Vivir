package services

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService valida JWTs emitidos por auth-service. education-service
// NUNCA genera tokens (no tiene GenerateTokenPair ni JWT_REFRESH_SECRET):
// solo los valida, usando el mismo JWT_SECRET (HS256) que auth-service.
type JWTService interface {
	ValidateToken(tokenString string) (jwt.MapClaims, error)
}

type jwtService struct {
	accessSecret []byte
}

// NewJWTService lee JWT_SECRET del entorno. Si está vacío o es alguno
// de los defaults inseguros conocidos del código de auth-service,
// el servicio rechaza arrancar (mismo criterio que auth-service).
func NewJWTService() JWTService {
	accessSecret := os.Getenv("JWT_SECRET")

	insecureDefaults := map[string]bool{
		"":                        true,
		"fallback-access-secret":  true,
		"fallback-refresh-secret": true,
		"super-secret-access-key":  true,
		"super-secret-refresh-key": true,
	}

	if insecureDefaults[accessSecret] {
		fmt.Fprintln(os.Stderr, "FATAL: JWT_SECRET is not set or uses an insecure default. Set a strong secret in environment variables.")
		os.Exit(1)
	}

	return &jwtService{
		accessSecret: []byte(accessSecret),
	}
}

func (s *jwtService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	secret := s.accessSecret
	expectedType := "access"

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims["type"] != expectedType {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}
