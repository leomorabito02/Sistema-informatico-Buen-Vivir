package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"auth/internal/dto"
	"auth/internal/events/producers"
	"auth/internal/models"
	"auth/internal/repositories"
	"auth/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

const (
	cookieTrustedDevice = "__Host-trusted_device"
	defaultTrustedDays  = 30
)

type TwofaService interface {
	InitiateSetup(ctx context.Context, userID, method string) (*dto.SetupTwofaResponse, error)
	ConfirmEnable(ctx context.Context, userID, method, code string) error
	Disable(ctx context.Context, userID, password, code string) error
	InitiateLoginChallenge(ctx context.Context, email string, userIDs []string, method string) (string, error)
	VerifyLoginChallenge(ctx context.Context, twofaNonce, code string) (*dto.CheckCredentialsResponse, *models.User, error)
	ResendLoginOTP(ctx context.Context, twofaNonce string) error
	CreateTrustedDeviceToken(user *models.User, userAgent string) (string, time.Time, error)
	ValidateTrustedDeviceToken(token string, user *models.User, userAgent string) bool
	GetActiveCodeForEmail(email string) string
}

type twofaService struct {
	userRepo   repositories.UserRepository
	querier    repositories.RoleProfileQuerier
	transactor repositories.Transactor
	producer   producers.UsuariosProducer
	store      *utils.TwofaStore
}

func NewTwofaService(
	userRepo repositories.UserRepository,
	querier repositories.RoleProfileQuerier,
	transactor repositories.Transactor,
	producer producers.UsuariosProducer,
) TwofaService {
	return &twofaService{
		userRepo:   userRepo,
		querier:    querier,
		transactor: transactor,
		producer:   producer,
		store:      utils.GetTwofaStore(),
	}
}

func (s *twofaService) InitiateSetup(ctx context.Context, userID, method string) (*dto.SetupTwofaResponse, error) {
	method = strings.ToLower(strings.TrimSpace(method))
	if method != "email" && method != "totp" {
		return nil, errors.New("método de autenticación inválido (debe ser 'email' o 'totp')")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("usuario no encontrado")
	}

	if method == "email" && !user.IsEmailVerified {
		return nil, errors.New("debes verificar tu correo electrónico antes de activar 2FA por email")
	}

	if method == "totp" {
		secret, err := utils.GenerateTOTPSecret()
		if err != nil {
			return nil, errors.New("error al generar secreto de autenticación")
		}

		setupNonce, err := s.store.CreateChallenge(user.Email, []string{user.UserID}, "totp_setup", "", secret, 10*time.Minute)
		if err != nil {
			return nil, errors.New("error al iniciar configuración")
		}

		issuer := "BuenVivir"
		otpauthURL := utils.BuildOTPAuthURI(issuer, user.Email, secret)

		return &dto.SetupTwofaResponse{
			Method:     "totp",
			Secret:     secret,
			OtpauthURL: otpauthURL,
			Message:    setupNonce,
		}, nil
	}

	// Method == "email"
	otpCode, err := utils.GenerateNumericOTP(6)
	if err != nil {
		return nil, errors.New("error al generar código de seguridad")
	}

	setupNonce, err := s.store.CreateChallenge(user.Email, []string{user.UserID}, "email_setup", otpCode, "", 10*time.Minute)
	if err != nil {
		return nil, errors.New("error al iniciar configuración")
	}

	if s.producer != nil {
		_ = s.producer.EmailSolicitado(ctx, setupNonce, user.UserID, "two_factor_code")
	}

	return &dto.SetupTwofaResponse{
		Method:  "email",
		Message: "Se envió un código de verificación de 6 dígitos a tu correo electrónico.",
	}, nil
}

