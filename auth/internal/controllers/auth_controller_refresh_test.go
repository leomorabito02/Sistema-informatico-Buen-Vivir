package controllers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"auth/internal/controllers"
	"auth/internal/dto"
	"auth/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthController_Refresh_MissingCookie(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	w := httptest.NewRecorder()

	ctrl.Refresh(w, httptest.NewRequest("POST", "/refresh", nil))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockSvc.AssertNotCalled(t, "RefreshToken", mock.Anything, mock.Anything)
}

func TestAuthController_Refresh_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	req := httptest.NewRequest("POST", "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "__Host-refresh_token", Value: "old"})
	w := httptest.NewRecorder()

	mockSvc.On("RefreshToken", mock.Anything, "old").Return(&dto.LoginResponse{
		Token: "new-access", RefreshToken: "new-refresh", User: &models.User{UserID: "usr-1"},
	}, nil)

	ctrl.Refresh(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	cookies := strings.Join(w.Header().Values("Set-Cookie"), "\n")
	assert.Contains(t, cookies, "new-access")
	assert.Contains(t, cookies, "new-refresh")
}

func TestAuthController_Refresh_Invalid(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	req := httptest.NewRequest("POST", "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "__Host-refresh_token", Value: "old"})
	w := httptest.NewRecorder()

	mockSvc.On("RefreshToken", mock.Anything, "old").Return(nil, errors.New("refresh token expired or revoked"))

	ctrl.Refresh(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Header().Get("Set-Cookie"), "Max-Age=0", "limpia las cookies de sesión")
}

func withUserIDParam(req *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestAuthController_GetUserProfile(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	mockSvc.On("GetUserProfile", mock.Anything, "usr-1").Return(map[string]interface{}{"user": map[string]interface{}{"user_id": "usr-1"}}, nil)
	mockSvc.On("GetUserProfile", mock.Anything, "zzz").Return(nil, errors.New("usuario no encontrado"))

	w := httptest.NewRecorder()
	ctrl.GetUserProfile(w, withUserIDParam(httptest.NewRequest("GET", "/users/usr-1/profile", nil), "usr-1"))
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":"usr-1"`)

	w = httptest.NewRecorder()
	ctrl.GetUserProfile(w, withUserIDParam(httptest.NewRequest("GET", "/users/zzz/profile", nil), "zzz"))
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	ctrl.GetUserProfile(w, httptest.NewRequest("GET", "/users//profile", nil))
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
