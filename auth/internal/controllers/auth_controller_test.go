package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"auth/internal/controllers"
	"auth/internal/dto"
	"auth/internal/middlewares"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginResponse), args.Error(1)
}

func (m *MockAuthService) CheckCredentials(ctx context.Context, req *dto.CheckCredentialsRequest) (*dto.CheckCredentialsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CheckCredentialsResponse), args.Error(1)
}

func (m *MockAuthService) CheckGoogleCredentials(ctx context.Context, req *dto.GoogleLoginRequest) (*dto.CheckCredentialsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.CheckCredentialsResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginResponse), args.Error(1)
}

func (m *MockAuthService) GetApprovalStatus(ctx context.Context, roleID, userID string) string {
	args := m.Called(ctx, roleID, userID)
	return args.String(0)
}

func (m *MockAuthService) UpdateMyProfile(ctx context.Context, roleID, userID string, data map[string]interface{}) error {
	args := m.Called(ctx, roleID, userID, data)
	return args.Error(0)
}

func (m *MockAuthService) GetMyProfile(ctx context.Context, roleID, userID string) (map[string]interface{}, error) {
	args := m.Called(ctx, roleID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockAuthService) UpdateMyPassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	args := m.Called(ctx, userID, oldPassword, newPassword)
	return args.Error(0)
}

func (m *MockAuthService) LinkGoogleAccount(ctx context.Context, userID string, req *dto.LinkGoogleRequest) error {
	args := m.Called(ctx, userID, req)
	return args.Error(0)
}

func (m *MockAuthService) UnlinkGoogleAccount(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestAuthController_Login_BadJSON(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)

	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("{invalid json"))
	w := httptest.NewRecorder()

	ctrl.Login(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_Login_ServiceError(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)

	reqBody, _ := json.Marshal(dto.LoginRequest{UserID: "usr", RoleID: "role_participant", LoginNonce: "nonce"})
	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*dto.LoginRequest")).Return(nil, errors.New("service error"))

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.Login(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "service error")
}

func TestAuthController_Login_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)

	resp := &dto.LoginResponse{Token: "acc", RefreshToken: "ref", UserStatus: "active"}
	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*dto.LoginRequest")).Return(resp, nil)

	reqBody, _ := json.Marshal(dto.LoginRequest{UserID: "usr", RoleID: "role_participant", LoginNonce: "nonce"})
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.Login(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Set-Cookie"), "access_token=acc")
}

