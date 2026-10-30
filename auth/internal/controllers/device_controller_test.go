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
	"auth/internal/middlewares"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDeviceRepository struct {
	mock.Mock
}

func (m *MockDeviceRepository) RegisterDevice(ctx context.Context, userID, token, userAgent string) error {
	args := m.Called(ctx, userID, token, userAgent)
	return args.Error(0)
}

func (m *MockDeviceRepository) GetUserTokens(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockDeviceRepository) DeleteDeviceTokens(ctx context.Context, userID string, tokens []string) error {
	args := m.Called(ctx, userID, tokens)
	return args.Error(0)
}

func TestDeviceController_RegisterDevice_Unauthorized(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	ctrl := controllers.NewDeviceController(mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/users/me/devices", bytes.NewBufferString(`{"token":"abc"}`))
	rec := httptest.NewRecorder()

	// Act
	ctrl.RegisterDevice(rec, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestDeviceController_RegisterDevice_MissingToken(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	ctrl := controllers.NewDeviceController(mockRepo)

	req := httptest.NewRequest(http.MethodPost, "/users/me/devices", bytes.NewBufferString(`{"token":""}`))
	ctx := context.WithValue(req.Context(), middlewares.UserIDKey, "usr-123")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Act
	ctrl.RegisterDevice(rec, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeviceController_RegisterDevice_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	mockRepo.On("RegisterDevice", mock.Anything, "usr-123", "fcm-token-val", "agent-x").Return(nil)

	ctrl := controllers.NewDeviceController(mockRepo)

	payload, _ := json.Marshal(map[string]string{
		"token":     "fcm-token-val",
		"userAgent": "agent-x",
	})
	req := httptest.NewRequest(http.MethodPost, "/users/me/devices", bytes.NewBuffer(payload))
	ctx := context.WithValue(req.Context(), middlewares.UserIDKey, "usr-123")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Act
	ctrl.RegisterDevice(rec, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rec.Code)
	mockRepo.AssertExpectations(t)
}

func TestDeviceController_RegisterDevice_Error(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	mockRepo.On("RegisterDevice", mock.Anything, "usr-123", "fcm-token-val", "agent-x").Return(errors.New("db error"))

	ctrl := controllers.NewDeviceController(mockRepo)

	payload, _ := json.Marshal(map[string]string{
		"token":     "fcm-token-val",
		"userAgent": "agent-x",
	})
	req := httptest.NewRequest(http.MethodPost, "/users/me/devices", bytes.NewBuffer(payload))
	ctx := context.WithValue(req.Context(), middlewares.UserIDKey, "usr-123")
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	// Act
	ctrl.RegisterDevice(rec, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	mockRepo.AssertExpectations(t)
}
