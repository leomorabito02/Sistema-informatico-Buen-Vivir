package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth/internal/models"
	"auth/internal/services"
	"auth/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

	// Edge cases for ValidateTrustedDeviceToken
	assert.False(t, twofaSvc.ValidateTrustedDeviceToken("", user, ua))
	assert.False(t, twofaSvc.ValidateTrustedDeviceToken(token, nil, ua))
	assert.False(t, twofaSvc.ValidateTrustedDeviceToken("invalidtokenwithoutdot", user, ua))
	assert.False(t, twofaSvc.ValidateTrustedDeviceToken("notbase64.signature", user, ua))
	assert.False(t, twofaSvc.ValidateTrustedDeviceToken(token+".extra", user, ua))
}

func TestTwofaService_InitiateSetup_EdgeCases(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	producer := &mockUsuariosProducer{}
	twofaSvc := services.NewTwofaService(mockUserRepo, nil, nil, producer)

	// Invalid method
	_, err := twofaSvc.InitiateSetup(context.Background(), "u1", "sms")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "método de autenticación inválido")

	// User not found
	mockUserRepo.On("FindByID", mock.Anything, "missing").Return(nil, errors.New("not found"))
	_, err = twofaSvc.InitiateSetup(context.Background(), "missing", "totp")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usuario no encontrado")

	// Valid email setup
	verifiedUser := &models.User{
		UserID:          "u-verified",
		Email:           "verified@example.com",
		IsEmailVerified: true,
	}
	mockUserRepo.On("FindByID", mock.Anything, "u-verified").Return(verifiedUser, nil)
	resp, err := twofaSvc.InitiateSetup(context.Background(), "u-verified", "email")
	assert.NoError(t, err)
	assert.Equal(t, "email", resp.Method)
	assert.Contains(t, resp.Message, "Se envió un código")
}

func TestTwofaService_ConfirmEnable_EmailAndEdgeCases(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	twofaSvc := services.NewTwofaService(mockUserRepo, nil, nil, nil)

	// Invalid code length
	err := twofaSvc.ConfirmEnable(context.Background(), "u1", "totp", "123")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "debe tener 6 dígitos")

	// User not found
	mockUserRepo.On("FindByID", mock.Anything, "u-missing").Return(nil, errors.New("not found"))
	err = twofaSvc.ConfirmEnable(context.Background(), "u-missing", "totp", "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usuario no encontrado")

	// AllUsers empty
	user := &models.User{UserID: "u1", Email: "empty@example.com"}
	mockUserRepo.On("FindByID", mock.Anything, "u1").Return(user, nil)
	mockUserRepo.On("FindAllByEmail", mock.Anything, "empty@example.com").Return([]models.User{}, nil)
	err = twofaSvc.ConfirmEnable(context.Background(), "u1", "totp", "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usuario no encontrado")

	// Expired / missing challenge
	mockUserRepo.On("FindAllByEmail", mock.Anything, "no-ch@example.com").Return([]models.User{{UserID: "u-noch", Email: "no-ch@example.com"}}, nil)
	mockUserRepo.On("FindByID", mock.Anything, "u-noch").Return(&models.User{UserID: "u-noch", Email: "no-ch@example.com"}, nil)
	err = twofaSvc.ConfirmEnable(context.Background(), "u-noch", "totp", "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expiró")

	// Email setup flow
	emailUser := &models.User{UserID: "u-email", Email: "email_setup@example.com", IsEmailVerified: true}
	mockUserRepo.On("FindByID", mock.Anything, "u-email").Return(emailUser, nil)
	mockUserRepo.On("FindAllByEmail", mock.Anything, "email_setup@example.com").Return([]models.User{*emailUser}, nil)
	mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	_, err = twofaSvc.InitiateSetup(context.Background(), "u-email", "email")
	require.NoError(t, err)
	code := utils.GetTwofaStore().GetCodeByEmail("email_setup@example.com")
	require.NotEmpty(t, code)

	// Wrong email code
	err = twofaSvc.ConfirmEnable(context.Background(), "u-email", "email", "000000")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incorrecto")

	// Correct email code
	err = twofaSvc.ConfirmEnable(context.Background(), "u-email", "email", code)
	assert.NoError(t, err)
}

