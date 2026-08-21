package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"auth/internal/middlewares"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) ValidateToken(token string) (jwt.MapClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(jwt.MapClaims), args.Error(1)
}

func (m *MockJWTService) GenerateTokenPair(userID string, role string, isExternal bool) (string, string, error) {
	return "", "", nil
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	mockJWT := new(MockJWTService)
	middleware := middlewares.AuthMiddleware(mockJWT)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_EmptyCookie(t *testing.T) {
	mockJWT := new(MockJWTService)
	middleware := middlewares.AuthMiddleware(mockJWT)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	// Add an empty cookie, should fallback to Authorization header which is also empty
	req.AddCookie(&http.Cookie{Name: "access_token", Value: ""})
	
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	mockJWT := new(MockJWTService)
	middleware := middlewares.AuthMiddleware(mockJWT)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	mockJWT := new(MockJWTService)
	middleware := middlewares.AuthMiddleware(mockJWT)

	claims := jwt.MapClaims{"sub": "usr-1", "role": "role_admin", "is_external": false}
	mockJWT.On("ValidateToken", "valid.token.here").Return(claims, nil)

	var contextUserID, contextRoleID string
	var contextIsExternal bool

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextUserID, _ = r.Context().Value(middlewares.UserIDKey).(string)
		contextRoleID, _ = r.Context().Value(middlewares.RoleIDKey).(string)
		contextIsExternal, _ = r.Context().Value(middlewares.IsExternalKey).(bool)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid.token.here")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "usr-1", contextUserID)
	assert.Equal(t, "role_admin", contextRoleID)
	assert.False(t, contextIsExternal)
}
