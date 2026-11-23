package middlewares_test

import (
	"context"
	"errors"
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

func TestAuthMiddleware_OptionsMethod(t *testing.T) {
	mockJWT := new(MockJWTService)
	middleware := middlewares.AuthMiddleware(mockJWT)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("OPTIONS", "/protected", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_HostCookie(t *testing.T) {
	mockJWT := new(MockJWTService)
	middleware := middlewares.AuthMiddleware(mockJWT)

	claims := jwt.MapClaims{"sub": "usr-host", "role": "role_dev", "is_external": true}
	mockJWT.On("ValidateToken", "host.token.here").Return(claims, nil)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "__Host-access_token", Value: "host.token.here"})
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthMiddleware_TokenValidationError(t *testing.T) {
	mockJWT := new(MockJWTService)
	middleware := middlewares.AuthMiddleware(mockJWT)

	mockJWT.On("ValidateToken", "bad.token").Return(nil, errors.New("signature expired"))

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer bad.token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_ServiceAuthorizationHeadersPrecedence(t *testing.T) {
	// Arrange
	mockJWT := new(MockJWTService)
	middleware := middlewares.AuthMiddleware(mockJWT)
	claims := jwt.MapClaims{
		"sub":  "notifications-service",
		"role": "role_dev",
		"type": "access",
	}
	mockJWT.On("ValidateToken", "internal.hs256.jwt").Return(claims, nil).Twice()

	var capturedRole string
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRole, _ = r.Context().Value(middlewares.RoleIDKey).(string)
		w.WriteHeader(http.StatusOK)
	}))

	// Act 1: X-Service-Authorization takes precedence over Cloud Run OIDC Authorization
	req1 := httptest.NewRequest("GET", "/api/v1/directory/internal/users/u1/contact", nil)
	req1.Header.Set("Authorization", "Bearer gcp.oidc.id.token")
	req1.Header.Set("X-Service-Authorization", "Bearer internal.hs256.jwt")
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, req1)

	// Assert 1
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, "role_dev", capturedRole)

	// Act 2: X-Forwarded-Authorization fallback
	capturedRole = ""
	req2 := httptest.NewRequest("GET", "/api/v1/directory/internal/users/u1/contact", nil)
	req2.Header.Set("Authorization", "Bearer gcp.oidc.id.token")
	req2.Header.Set("X-Forwarded-Authorization", "Bearer internal.hs256.jwt")
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, req2)

	// Assert 2
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "role_dev", capturedRole)
	mockJWT.AssertExpectations(t)
}

type MockPermissionChecker struct {
	mock.Mock
}

func (m *MockPermissionChecker) CheckPermission(ctx context.Context, roleID, action, resource string) (bool, error) {
	args := m.Called(ctx, roleID, action, resource)
	return args.Bool(0), args.Error(1)
}

func TestRequirePermissionMiddleware(t *testing.T) {
	mockPerm := new(MockPermissionChecker)
	middleware := middlewares.RequirePermissionMiddleware(mockPerm, "read", "users")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Missing UserID
	req := httptest.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Missing RoleID
	req = httptest.NewRequest("GET", "/users", nil)
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u1"))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Permission denied
	mockPerm.On("CheckPermission", mock.Anything, "role_volunteer", "read", "users").Return(false, nil).Once()
	req = httptest.NewRequest("GET", "/users", nil)
	ctx := context.WithValue(req.Context(), middlewares.UserIDKey, "u1")
	ctx = context.WithValue(ctx, middlewares.RoleIDKey, "role_volunteer")
	req = req.WithContext(ctx)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Permission granted
	mockPerm.On("CheckPermission", mock.Anything, "role_admin", "read", "users").Return(true, nil).Once()
	ctxAdmin := context.WithValue(req.Context(), middlewares.UserIDKey, "u1")
	ctxAdmin = context.WithValue(ctxAdmin, middlewares.RoleIDKey, "role_admin")
	req = req.WithContext(ctxAdmin)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
