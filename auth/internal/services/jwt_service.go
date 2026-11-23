package services

import (
	"auth/internal/utils"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService handles generation and verification of JWTs
type JWTService interface {
	GenerateTokenPair(userID string, role string, isExternal bool) (string, string, error)
	ValidateToken(tokenString string) (jwt.MapClaims, error)
}

type jwtService struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewJWTServiceFromSecrets(access, refresh string) (JWTService, error) {
	if access == "" || refresh == "" {
		return nil, errors.New("jwt: access/refresh secret vacío")
	}
	if access == refresh {
		return nil, errors.New("jwt: access y refresh secret no pueden ser iguales")
	}
	if len(access) < 32 || len(refresh) < 32 {
		return nil, errors.New("jwt: secret con longitud insuficiente (min 32 bytes)")
	}

	insecureDefaults := map[string]bool{
		"fallback-access-secret":   true,
		"fallback-refresh-secret":  true,
		"super-secret-access-key":  true,
		"super-secret-refresh-key": true,
	}

	if insecureDefaults[access] || insecureDefaults[refresh] {
		return nil, errors.New("jwt: secret inseguro por defecto no permitido")
	}

	accessMin, _ := strconv.Atoi(os.Getenv("JWT_EXPIRATION_MINUTES"))
	if accessMin == 0 {
		accessMin = 15
	}

	refreshDays, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_EXPIRATION_DAYS"))
	if refreshDays == 0 {
		refreshDays = 7
	}

	return &jwtService{
		accessSecret:  []byte(access),
		refreshSecret: []byte(refresh),
		accessTTL:     time.Duration(accessMin) * time.Minute,
		refreshTTL:    time.Duration(refreshDays) * 24 * time.Hour,
	}, nil
}

func NewJWTServiceFromEnv() (JWTService, error) {
	access := os.Getenv("JWT_SECRET")
	refresh := os.Getenv("JWT_REFRESH_SECRET")
	return NewJWTServiceFromSecrets(access, refresh)
}

func (s *jwtService) GenerateTokenPair(userID string, role string, isExternal bool) (string, string, error) {
	accessTokenClaims := jwt.MapClaims{
		"sub":         userID,
		"role":        role,
		"is_external": isExternal,
		"exp":         time.Now().Add(s.accessTTL).Unix(),
		"iat":  time.Now().Unix(),
		"type": "access",
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessString, err := accessToken.SignedString(s.accessSecret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	refreshString, err := utils.GenerateSecureToken(32)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessString, refreshString, nil
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
