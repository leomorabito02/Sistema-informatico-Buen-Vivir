package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"auth/internal/controllers"
	"auth/internal/dto"
	"auth/internal/middlewares"
	"auth/internal/models"
	"auth/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTwofaService struct {
	mock.Mock
}

func (m *MockTwofaService) InitiateSetup(ctx context.Context, userID, method string) (*dto.SetupTwofaResponse, error) {
	args := m.Called(ctx, userID, method)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SetupTwofaResponse), args.Error(1)
}

func (m *MockTwofaService) ConfirmEnable(ctx context.Context, userID, method, code string) error {
	args := m.Called(ctx, userID, method, code)
	return args.Error(0)
}

func (m *MockTwofaService) Disable(ctx context.Context, userID, password, code string) error {
	args := m.Called(ctx, userID, password, code)
	return args.Error(0)
}

func (m *MockTwofaService) InitiateLoginChallenge(ctx context.Context, email string, userIDs []string, method string) (string, error) {
	args := m.Called(ctx, email, userIDs, method)
	return args.String(0), args.Error(1)
}

func (m *MockTwofaService) VerifyLoginChallenge(ctx context.Context, twofaNonce, code string) (*dto.CheckCredentialsResponse, *models.User, error) {
	args := m.Called(ctx, twofaNonce, code)
	var resp *dto.CheckCredentialsResponse
	if args.Get(0) != nil {
		resp = args.Get(0).(*dto.CheckCredentialsResponse)
	}
	var user *models.User
	if args.Get(1) != nil {
		user = args.Get(1).(*models.User)
	}
	return resp, user, args.Error(2)
}

func (m *MockTwofaService) ResendLoginOTP(ctx context.Context, twofaNonce string) error {
	args := m.Called(ctx, twofaNonce)
	return args.Error(0)
}

func (m *MockTwofaService) CreateTrustedDeviceToken(user *models.User, userAgent string) (string, time.Time, error) {
	args := m.Called(user, userAgent)
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockTwofaService) ValidateTrustedDeviceToken(token string, user *models.User, userAgent string) bool {
	args := m.Called(token, user, userAgent)
	return args.Bool(0)
}

func (m *MockTwofaService) GetActiveCodeForEmail(email string) string {
	args := m.Called(email)
	return args.String(0)
}

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.LoginResponse, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) != nil {
		return args.Get(0).(*dto.LoginResponse), args.Error(1)
	}
	return nil, args.Error(1)
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

