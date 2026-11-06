package services_test

import (
	"context"
	"testing"
	"time"

	"auth/internal/models"
	"auth/internal/services"
	"auth/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestTwofaService_InitiateSetup_TOTP(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	user := &models.User{
		UserID:          "usr-1",
		Email:           "test@example.com",
		IsEmailVerified: true,
	}
	mockUserRepo.On("FindByID", mock.Anything, "usr-1").Return(user, nil)
	twofaSvc := services.NewTwofaService(mockUserRepo, nil, nil, nil)

	// Act
	resp, err := twofaSvc.InitiateSetup(context.Background(), "usr-1", "totp")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "totp", resp.Method)
	assert.NotEmpty(t, resp.Secret)
	assert.Contains(t, resp.OtpauthURL, "otpauth://totp/BuenVivir:test@example.com")
}

func TestTwofaService_InitiateSetup_Email_NotVerified(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	user := &models.User{
		UserID:          "usr-2",
		Email:           "unverified@example.com",
		IsEmailVerified: false,
	}
	mockUserRepo.On("FindByID", mock.Anything, "usr-2").Return(user, nil)
	twofaSvc := services.NewTwofaService(mockUserRepo, nil, nil, nil)

	// Act
	resp, err := twofaSvc.InitiateSetup(context.Background(), "usr-2", "email")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "debes verificar tu correo")
}

func TestTwofaService_ConfirmEnable_TOTP_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	user := &models.User{
		UserID:          "usr-3",
		Email:           "totp@example.com",
		IsEmailVerified: true,
	}
	mockUserRepo.On("FindByID", mock.Anything, "usr-3").Return(user, nil)
	mockUserRepo.On("FindAllByEmail", mock.Anything, "totp@example.com").Return([]models.User{*user}, nil)
	mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	twofaSvc := services.NewTwofaService(mockUserRepo, nil, nil, nil)
	setupResp, err := twofaSvc.InitiateSetup(context.Background(), "usr-3", "totp")
	assert.NoError(t, err)

	// Generate valid TOTP code
	validCode, err := utils.GenerateTOTPCode(setupResp.Secret, time.Now())
	assert.NoError(t, err)

	// Act
	confirmErr := twofaSvc.ConfirmEnable(context.Background(), "usr-3", "totp", validCode)

	// Assert
	assert.NoError(t, confirmErr)
}

func TestTwofaService_ConfirmEnable_TOTP_InvalidCode(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	user := &models.User{
		UserID:          "usr-4",
		Email:           "invalid_totp@example.com",
		IsEmailVerified: true,
	}
	mockUserRepo.On("FindByID", mock.Anything, "usr-4").Return(user, nil)
	mockUserRepo.On("FindAllByEmail", mock.Anything, "invalid_totp@example.com").Return([]models.User{*user}, nil)

	twofaSvc := services.NewTwofaService(mockUserRepo, nil, nil, nil)
	_, err := twofaSvc.InitiateSetup(context.Background(), "usr-4", "totp")
	assert.NoError(t, err)

	// Act
	confirmErr := twofaSvc.ConfirmEnable(context.Background(), "usr-4", "totp", "000000")

	// Assert
	assert.Error(t, confirmErr)
	assert.Contains(t, confirmErr.Error(), "código de la aplicación autenticadora es incorrecto")
}

func TestTwofaService_Disable_WithPassword(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	hashed, _ := bcrypt.GenerateFromPassword([]byte("StrongPass123!"), bcrypt.DefaultCost)
	user := &models.User{
		UserID:       "usr-5",
		Email:        "disable@example.com",
		PasswordHash: string(hashed),
		TwofaEnabled: true,
		TwofaSeed:    "email",
	}
	mockUserRepo.On("FindByID", mock.Anything, "usr-5").Return(user, nil)
	mockUserRepo.On("FindAllByEmail", mock.Anything, "disable@example.com").Return([]models.User{*user}, nil)
	mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	twofaSvc := services.NewTwofaService(mockUserRepo, nil, nil, nil)

	// Act
	err := twofaSvc.Disable(context.Background(), "usr-5", "StrongPass123!", "")

	// Assert
	assert.NoError(t, err)
}

func TestTwofaService_TrustedDeviceToken(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	twofaSvc := services.NewTwofaService(mockUserRepo, nil, nil, nil)
	user := &models.User{
		UserID:       "usr-6",
		Email:        "trusted@example.com",
		PasswordHash: "$2a$10$xyz",
		TwofaSecret:  "JBSWY3DPEHPK3PXP",
		TwofaSeed:    "totp",
	}
	ua := "Mozilla/5.0 (X11; Linux x86_64)"

	// Act
	token, exp, err := twofaSvc.CreateTrustedDeviceToken(user, ua)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, exp.After(time.Now()))

	// Valid token matches
	isValid := twofaSvc.ValidateTrustedDeviceToken(token, user, ua)
	assert.True(t, isValid)

	// Tampered UA fails
	isInvalidUA := twofaSvc.ValidateTrustedDeviceToken(token, user, "Different UA")
	assert.False(t, isInvalidUA)

	// Altered credentials (changed password hash or 2FA secret) invalidates token
	alteredUser := &models.User{
		UserID:       "usr-6",
		Email:        "trusted@example.com",
		PasswordHash: "$2a$10$new_hash",
		TwofaSecret:  "JBSWY3DPEHPK3PXP",
		TwofaSeed:    "totp",
	}
	isInvalidStamp := twofaSvc.ValidateTrustedDeviceToken(token, alteredUser, ua)
	assert.False(t, isInvalidStamp)
}
