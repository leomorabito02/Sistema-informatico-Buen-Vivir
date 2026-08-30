package controllers_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth/internal/controllers"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTokenValidatorJWTService struct {
	mock.Mock
}

func (m *MockTokenValidatorJWTService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(jwt.MapClaims), args.Error(1)
}

func TestTokenValidatorController_ValidateToken_Success_Cookie(t *testing.T) {
	mockSvc := new(MockTokenValidatorJWTService)
	ctrl := controllers.NewTokenValidatorController(mockSvc)

	claims := jwt.MapClaims{"sub": "usr-1", "role": "admin"}
	mockSvc.On("ValidateToken", "valid_cookie_token").Return(claims, nil)

	req := httptest.NewRequest("GET", "/validate", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "valid_cookie_token"})
	w := httptest.NewRecorder()

	ctrl.ValidateToken(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "usr-1", w.Header().Get("X-User-ID"))
	assert.Equal(t, "admin", w.Header().Get("X-User-Role"))
}

func TestTokenValidatorController_ValidateToken_Success_Header(t *testing.T) {
	mockSvc := new(MockTokenValidatorJWTService)
	ctrl := controllers.NewTokenValidatorController(mockSvc)

	claims := jwt.MapClaims{"sub": "usr-1", "role": "admin"}
	mockSvc.On("ValidateToken", "valid_header_token").Return(claims, nil)

	req := httptest.NewRequest("GET", "/validate", nil)
	req.Header.Set("Authorization", "Bearer valid_header_token")
	w := httptest.NewRecorder()

	ctrl.ValidateToken(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "usr-1", w.Header().Get("X-User-ID"))
	assert.Equal(t, "admin", w.Header().Get("X-User-Role"))
}

func TestTokenValidatorController_ValidateToken_MissingToken(t *testing.T) {
	mockSvc := new(MockTokenValidatorJWTService)
	ctrl := controllers.NewTokenValidatorController(mockSvc)

	req := httptest.NewRequest("GET", "/validate", nil)
	w := httptest.NewRecorder()

	ctrl.ValidateToken(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTokenValidatorController_ValidateToken_InvalidFormat(t *testing.T) {
	mockSvc := new(MockTokenValidatorJWTService)
	ctrl := controllers.NewTokenValidatorController(mockSvc)

	req := httptest.NewRequest("GET", "/validate", nil)
	req.Header.Set("Authorization", "InvalidFormat token")
	w := httptest.NewRecorder()

	ctrl.ValidateToken(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTokenValidatorController_ValidateToken_InvalidToken(t *testing.T) {
	mockSvc := new(MockTokenValidatorJWTService)
	ctrl := controllers.NewTokenValidatorController(mockSvc)

	mockSvc.On("ValidateToken", "invalid_token").Return(nil, errors.New("invalid"))

	req := httptest.NewRequest("GET", "/validate", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	w := httptest.NewRecorder()

	ctrl.ValidateToken(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