func (s *twofaService) ConfirmEnable(ctx context.Context, userID, method, code string) error {
	method = strings.ToLower(strings.TrimSpace(method))
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return errors.New("el código debe tener 6 dígitos")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("usuario no encontrado")
	}

	allUsers, _ := s.userRepo.FindAllByEmail(ctx, user.Email)
	if len(allUsers) == 0 {
		return errors.New("usuario no encontrado")
	}

	challenge, exists := s.store.GetChallengeByEmail(user.Email)
	if !exists {
		return errors.New("la sesión de configuración de 2FA expiró. Iniciá la configuración nuevamente")
	}

	if method == "totp" {
		if err := validateTOTPSetup(challenge, code); err != nil {
			return err
		}
		if err := s.updateUser2FA(ctx, allUsers, true, challenge.Secret, "totp"); err != nil {
			return errors.New("error al guardar configuración de 2FA")
		}

		s.store.ConsumeChallenge(challenge.Nonce)
		slog.Info("security: 2FA activated with TOTP", slog.String("event", "2fa_enabled"), slog.String("user_id", userID))
		return nil
	}

	if err := s.validateEmailSetup(challenge, code); err != nil {
		return err
	}

	if err := s.updateUser2FA(ctx, allUsers, true, "", "email"); err != nil {
		return errors.New("error al guardar configuración de 2FA")
	}

	s.store.ConsumeChallenge(challenge.Nonce)
	slog.Info("security: 2FA activated with email OTP", slog.String("event", "2fa_enabled"), slog.String("user_id", userID))
	return nil
}

func validateTOTPSetup(challenge *utils.TwofaChallenge, code string) error {
	if challenge.Method != "totp_setup" {
		return errors.New("la sesión de configuración no corresponde al método de aplicación autenticadora")
	}
	if !utils.ValidateTOTPCode(challenge.Secret, code, time.Now()) {
		return errors.New("el código de la aplicación autenticadora es incorrecto")
	}
	return nil
}

func (s *twofaService) validateEmailSetup(challenge *utils.TwofaChallenge, code string) error {
	if challenge.Method != "email_setup" {
		return errors.New("la sesión de configuración no corresponde al método de correo electrónico")
	}
	if !hmac.Equal([]byte(challenge.Code), []byte(code)) {
		left, locked := s.store.RecordFailedAttempt(challenge.Nonce)
		if locked {
			return errors.New("demasiados intentos fallidos. Por favor iniciá la configuración nuevamente")
		}
		return fmt.Errorf("el código ingresado es incorrecto (%d intentos restantes)", left)
	}
	return nil
}