func TestTwofaService_Disable_EdgeCases(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	twofaSvc := services.NewTwofaService(mockUserRepo, nil, nil, nil)

	// User not found
	mockUserRepo.On("FindByID", mock.Anything, "u-none").Return(nil, errors.New("err"))
	err := twofaSvc.Disable(context.Background(), "u-none", "pass", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usuario no encontrado")

	// 2FA not enabled
	mockUserRepo.On("FindByID", mock.Anything, "u-disabled").Return(&models.User{UserID: "u-disabled", TwofaEnabled: false}, nil)
	err = twofaSvc.Disable(context.Background(), "u-disabled", "pass", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no está activada")

	// AllUsers not found
	mockUserRepo.On("FindByID", mock.Anything, "u-noall").Return(&models.User{UserID: "u-noall", Email: "noall@example.com", TwofaEnabled: true}, nil)
	mockUserRepo.On("FindAllByEmail", mock.Anything, "noall@example.com").Return([]models.User{}, nil)
	err = twofaSvc.Disable(context.Background(), "u-noall", "pass", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "usuario no encontrado")

	// Unauthorized
	enabledUser := &models.User{
		UserID:       "u-authz",
		Email:        "authz@example.com",
		TwofaEnabled: true,
		PasswordHash: "$2a$10$somehash",
	}
	mockUserRepo.On("FindByID", mock.Anything, "u-authz").Return(enabledUser, nil)
	mockUserRepo.On("FindAllByEmail", mock.Anything, "authz@example.com").Return([]models.User{*enabledUser}, nil)
	err = twofaSvc.Disable(context.Background(), "u-authz", "wrongpass", "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "debés ingresar tu contraseña actual")

	// Disable with valid TOTP code
	secret, _ := utils.GenerateTOTPSecret()
	totpUser := &models.User{
		UserID:       "u-totp-dis",
		Email:        "totpdis@example.com",
		TwofaEnabled: true,
		TwofaSeed:    "totp",
		TwofaSecret:  secret,
	}
	mockUserRepo.On("FindByID", mock.Anything, "u-totp-dis").Return(totpUser, nil)
	mockUserRepo.On("FindAllByEmail", mock.Anything, "totpdis@example.com").Return([]models.User{*totpUser}, nil)
	mockUserRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil)

	validCode, _ := utils.GenerateTOTPCode(secret, time.Now())
	err = twofaSvc.Disable(context.Background(), "u-totp-dis", "", validCode)
	assert.NoError(t, err)
}

func TestTwofaService_LoginChallengeAndVerification(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockQuerier := new(MockRoleProfileQuerier)
	producer := &mockUsuariosProducer{}
	twofaSvc := services.NewTwofaService(mockUserRepo, mockQuerier, nil, producer)

	// InitiateLoginChallenge email
	nonce, err := twofaSvc.InitiateLoginChallenge(context.Background(), "challenge@example.com", []string{"u1"}, "email")
	require.NoError(t, err)
	require.NotEmpty(t, nonce)

	// GetActiveCodeForEmail
	activeCode := twofaSvc.GetActiveCodeForEmail("challenge@example.com")
	require.NotEmpty(t, activeCode)

	// ResendLoginOTP
	err = twofaSvc.ResendLoginOTP(context.Background(), nonce)
	assert.NoError(t, err)
	newCode := twofaSvc.GetActiveCodeForEmail("challenge@example.com")
	assert.NotEmpty(t, newCode)

	// VerifyLoginChallenge: invalid nonce
	_, _, err = twofaSvc.VerifyLoginChallenge(context.Background(), "bad-nonce", "123456")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expiró o es inválido")

	// Setup users for verification
	users := []models.User{
		{UserID: "u1", Email: "challenge@example.com", RoleID: "role_participant", Status: "active", IsEmailVerified: true},
		{UserID: "u2", Email: "challenge@example.com", RoleID: "role_teacher", Status: "active", IsEmailVerified: true},
		{UserID: "u3", Email: "challenge@example.com", RoleID: "role_admin", Status: "active", IsEmailVerified: true},
		{UserID: "u4", Email: "challenge@example.com", RoleID: "role_dev", Status: "active", IsEmailVerified: true},
		{UserID: "u5", Email: "challenge@example.com", RoleID: "role_company", Status: "active", IsEmailVerified: true},
		{UserID: "u6", Email: "challenge@example.com", RoleID: "role_volunteer", Status: "active", IsEmailVerified: true},
		{UserID: "u7", Email: "challenge@example.com", RoleID: "role_custom", Status: "active", IsEmailVerified: true},
	}
	mockUserRepo.On("FindAllByEmail", mock.Anything, "challenge@example.com").Return(users, nil)
	mockQuerier.On("GetApprovalStatus", mock.Anything, "participants", "u1").Return("approved", nil)
	mockQuerier.On("GetApprovalStatus", mock.Anything, "teachers", "u2").Return("pending", nil)
	mockQuerier.On("GetApprovalStatus", mock.Anything, "companies", "u5").Return("", errors.New("err"))
	mockQuerier.On("GetApprovalStatus", mock.Anything, "volunteers", "u6").Return("rejected", nil)

	// Verify with wrong code
	_, _, err = twofaSvc.VerifyLoginChallenge(context.Background(), nonce, "000000")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "código de seguridad incorrecto")

	// Verify with correct code
	resp, primaryUser, err := twofaSvc.VerifyLoginChallenge(context.Background(), nonce, newCode)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "u1", primaryUser.UserID)
	assert.Len(t, resp.Accounts, 7)

	// Verify TOTP challenge flow
	totpSecret, _ := utils.GenerateTOTPSecret()
	totpUsers := []models.User{
		{UserID: "u-totp", Email: "totp_login@example.com", RoleID: "role_admin", TwofaSecret: totpSecret},
	}
	mockUserRepo.On("FindAllByEmail", mock.Anything, "totp_login@example.com").Return(totpUsers, nil)

	totpNonce, err := twofaSvc.InitiateLoginChallenge(context.Background(), "totp_login@example.com", []string{"u-totp"}, "totp")
	require.NoError(t, err)

	// Resend on TOTP challenge fails
	err = twofaSvc.ResendLoginOTP(context.Background(), totpNonce)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "solo está disponible para autenticación por correo")

	// Valid TOTP code
	validTotp, _ := utils.GenerateTOTPCode(totpSecret, time.Now())
	resp, primaryUser, err = twofaSvc.VerifyLoginChallenge(context.Background(), totpNonce, validTotp)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "u-totp", primaryUser.UserID)
}
