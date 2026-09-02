package controllers

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService interface for validation, to avoid circular dependencies
type TokenValidatorJWTService interface {
	ValidateToken(tokenString string) (jwt.MapClaims, error)
}

type TokenValidatorController struct {
	jwtService TokenValidatorJWTService
}

func NewTokenValidatorController(jwtService TokenValidatorJWTService) *TokenValidatorController {
	return &TokenValidatorController{
		jwtService: jwtService,
	}
}

// ValidateToken is an internal endpoint called by NGINX auth_request
// It returns 200 OK and sets X-User-ID and X-User-Role headers if valid
// Otherwise, it returns 401 Unauthorized
func (c *TokenValidatorController) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var tokenString string

	// Try extracting from cookie first (__Host-access_token or access_token)
	cookie, err := r.Cookie("__Host-access_token")
	if err != nil || cookie.Value == "" {
		cookie, err = r.Cookie("access_token")
	}

	if err == nil && cookie.Value != "" {
		tokenString = cookie.Value
	} else {
		// Fallback to Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}
		tokenString = parts[1]
	}

	claims, err := c.jwtService.ValidateToken(tokenString)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	userID, _ := claims["sub"].(string)
	roleID, _ := claims["role"].(string)

	w.Header().Set("X-User-ID", userID)
	w.Header().Set("X-User-Role", roleID)
	w.WriteHeader(http.StatusOK)
}