func (s *twofaService) updateUser2FA(ctx context.Context, allUsers []models.User, enabled bool, secret, seed string) error {
	for i := range allUsers {
		allUsers[i].TwofaEnabled = enabled
		allUsers[i].TwofaSecret = secret
		allUsers[i].TwofaSeed = seed
		if err := s.userRepo.Update(ctx, &allUsers[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s *twofaService) Disable(ctx context.Context, userID, password, code string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("usuario no encontrado")
	}

	if !user.TwofaEnabled {
		return errors.New("la autenticación de doble factor no está activada en esta cuenta")
	}

	allUsers, err := s.userRepo.FindAllByEmail(ctx, user.Email)
	if err != nil || len(allUsers) == 0 {
		return errors.New("usuario no encontrado")
	}

	if !isAuthorizedToDisable2FA(user, password, code) {
		return errors.New("para desactivar 2FA debés ingresar tu contraseña actual o un código de verificación válido")
	}

	if err := s.updateUser2FA(ctx, allUsers, false, "", ""); err != nil {
		return errors.New("error al desactivar 2FA")
	}

	slog.Info("security: 2FA disabled", slog.String("event", "2fa_disabled"), slog.String("user_id", userID))
	return nil
}

func isAuthorizedToDisable2FA(user *models.User, password, code string) bool {
	password = strings.TrimSpace(password)
	if password != "" && user.PasswordHash != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err == nil {
			return true
		}
	}

	code = strings.TrimSpace(code)
	if len(code) == 6 && user.TwofaSeed == "totp" && user.TwofaSecret != "" {
		return utils.ValidateTOTPCode(user.TwofaSecret, code, time.Now())
	}

	return false
}

func (s *twofaService) InitiateLoginChallenge(ctx context.Context, email string, userIDs []string, method string) (string, error) {
	method = strings.ToLower(strings.TrimSpace(method))
	if method == "" {
		method = "email"
	}

	var otpCode string
	if method == "email" {
		var err error
		otpCode, err = utils.GenerateNumericOTP(6)
		if err != nil {
			return "", err
		}
	}

	nonce, err := s.store.CreateChallenge(email, userIDs, method, otpCode, "", 10*time.Minute)
	if err != nil {
		return "", err
	}

	if method == "email" && len(userIDs) > 0 && s.producer != nil {
		_ = s.producer.EmailSolicitado(ctx, nonce, userIDs[0], "two_factor_code")
	}

	return nonce, nil
}

func (s *twofaService) VerifyLoginChallenge(ctx context.Context, twofaNonce, code string) (*dto.CheckCredentialsResponse, *models.User, error) {
	twofaNonce = strings.TrimSpace(twofaNonce)
	code = strings.TrimSpace(code)

	challenge, ok := s.store.GetChallenge(twofaNonce)
	if !ok {
		return nil, nil, errors.New("el código de verificación expiró o es inválido. Por favor, iniciá sesión nuevamente")
	}

	users, err := s.userRepo.FindAllByEmail(ctx, challenge.Email)
	if err != nil || len(users) == 0 {
		return nil, nil, errors.New("cuenta no encontrada")
	}
	primaryUser := &users[0]

	valid := false
	if challenge.Method == "totp" {
		if primaryUser.TwofaSecret != "" {
			valid = utils.ValidateTOTPCode(primaryUser.TwofaSecret, code, time.Now())
		}
	} else {
		// Email OTP
		valid = hmac.Equal([]byte(challenge.Code), []byte(code))
	}

	if !valid {
		left, locked := s.store.RecordFailedAttempt(twofaNonce)
		if locked {
			slog.Warn("security: 2FA challenge locked after 5 failed attempts",
				slog.String("event", "2fa_locked"),
				slog.String("email_hash", utils.HashToken(challenge.Email)),
			)
			return nil, nil, errors.New("se superó el número máximo de intentos. Iniciá sesión nuevamente")
		}
		return nil, nil, fmt.Errorf("código de seguridad incorrecto (%d intentos restantes)", left)
	}

	s.store.ConsumeChallenge(twofaNonce)

	// Issue the Step 2 login nonce
	userIDs := make([]string, 0, len(users))
	for _, u := range users {
		userIDs = append(userIDs, u.UserID)
	}
	loginNonce, err := utils.GetNonceStore().Generate(userIDs, 2*time.Minute)
	if err != nil {
		return nil, nil, errors.New("error al generar sesión de ingreso")
	}

	accounts := s.buildAvailableAccounts(ctx, users)

	return &dto.CheckCredentialsResponse{
		Accounts:        accounts,
		LoginNonce:      loginNonce,
		IsEmailVerified: primaryUser.IsEmailVerified,
		Email:           primaryUser.Email,
		RequiresTwofa:   false,
	}, primaryUser, nil
}

func (s *twofaService) ResendLoginOTP(ctx context.Context, twofaNonce string) error {
	challenge, ok := s.store.GetChallenge(twofaNonce)
	if !ok {
		return errors.New("la sesión de verificación expiró. Iniciá sesión nuevamente")
	}

	if challenge.Method != "email" {
		return errors.New("el reenvío de código solo está disponible para autenticación por correo")
	}

	newCode, err := utils.GenerateNumericOTP(6)
	if err != nil {
		return errors.New("error generando nuevo código")
	}

	s.store.UpdateCode(twofaNonce, newCode, 10*time.Minute)

	if len(challenge.UserIDs) > 0 && s.producer != nil {
		_ = s.producer.EmailSolicitado(ctx, twofaNonce, challenge.UserIDs[0], "two_factor_code")
	}

	return nil
}

func (s *twofaService) GetActiveCodeForEmail(email string) string {
	return s.store.GetCodeByEmail(email)
}

// Trusted Device token payload
type trustedDevicePayload struct {
	EmailHash string `json:"eh"`
	UAHash    string `json:"uh"`
	SecStamp  string `json:"ss"`
	ExpiresAt int64  `json:"exp"`
}

func (s *twofaService) CreateTrustedDeviceToken(user *models.User, userAgent string) (string, time.Time, error) {
	ttlDays := defaultTrustedDays
	if envDays := os.Getenv("TRUSTED_DEVICE_TTL_DAYS"); envDays != "" {
		if d, err := strconv.Atoi(envDays); err == nil && d > 0 {
			ttlDays = d
		}
	}
	expiresAt := time.Now().Add(time.Duration(ttlDays*24) * time.Hour)

	payload := trustedDevicePayload{
		EmailHash: hashString(user.Email),
		UAHash:    hashString(userAgent),
		SecStamp:  calculateSecurityStamp(user),
		ExpiresAt: expiresAt.Unix(),
	}

	rawJSON, err := json.Marshal(payload)
	if err != nil {
		return "", time.Time{}, err
	}

	b64Payload := base64.RawURLEncoding.EncodeToString(rawJSON)
	sig := signString(b64Payload)
	token := fmt.Sprintf("%s.%s", b64Payload, sig)

	return token, expiresAt, nil
}

func (s *twofaService) ValidateTrustedDeviceToken(token string, user *models.User, userAgent string) bool {
	if token == "" || user == nil {
		return false
	}

	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return false
	}
	b64Payload, sig := parts[0], parts[1]

	expectedSig := signString(b64Payload)
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return false
	}

	rawJSON, err := base64.RawURLEncoding.DecodeString(b64Payload)
	if err != nil {
		return false
	}

	var payload trustedDevicePayload
	if err := json.Unmarshal(rawJSON, &payload); err != nil {
		return false
	}

	if time.Now().Unix() > payload.ExpiresAt {
		return false
	}

	if payload.EmailHash != hashString(user.Email) {
		return false
	}

	if payload.UAHash != hashString(userAgent) {
		return false
	}

	if payload.SecStamp != calculateSecurityStamp(user) {
		return false
	}

	return true
}

