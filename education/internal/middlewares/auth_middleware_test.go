package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"buenvivir/education/internal/middlewares"
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

func TestAuthMiddleware_MissingToken(t *testing.T) {
	mockJWT := new(MockJWTService)
	middleware := middlewares.AuthMiddleware(mockJWT)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/cursos", nil)
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

	req := httptest.NewRequest("GET", "/api/cursos", nil)
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

	req := httptest.NewRequest("GET", "/api/cursos", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	mockJWT := new(MockJWTService)
	middleware := middlewares.AuthMiddleware(mockJWT)

	mockJWT.On("ValidateToken", "token-invalido").Return(nil, assert.AnError)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/cursos", nil)
	req.Header.Set("Authorization", "Bearer token-invalido")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ValidToken_PopulatesContext(t *testing.T) {
	mockJWT := new(MockJWTService)
	middleware := middlewares.AuthMiddleware(mockJWT)

	claims := jwt.MapClaims{"sub": "teacher-1", "role": "role_teacher", "is_external": true}
	mockJWT.On("ValidateToken", "valid.token.here").Return(claims, nil)

	var contextUserID, contextRoleID string
	var contextIsExternal bool

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextUserID, _ = r.Context().Value(middlewares.UserIDKey).(string)
		contextRoleID, _ = r.Context().Value(middlewares.RoleIDKey).(string)
		contextIsExternal, _ = r.Context().Value(middlewares.IsExternalKey).(bool)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/cursos", nil)
	req.Header.Set("Authorization", "Bearer valid.token.here")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "teacher-1", contextUserID)
	assert.Equal(t, "role_teacher", contextRoleID)
	assert.True(t, contextIsExternal)
}

func TestAuthMiddleware_ValidToken_FromCookie(t *testing.T) {
	mockJWT := new(MockJWTService)
	middleware := middlewares.AuthMiddleware(mockJWT)

	claims := jwt.MapClaims{"sub": "usr-1", "role": "role_participant", "is_external": false}
	mockJWT.On("ValidateToken", "cookie-token").Return(claims, nil)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/cursos", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