func (m *MockAuthService) GetUserProfile(ctx context.Context, userID string) (map[string]interface{}, error) {
	args := m.Called(ctx, userID)
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

func TestAuthController_GetMe_DiceSiElDocenteEsExterno(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)
	ctx := context.WithValue(context.Background(), middlewares.UserIDKey, "usr")
	ctx = context.WithValue(ctx, middlewares.RoleIDKey, "role_teacher")
	ctx = context.WithValue(ctx, middlewares.IsExternalKey, true)
	req := httptest.NewRequest("GET", "/me", nil).WithContext(ctx)
	w := httptest.NewRecorder()

	mockSvc.On("GetApprovalStatus", mock.Anything, "role_teacher", "usr").Return("approved")
	ctrl.GetMe(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_external":true`)
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

type MockVerificationService struct {
	mock.Mock
}

func (m *MockVerificationService) GenerateAndSendVerification(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockVerificationService) VerifyEmail(ctx context.Context, rawToken string) error {
	args := m.Called(ctx, rawToken)
	return args.Error(0)
}

func (m *MockVerificationService) ResendVerification(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockVerificationService) IssueActionLink(ctx context.Context, requestID, userID string) (string, error) {
	args := m.Called(ctx, requestID, userID)
	return args.String(0), args.Error(1)
}

func TestAuthController_VerifyEmail_Success(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockVerifSvc := new(MockVerificationService)
	ctrl := controllers.NewAuthController(mockAuthSvc, mockVerifSvc)

	mockVerifSvc.On("VerifyEmail", mock.Anything, "valid-token-xyz").Return(nil)

	reqBody, _ := json.Marshal(dto.VerifyEmailRequest{Token: "valid-token-xyz"})
	req := httptest.NewRequest("POST", "/verify-email", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.VerifyEmail(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "exitosamente")
}

func TestAuthController_VerifyEmail_PatchSuccess(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockVerifSvc := new(MockVerificationService)
	ctrl := controllers.NewAuthController(mockAuthSvc, mockVerifSvc)

	mockVerifSvc.On("VerifyEmail", mock.Anything, "patch-token-123").Return(nil)

	reqBody, _ := json.Marshal(dto.VerifyEmailRequest{Token: "patch-token-123"})
	req := httptest.NewRequest("PATCH", "/verify-email", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.VerifyEmail(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "exitosamente")
}

func TestAuthController_VerifyEmail_InvalidToken(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockVerifSvc := new(MockVerificationService)
	ctrl := controllers.NewAuthController(mockAuthSvc, mockVerifSvc)

	mockVerifSvc.On("VerifyEmail", mock.Anything, "expired-token").Return(errors.New("el token de verificación ha expirado"))

	reqBody, _ := json.Marshal(dto.VerifyEmailRequest{Token: "expired-token"})
	req := httptest.NewRequest("POST", "/verify-email", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.VerifyEmail(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "expirado")
}

func TestAuthController_ResendVerification_Success(t *testing.T) {
	mockAuthSvc := new(MockAuthService)
	mockVerifSvc := new(MockVerificationService)
	ctrl := controllers.NewAuthController(mockAuthSvc, mockVerifSvc)

	mockVerifSvc.On("ResendVerification", mock.Anything, "user@example.com").Return(nil)

	reqBody, _ := json.Marshal(dto.ResendVerificationRequest{Email: "user@example.com"})
	req := httptest.NewRequest("POST", "/resend-verification", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.ResendVerification(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "reenviado")
}

func TestAuthController_Login_EmailNotVerified403(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)

	reqBody, _ := json.Marshal(dto.LoginRequest{UserID: "usr", RoleID: "role_participant", LoginNonce: "nonce"})
	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*dto.LoginRequest")).Return(nil, services.ErrEmailNotVerified)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.Login(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "EMAIL_NOT_VERIFIED")
}

func TestAuthController_Login_PendingApproval403(t *testing.T) {
	mockSvc := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockSvc)

	reqBody, _ := json.Marshal(dto.LoginRequest{UserID: "usr", RoleID: "role_participant", LoginNonce: "nonce"})
	mockSvc.On("Login", mock.Anything, mock.AnythingOfType("*dto.LoginRequest")).Return(nil, services.ErrAccountPendingApproval)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.Login(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ACCOUNT_PENDING_APPROVAL")
}

type MockPasswordRecoveryService struct {
	mock.Mock
}

func (m *MockPasswordRecoveryService) RequestReset(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

func (m *MockPasswordRecoveryService) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	args := m.Called(ctx, rawToken, newPassword)
	return args.Error(0)
}

func (m *MockPasswordRecoveryService) IssueActionLink(ctx context.Context, requestID, userID string) (string, error) {
	args := m.Called(ctx, requestID, userID)
	return args.String(0), args.Error(1)
}

func TestAuthController_ForgotPassword_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	mockRecoverySvc := new(MockPasswordRecoveryService)
	ctrl := controllers.NewAuthController(mockSvc, mockRecoverySvc)

	mockRecoverySvc.On("RequestReset", mock.Anything, "user@example.com").Return(nil)

	reqBody, _ := json.Marshal(dto.ForgotPasswordRequest{Email: "user@example.com"})
	req := httptest.NewRequest("POST", "/forgot-password", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.ForgotPassword(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "recibirás un enlace")
}

func TestAuthController_ForgotPassword_InvalidEmail(t *testing.T) {
	mockSvc := new(MockAuthService)
	mockRecoverySvc := new(MockPasswordRecoveryService)
	ctrl := controllers.NewAuthController(mockSvc, mockRecoverySvc)

	reqBody, _ := json.Marshal(dto.ForgotPasswordRequest{Email: "not-an-email"})
	req := httptest.NewRequest("POST", "/forgot-password", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.ForgotPassword(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthController_ResetPassword_Success(t *testing.T) {
	mockSvc := new(MockAuthService)
	mockRecoverySvc := new(MockPasswordRecoveryService)
	ctrl := controllers.NewAuthController(mockSvc, mockRecoverySvc)

	mockRecoverySvc.On("ResetPassword", mock.Anything, "token-123", "ValidPass123!").Return(nil)

	reqBody, _ := json.Marshal(dto.ResetPasswordRequest{Token: "token-123", NewPassword: "ValidPass123!"})
	req := httptest.NewRequest("POST", "/reset-password", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.ResetPassword(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "restablecida exitosamente")
}

func TestAuthController_ResetPassword_ServiceError(t *testing.T) {
	mockSvc := new(MockAuthService)
	mockRecoverySvc := new(MockPasswordRecoveryService)
	ctrl := controllers.NewAuthController(mockSvc, mockRecoverySvc)

	mockRecoverySvc.On("ResetPassword", mock.Anything, "expired-token", "ValidPass123!").Return(errors.New("el enlace de restablecimiento ha expirado"))

	reqBody, _ := json.Marshal(dto.ResetPasswordRequest{Token: "expired-token", NewPassword: "ValidPass123!"})
	req := httptest.NewRequest("POST", "/reset-password", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	ctrl.ResetPassword(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ha expirado")
}

func TestAuthController_VerifyTwofa(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockTwofa := new(MockTwofaService)
	ctrl := controllers.NewAuthController(mockAuth, mockTwofa)

	// Invalid body
	req := httptest.NewRequest("POST", "/verify-2fa", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()
	ctrl.VerifyTwofa(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Validation error (code not 6 digits)
	reqBody, _ := json.Marshal(dto.VerifyTwofaRequest{TwofaNonce: "nonce", Code: "12"})
	req = httptest.NewRequest("POST", "/verify-2fa", bytes.NewBuffer(reqBody))
	w = httptest.NewRecorder()
	ctrl.VerifyTwofa(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Twofa unavailable
	ctrlNoTwofa := controllers.NewAuthController(mockAuth)
	reqBody, _ = json.Marshal(dto.VerifyTwofaRequest{TwofaNonce: "nonce", Code: "123456"})
	req = httptest.NewRequest("POST", "/verify-2fa", bytes.NewBuffer(reqBody))
	w = httptest.NewRecorder()
	ctrlNoTwofa.VerifyTwofa(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// VerifyLoginChallenge unauthorized
	mockTwofa.On("VerifyLoginChallenge", mock.Anything, "bad-nonce", "123456").
		Return(nil, nil, errors.New("código incorrecto"))
	reqBody, _ = json.Marshal(dto.VerifyTwofaRequest{TwofaNonce: "bad-nonce", Code: "123456"})
	req = httptest.NewRequest("POST", "/verify-2fa", bytes.NewBuffer(reqBody))
	w = httptest.NewRecorder()
	ctrl.VerifyTwofa(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Success with TrustDevice
	validUser := &models.User{UserID: "u1", Email: "test@example.com"}
	expectedResp := &dto.CheckCredentialsResponse{LoginNonce: "loginnonce123"}
	mockTwofa.On("VerifyLoginChallenge", mock.Anything, "valid-nonce", "654321").
		Return(expectedResp, validUser, nil)
	mockTwofa.On("CreateTrustedDeviceToken", validUser, "TestUA").
		Return("trusted.token", time.Now().Add(24*time.Hour), nil)

	reqBody, _ = json.Marshal(dto.VerifyTwofaRequest{TwofaNonce: "valid-nonce", Code: "654321", TrustDevice: true})
	req = httptest.NewRequest("POST", "/verify-2fa", bytes.NewBuffer(reqBody))
	req.Header.Set("User-Agent", "TestUA")
	w = httptest.NewRecorder()
	ctrl.VerifyTwofa(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Set-Cookie"), "__Host-trusted_device=trusted.token")
}

func TestAuthController_ResendTwofa(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockTwofa := new(MockTwofaService)
	ctrl := controllers.NewAuthController(mockAuth, mockTwofa)

	// Invalid body
	req := httptest.NewRequest("POST", "/resend-2fa", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()
	ctrl.ResendTwofa(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Empty nonce
	reqBody, _ := json.Marshal(dto.ResendTwofaRequest{TwofaNonce: ""})
	req = httptest.NewRequest("POST", "/resend-2fa", bytes.NewBuffer(reqBody))
	w = httptest.NewRecorder()
	ctrl.ResendTwofa(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Twofa unavailable
	ctrlNoTwofa := controllers.NewAuthController(mockAuth)
	reqBody, _ = json.Marshal(dto.ResendTwofaRequest{TwofaNonce: "nonce1"})
	req = httptest.NewRequest("POST", "/resend-2fa", bytes.NewBuffer(reqBody))
	w = httptest.NewRecorder()
	ctrlNoTwofa.ResendTwofa(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Service error
	mockTwofa.On("ResendLoginOTP", mock.Anything, "bad-nonce").Return(errors.New("expiró"))
	reqBody, _ = json.Marshal(dto.ResendTwofaRequest{TwofaNonce: "bad-nonce"})
	req = httptest.NewRequest("POST", "/resend-2fa", bytes.NewBuffer(reqBody))
	w = httptest.NewRecorder()
	ctrl.ResendTwofa(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Success
	mockTwofa.On("ResendLoginOTP", mock.Anything, "good-nonce").Return(nil)
	reqBody, _ = json.Marshal(dto.ResendTwofaRequest{TwofaNonce: "good-nonce"})
	req = httptest.NewRequest("POST", "/resend-2fa", bytes.NewBuffer(reqBody))
	w = httptest.NewRecorder()
	ctrl.ResendTwofa(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthController_SetupTwofa(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockTwofa := new(MockTwofaService)
	ctrl := controllers.NewAuthController(mockAuth, mockTwofa)

	// Missing UserID in context
	req := httptest.NewRequest("POST", "/setup-2fa", nil)
	w := httptest.NewRecorder()
	ctrl.SetupTwofa(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Invalid body
	req = httptest.NewRequest("POST", "/setup-2fa", bytes.NewBufferString("{invalid"))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u1"))
	w = httptest.NewRecorder()
	ctrl.SetupTwofa(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Validation error
	reqBody, _ := json.Marshal(dto.SetupTwofaRequest{Method: "invalid_method"})
	req = httptest.NewRequest("POST", "/setup-2fa", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u1"))
	w = httptest.NewRecorder()
	ctrl.SetupTwofa(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Twofa unavailable
	ctrlNoTwofa := controllers.NewAuthController(mockAuth)
	reqBody, _ = json.Marshal(dto.SetupTwofaRequest{Method: "email"})
	req = httptest.NewRequest("POST", "/setup-2fa", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u1"))
	w = httptest.NewRecorder()
	ctrlNoTwofa.SetupTwofa(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Service error
	mockTwofa.On("InitiateSetup", mock.Anything, "u1", "email").Return(nil, errors.New("err"))
	req = httptest.NewRequest("POST", "/setup-2fa", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u1"))
	w = httptest.NewRecorder()
	ctrl.SetupTwofa(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Success
	expectedResp := &dto.SetupTwofaResponse{Method: "email", Message: "ok"}
	mockTwofa.On("InitiateSetup", mock.Anything, "u2", "email").Return(expectedResp, nil)
	req = httptest.NewRequest("POST", "/setup-2fa", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u2"))
	w = httptest.NewRecorder()
	ctrl.SetupTwofa(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthController_EnableTwofa(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockTwofa := new(MockTwofaService)
	ctrl := controllers.NewAuthController(mockAuth, mockTwofa)

	// Missing UserID
	req := httptest.NewRequest("POST", "/enable-2fa", nil)
	w := httptest.NewRecorder()
	ctrl.EnableTwofa(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Validation error
	reqBody, _ := json.Marshal(dto.EnableTwofaRequest{Method: "email", Code: "123"})
	req = httptest.NewRequest("POST", "/enable-2fa", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u1"))
	w = httptest.NewRecorder()
	ctrl.EnableTwofa(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Twofa unavailable
	ctrlNoTwofa := controllers.NewAuthController(mockAuth)
	reqBody, _ = json.Marshal(dto.EnableTwofaRequest{Method: "email", Code: "123456"})
	req = httptest.NewRequest("POST", "/enable-2fa", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u1"))
	w = httptest.NewRecorder()
	ctrlNoTwofa.EnableTwofa(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Service error
	mockTwofa.On("ConfirmEnable", mock.Anything, "u1", "email", "123456").Return(errors.New("err"))
	req = httptest.NewRequest("POST", "/enable-2fa", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u1"))
	w = httptest.NewRecorder()
	ctrl.EnableTwofa(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Success
	mockTwofa.On("ConfirmEnable", mock.Anything, "u2", "email", "123456").Return(nil)
	req = httptest.NewRequest("POST", "/enable-2fa", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u2"))
	w = httptest.NewRecorder()
	ctrl.EnableTwofa(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthController_DisableTwofa(t *testing.T) {
	mockAuth := new(MockAuthService)
	mockTwofa := new(MockTwofaService)
	ctrl := controllers.NewAuthController(mockAuth, mockTwofa)

	// Missing UserID
	req := httptest.NewRequest("POST", "/disable-2fa", nil)
	w := httptest.NewRecorder()
	ctrl.DisableTwofa(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Invalid body
	req = httptest.NewRequest("POST", "/disable-2fa", bytes.NewBufferString("{bad"))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u1"))
	w = httptest.NewRecorder()
	ctrl.DisableTwofa(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Twofa unavailable
	ctrlNoTwofa := controllers.NewAuthController(mockAuth)
	reqBody, _ := json.Marshal(dto.DisableTwofaRequest{Password: "pass"})
	req = httptest.NewRequest("POST", "/disable-2fa", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u1"))
	w = httptest.NewRecorder()
	ctrlNoTwofa.DisableTwofa(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Service error
	mockTwofa.On("Disable", mock.Anything, "u1", "wrong", "").Return(errors.New("err"))
	reqBody, _ = json.Marshal(dto.DisableTwofaRequest{Password: "wrong"})
	req = httptest.NewRequest("POST", "/disable-2fa", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u1"))
	w = httptest.NewRecorder()
	ctrl.DisableTwofa(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Success (clears cookie)
	mockTwofa.On("Disable", mock.Anything, "u2", "correct", "").Return(nil)
	reqBody, _ = json.Marshal(dto.DisableTwofaRequest{Password: "correct"})
	req = httptest.NewRequest("POST", "/disable-2fa", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), middlewares.UserIDKey, "u2"))
	w = httptest.NewRecorder()
	ctrl.DisableTwofa(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Set-Cookie"), "__Host-trusted_device=;")
}

func TestAuthController_RevokeTrustedDevice(t *testing.T) {
	mockAuth := new(MockAuthService)
	ctrl := controllers.NewAuthController(mockAuth)

	req := httptest.NewRequest("POST", "/revoke-trusted-device", nil)
	w := httptest.NewRecorder()
	ctrl.RevokeTrustedDevice(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Set-Cookie"), "__Host-trusted_device=;")
}