func (s *twofaService) buildAvailableAccounts(ctx context.Context, users []models.User) []dto.AvailableAccount {
	roleNames := map[string]string{
		"role_admin":       "Responsable",
		"role_dev":         "Desarrollador",
		"role_participant": "Participante",
		"role_company":     "Empresa",
		"role_teacher":     "Docente",
		"role_volunteer":   "Voluntario",
	}

	accounts := make([]dto.AvailableAccount, 0, len(users))
	for _, u := range users {
		roleName := roleNames[u.RoleID]
		if roleName == "" {
			roleName = u.RoleID
		}

		accounts = append(accounts, dto.AvailableAccount{
			UserID:          u.UserID,
			RoleID:          u.RoleID,
			RoleName:        roleName,
			Status:          u.Status,
			ApprovalStatus:  s.resolveApprovalStatus(ctx, u.RoleID, u.UserID),
			IsEmailVerified: u.IsEmailVerified,
		})
	}
	return accounts
}

func (s *twofaService) resolveApprovalStatus(ctx context.Context, roleID, userID string) string {
	if s.querier == nil {
		return "approved"
	}
	tableName := roleTableName(roleID)
	if tableName == "" {
		return "approved"
	}
	if st, err := s.querier.GetApprovalStatus(ctx, tableName, userID); err == nil && st != "" {
		return st
	}
	return "approved"
}

func roleTableName(roleID string) string {
	switch roleID {
	case "role_teacher":
		return "teachers"
	case "role_volunteer":
		return "volunteers"
	case "role_participant":
		return "participants"
	case "role_company":
		return "companies"
	default:
		return ""
	}
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(strings.TrimSpace(s)))
	return hex.EncodeToString(h[:])
}

func calculateSecurityStamp(user *models.User) string {
	raw := fmt.Sprintf("%s:%s:%s", user.PasswordHash, user.TwofaSecret, user.TwofaSeed)
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:16])
}

func getSigningSecret() []byte {
	secret := os.Getenv("TRUSTED_DEVICE_SECRET")
	if secret == "" {
		secret = os.Getenv("JWT_SECRET")
	}
	if secret == "" {
		secret = "buenvivir-trusted-device-fallback-key-2026"
	}
	return []byte(secret)
}

func signString(data string) string {
	h := hmac.New(sha256.New, getSigningSecret())
	_, _ = h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