func TestAuthController_Register_BadJSON(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)

	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString("{invalid json"))
	w := httptest.NewRecorder()

	ctrl.Register(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_Register_ValidationError(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)

	reqBody, _ := json.Marshal(dto.RegisterRequest{Email: "invalid"})
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.Register(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_Register_ServiceError(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)

	reqBody, _ := json.Marshal(dto.RegisterRequest{
		Email: "test@test.com", FirstName: "AA", LastName: "BB", Password: "Password123!", Phone: "1234567890", Role: "participant",
	})
	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*dto.RegisterRequest")).Return(nil, errors.New("service error"))

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.Register(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "service error")
}

func TestAuthController_Register_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)

	resp := &dto.LoginResponse{Token: "acc", RefreshToken: "ref"}
	mockSvc.On("Register", mock.Anything, mock.AnythingOfType("*dto.RegisterRequest")).Return(resp, nil)

	reqBody, _ := json.Marshal(dto.RegisterRequest{
		Email: "test@test.com", FirstName: "AA", LastName: "BB", Password: "Password123!", Phone: "1234567890", Role: "participant",
	})
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.Register(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAuthController_CheckCredentials_BadJSON(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)

	req := httptest.NewRequest("POST", "/check-credentials", bytes.NewBufferString("{invalid json"))
	w := httptest.NewRecorder()

	ctrl.CheckCredentials(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_CheckCredentials_ValidationError(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)

	reqBody, _ := json.Marshal(dto.CheckCredentialsRequest{Email: ""})
	req := httptest.NewRequest("POST", "/check-credentials", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CheckCredentials(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_CheckCredentials_ServiceError(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)

	reqBody, _ := json.Marshal(dto.CheckCredentialsRequest{Email: "test@test.com", Password: "Password123!"})
	mockSvc.On("CheckCredentials", mock.Anything, mock.AnythingOfType("*dto.CheckCredentialsRequest")).Return(nil, errors.New("service error"))

	req := httptest.NewRequest("POST", "/check-credentials", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CheckCredentials(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthController_CheckCredentials_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)

	resp := &dto.CheckCredentialsResponse{LoginNonce: "nonce123"}
	mockSvc.On("CheckCredentials", mock.Anything, mock.AnythingOfType("*dto.CheckCredentialsRequest")).Return(resp, nil)

	reqBody, _ := json.Marshal(dto.CheckCredentialsRequest{Email: "test@test.com", Password: "Password123!"})
	req := httptest.NewRequest("POST", "/check-credentials", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.CheckCredentials(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthController_CheckGoogleCredentials_BadJSON(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	req := httptest.NewRequest("POST", "/google-login", bytes.NewBufferString("{invalid"))
	w := httptest.NewRecorder()
	ctrl.CheckGoogleCredentials(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}



func TestAuthController_CheckGoogleCredentials_ServiceError(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	reqBody, _ := json.Marshal(dto.GoogleLoginRequest{IDToken: "token"})
	mockSvc.On("CheckGoogleCredentials", mock.Anything, mock.AnythingOfType("*dto.GoogleLoginRequest")).Return(nil, errors.New("err"))
	req := httptest.NewRequest("POST", "/google-login", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	ctrl.CheckGoogleCredentials(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthController_CheckGoogleCredentials_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	resp := &dto.CheckCredentialsResponse{LoginNonce: "nonce"}
	mockSvc.On("CheckGoogleCredentials", mock.Anything, mock.AnythingOfType("*dto.GoogleLoginRequest")).Return(resp, nil)
	reqBody, _ := json.Marshal(dto.GoogleLoginRequest{IDToken: "token"})
	req := httptest.NewRequest("POST", "/google-login", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()
	ctrl.CheckGoogleCredentials(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthController_Logout_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	req := httptest.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()
	ctrl.Logout(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	// Check if cookies are cleared (MaxAge=-1)
	assert.Contains(t, w.Header().Get("Set-Cookie"), "Max-Age=0")
}

func TestAuthController_GetMe_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	ctx = context.WithValue(ctx, middlewares.RoleIDKey, "role")
	req := httptest.NewRequest("GET", "/me", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.On("GetApprovalStatus", mock.Anything, "role", "usr").Return("approved")
	ctrl.GetMe(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}



func TestAuthController_UpdateMyProfile_BadJSON(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	ctx = context.WithValue(ctx, middlewares.RoleIDKey, "role")
	req := httptest.NewRequest("PUT", "/me", bytes.NewBufferString("{invalid")).WithContext(ctx)
	w := httptest.NewRecorder()
	ctrl.UpdateMyProfile(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_UpdateMyProfile_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	ctx = context.WithValue(ctx, middlewares.RoleIDKey, "role")
	req := httptest.NewRequest("PUT", "/me", bytes.NewBufferString(`{"name":"test"}`)).WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.On("UpdateMyProfile", mock.Anything, "role", "usr", mock.Anything).Return(nil)
	ctrl.UpdateMyProfile(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthController_UpdateMyProfile_ServiceError(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	ctx = context.WithValue(ctx, middlewares.RoleIDKey, "role")
	req := httptest.NewRequest("PUT", "/me", bytes.NewBufferString(`{"name":"test"}`)).WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.On("UpdateMyProfile", mock.Anything, "role", "usr", mock.Anything).Return(errors.New("err"))
	ctrl.UpdateMyProfile(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_GetMyProfileDetails_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	ctx = context.WithValue(ctx, middlewares.RoleIDKey, "role")
	req := httptest.NewRequest("GET", "/me/details", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.On("GetMyProfile", mock.Anything, "role", "usr").Return(map[string]interface{}{"id": "usr"}, nil)
	ctrl.GetMyProfileDetails(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthController_UpdateMyPassword_BadJSON(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	req := httptest.NewRequest("PUT", "/me/password", bytes.NewBufferString("{invalid")).WithContext(ctx)
	w := httptest.NewRecorder()
	ctrl.UpdateMyPassword(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_UpdateMyPassword_ValidationError(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	reqBody, _ := json.Marshal(map[string]string{})
	req := httptest.NewRequest("PUT", "/me/password", bytes.NewBuffer(reqBody)).WithContext(ctx)
	w := httptest.NewRecorder()
	ctrl.UpdateMyPassword(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_UpdateMyPassword_ServiceError(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	reqBody, _ := json.Marshal(map[string]string{"old_password": "old", "new_password": "Password123!"})
	req := httptest.NewRequest("PUT", "/me/password", bytes.NewBuffer(reqBody)).WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.On("UpdateMyPassword", mock.Anything, "usr", "old", "Password123!").Return(errors.New("err"))
	ctrl.UpdateMyPassword(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_UpdateMyPassword_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	reqBody, _ := json.Marshal(map[string]string{"old_password": "old", "new_password": "Password123!"})
	req := httptest.NewRequest("PUT", "/me/password", bytes.NewBuffer(reqBody)).WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.On("UpdateMyPassword", mock.Anything, "usr", "old", "Password123!").Return(nil)
	ctrl.UpdateMyPassword(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthController_HealthCheck_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	ctrl.HealthCheck(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthController_LinkGoogleAccount_BadJSON(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	req := httptest.NewRequest("POST", "/me/link-google", bytes.NewBufferString("{invalid")).WithContext(ctx)
	w := httptest.NewRecorder()
	ctrl.LinkGoogleAccount(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}


func TestAuthController_LinkGoogleAccount_ServiceError(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	reqBody, _ := json.Marshal(dto.LinkGoogleRequest{IDToken: "token"})
	req := httptest.NewRequest("POST", "/me/link-google", bytes.NewBuffer(reqBody)).WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.On("LinkGoogleAccount", mock.Anything, "usr", mock.AnythingOfType("*dto.LinkGoogleRequest")).Return(errors.New("err"))
	ctrl.LinkGoogleAccount(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_LinkGoogleAccount_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	reqBody, _ := json.Marshal(dto.LinkGoogleRequest{IDToken: "token"})
	req := httptest.NewRequest("POST", "/me/link-google", bytes.NewBuffer(reqBody)).WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.On("LinkGoogleAccount", mock.Anything, "usr", mock.AnythingOfType("*dto.LinkGoogleRequest")).Return(nil)
	ctrl.LinkGoogleAccount(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthController_UnlinkGoogleAccount_ServiceError(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	req := httptest.NewRequest("POST", "/me/unlink-google", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.On("UnlinkGoogleAccount", mock.Anything, "usr").Return(errors.New("err"))
	ctrl.UnlinkGoogleAccount(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_UnlinkGoogleAccount_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	req := httptest.NewRequest("POST", "/me/unlink-google", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.On("UnlinkGoogleAccount", mock.Anything, "usr").Return(nil)
	ctrl.UnlinkGoogleAccount(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
