package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"auth/internal/dto"
	"auth/internal/middlewares"
	"auth/internal/services"
	"auth/internal/utils"
)

const (
	contentTypeHeader   = "Content-Type"
	contentTypeJSON     = "application/json"
	errUnauthorized     = "unauthorized"
	cookieHostAccess    = "__Host-access_token"
	cookieHostRefresh   = "__Host-refresh_token"
	cookieTrustedDevice = "__Host-trusted_device"
	msgInvalidReqBody   = "Cuerpo de solicitud inválido"
	msgInvalidReqFormat = "El formato de la solicitud es inválido."
	msgTwofaUnavailable = "Servicio de 2FA no disponible."
)

func writeJSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSONResponse(w, status, map[string]string{"error": message})
}

// AuthController representa el Controlador (C de MVC)
type AuthController struct {
	authService             services.AuthService
	verificationService     services.VerificationService
	passwordRecoveryService services.PasswordRecoveryService
	twofaService            services.TwofaService
	validate                *validator.Validate
}

func NewAuthController(authService services.AuthService, opts ...any) *AuthController {
	ctrl := &AuthController{
		authService: authService,
		validate:    utils.GetValidator(),
	}
	for _, opt := range opts {
		if vs, ok := opt.(services.VerificationService); ok {
			ctrl.verificationService = vs
		}
		if prs, ok := opt.(services.PasswordRecoveryService); ok {
			ctrl.passwordRecoveryService = prs
		}
		if tfa, ok := opt.(services.TwofaService); ok {
			ctrl.twofaService = tfa
		}
	}
	return ctrl
}

func (c *AuthController) setAuthCookies(w http.ResponseWriter, token, refreshToken string) {
	// #nosec G124 -- SameSiteNoneMode is required for cross-origin requests
	http.SetCookie(w, &http.Cookie{
		Name:     cookieHostAccess,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   900, // 15 minutes
	})
	// #nosec G124 -- SameSiteNoneMode is required for cross-origin requests
	http.SetCookie(w, &http.Cookie{
		Name:     cookieHostRefresh,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   7 * 24 * 3600, // 7 days
	})
}

func (c *AuthController) clearAuthCookies(w http.ResponseWriter) {
	// #nosec G124 -- SameSiteNoneMode is required for cross-origin requests
	http.SetCookie(w, &http.Cookie{
		Name:     cookieHostAccess,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	})
	// #nosec G124 -- SameSiteNoneMode is required for cross-origin requests
	http.SetCookie(w, &http.Cookie{
		Name:     cookieHostRefresh,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	})
}

func (c *AuthController) setTrustedDeviceCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	// #nosec G124 -- SameSiteNoneMode is required for cross-origin requests
	http.SetCookie(w, &http.Cookie{
		Name:     cookieTrustedDevice,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   maxAge,
	})
}

func (c *AuthController) clearTrustedDeviceCookie(w http.ResponseWriter) {
	// #nosec G124 -- SameSiteNoneMode is required for cross-origin requests
	http.SetCookie(w, &http.Cookie{
		Name:     cookieTrustedDevice,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	})
}

func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max body
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "El formato de la solicitud de registro es inválido.")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Phone = strings.TrimSpace(req.Phone)

	if err := c.validate.Struct(req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Datos de registro inválidos: revise nombre, apellido, email, teléfono y contraseña (mínimo 8 caracteres, con mayúscula, minúscula y carácter especial).")
		return
	}

	resp, err := c.authService.Register(r.Context(), &req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var userID, email string
	var isEmailVerified bool
	if resp != nil && resp.User != nil {
		userID = resp.User.UserID
		email = resp.User.Email
		isEmailVerified = resp.User.IsEmailVerified
	}

	approvalStatus := ""
	if resp != nil {
		approvalStatus = resp.ApprovalStatus
	}

	writeJSONResponse(w, http.StatusCreated, dto.RegisterResponse{
		Message:         "Registro exitoso. Se ha enviado un correo electrónico de verificación.",
		UserID:          userID,
		Email:           email,
		IsEmailVerified: isEmailVerified,
		ApprovalStatus:  approvalStatus,
	})
}

// CheckCredentials handles step 1 of login: validates email + password and returns available accounts/roles.
func (c *AuthController) CheckCredentials(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max body
	var req dto.CheckCredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "El formato de la solicitud de inicio de sesión es inválido.")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if cookie, err := r.Cookie(cookieTrustedDevice); err == nil {
		req.TrustedDevice = cookie.Value
	}
	req.UserAgent = r.UserAgent()

	if err := c.validate.Struct(req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Debe ingresar un correo electrónico válido y su contraseña.")
		return
	}

	resp, err := c.authService.CheckCredentials(r.Context(), &req)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

// CheckGoogleCredentials handles step 1 of Google login: validates Firebase token and returns available accounts/roles.
func (c *AuthController) CheckGoogleCredentials(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max body
	var req dto.GoogleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "El formato de la solicitud de Google es inválido.")
		return
	}

	if cookie, err := r.Cookie(cookieTrustedDevice); err == nil {
		req.TrustedDevice = cookie.Value
	}
	req.UserAgent = r.UserAgent()

	resp, err := c.authService.CheckGoogleCredentials(r.Context(), &req)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

// Login handles step 2: given a user_id and role_id (no password), issues the JWT token.
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max body
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "El formato de la solicitud de selección de rol es inválido.")
		return
	}

	if err := c.validate.Struct(req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "El usuario y el rol son obligatorios.")
		return
	}

	resp, err := c.authService.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, services.ErrEmailNotVerified) || err.Error() == "EMAIL_NOT_VERIFIED" {
			writeJSONResponse(w, http.StatusForbidden, map[string]interface{}{
				"error":   "EMAIL_NOT_VERIFIED",
				"message": "Debes verificar tu correo electrónico antes de iniciar sesión.",
			})
			return
		}
		if errors.Is(err, services.ErrAccountPendingApproval) || err.Error() == "ACCOUNT_PENDING_APPROVAL" {
			writeJSONResponse(w, http.StatusForbidden, map[string]interface{}{
				"error":   "ACCOUNT_PENDING_APPROVAL",
				"message": "Tu cuenta está pendiente de aprobación por parte de un administrador.",
			})
			return
		}
		if errors.Is(err, services.ErrAccountBlocked) || err.Error() == "ACCOUNT_BLOCKED" {
			writeJSONResponse(w, http.StatusForbidden, map[string]interface{}{
				"error":   "ACCOUNT_BLOCKED",
				"message": "Tu cuenta se encuentra bloqueada.",
			})
			return
		}
		writeJSONError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if resp.UserStatus == "blocked" {
		writeJSONResponse(w, http.StatusOK, resp)
		return
	}
	c.setAuthCookies(w, resp.Token, resp.RefreshToken)
	writeJSONResponse(w, http.StatusOK, resp)
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	c.clearAuthCookies(w)
	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Logout exitoso"})
}

func (c *AuthController) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, errUnauthorized, http.StatusUnauthorized)
		return
	}
	roleID, _ := r.Context().Value(middlewares.RoleIDKey).(string)
	// is_external viene del token: el frontend lo usa para mostrarle al
	// docente externo solo la carga y consulta de contenido (HU-23).
	isExternal, _ := r.Context().Value(middlewares.IsExternalKey).(bool)

	status := c.authService.GetApprovalStatus(r.Context(), roleID, userID)

	writeJSONResponse(w, http.StatusOK, map[string]any{
		"user_id":         userID,
		"role_id":         roleID,
		"approval_status": status,
		"is_external":     isExternal,
	})
}

func (c *AuthController) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, errUnauthorized, http.StatusUnauthorized)
		return
	}
	roleID, ok := r.Context().Value(middlewares.RoleIDKey).(string)
	if !ok {
		http.Error(w, errUnauthorized, http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max body
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, msgInvalidReqBody, http.StatusBadRequest)
		return
	}

	if err := c.authService.UpdateMyProfile(r.Context(), roleID, userID, data); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Perfil actualizado correctamente"})
}

func (c *AuthController) GetMyProfileDetails(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, errUnauthorized, http.StatusUnauthorized)
		return
	}
	roleID, ok := r.Context().Value(middlewares.RoleIDKey).(string)
	if !ok {
		http.Error(w, errUnauthorized, http.StatusUnauthorized)
		return
	}

	profile, err := c.authService.GetMyProfile(r.Context(), roleID, userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, profile)
}

func (c *AuthController) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSONError(w, http.StatusBadRequest, "ID es requerido")
		return
	}

	profile, err := c.authService.GetUserProfile(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    profile,
	})
}

func (c *AuthController) UpdateMyPassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, errUnauthorized, http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max body
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, msgInvalidReqBody, http.StatusBadRequest)
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		writeJSONError(w, http.StatusBadRequest, "Las contraseñas son obligatorias")
		return
	}

	if err := c.authService.UpdateMyPassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Contraseña actualizada correctamente"})
}

func (c *AuthController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	writeJSONResponse(w, http.StatusOK, map[string]string{
		"status":  "UP",
		"service": "auth-service",
	})
}

func (c *AuthController) LinkGoogleAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, errUnauthorized, http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.LinkGoogleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, msgInvalidReqBody)
		return
	}

	if err := c.authService.LinkGoogleAccount(r.Context(), userID, &req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Cuenta de Google vinculada correctamente"})
}

func (c *AuthController) UnlinkGoogleAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, errUnauthorized, http.StatusUnauthorized)
		return
	}

	if err := c.authService.UnlinkGoogleAccount(r.Context(), userID); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Cuenta de Google desvinculada correctamente"})
}

func (c *AuthController) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, msgInvalidReqFormat)
		return
	}

	if err := c.validate.Struct(req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "El token de verificación es obligatorio.")
		return
	}

	if c.verificationService == nil {
		writeJSONError(w, http.StatusInternalServerError, "Servicio de verificación no disponible.")
		return
	}

	if err := c.verificationService.VerifyEmail(r.Context(), req.Token); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":           "Correo electrónico verificado exitosamente.",
		"is_email_verified": true,
	})
}

func (c *AuthController) ResendVerification(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, msgInvalidReqFormat)
		return
	}

	if err := c.validate.Struct(req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Debe ingresar una dirección de correo electrónico válida.")
		return
	}

	if c.verificationService == nil {
		writeJSONError(w, http.StatusInternalServerError, "Servicio de verificación no disponible.")
		return
	}

	if err := c.verificationService.ResendVerification(r.Context(), req.Email); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Se ha reenviado el correo electrónico de verificación.",
	})
}

func (c *AuthController) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, msgInvalidReqFormat)
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if err := c.validate.Struct(req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Debe ingresar una dirección de correo electrónico válida.")
		return
	}

	if c.passwordRecoveryService == nil {
		writeJSONError(w, http.StatusInternalServerError, "Servicio de recuperación no disponible.")
		return
	}

	_ = c.passwordRecoveryService.RequestReset(r.Context(), req.Email)

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Si la dirección de correo electrónico está registrada, recibirás un enlace para restablecer tu contraseña.",
	})
}

func (c *AuthController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, msgInvalidReqFormat)
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	if err := c.validate.Struct(req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Datos inválidos: revise el enlace y la contraseña (mínimo 8 caracteres, con mayúscula, minúscula y carácter especial).")
		return
	}

	if c.passwordRecoveryService == nil {
		writeJSONError(w, http.StatusInternalServerError, "Servicio de recuperación no disponible.")
		return
	}

	if err := c.passwordRecoveryService.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"message": "Contraseña restablecida exitosamente. Ya podés iniciar sesión con tu nueva contraseña.",
	})
}

// VerifyTwofa verifies 2FA code during login and sets trusted device cookie if requested
func (c *AuthController) VerifyTwofa(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.VerifyTwofaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, msgInvalidReqFormat)
		return
	}

	req.TwofaNonce = strings.TrimSpace(req.TwofaNonce)
	req.Code = strings.TrimSpace(req.Code)

	if err := c.validate.Struct(req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Debe ingresar un código de 6 dígitos válido.")
		return
	}

	if c.twofaService == nil {
		writeJSONError(w, http.StatusInternalServerError, msgTwofaUnavailable)
		return
	}

	resp, user, err := c.twofaService.VerifyLoginChallenge(r.Context(), req.TwofaNonce, req.Code)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, err.Error())
		return
	}

	if req.TrustDevice && user != nil {
		if token, exp, tokenErr := c.twofaService.CreateTrustedDeviceToken(user, r.UserAgent()); tokenErr == nil {
			c.setTrustedDeviceCookie(w, token, exp)
		}
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

// ResendTwofa sends a new OTP code for an active email 2FA login challenge
func (c *AuthController) ResendTwofa(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.ResendTwofaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, msgInvalidReqFormat)
		return
	}

	req.TwofaNonce = strings.TrimSpace(req.TwofaNonce)
	if req.TwofaNonce == "" {
		writeJSONError(w, http.StatusBadRequest, "Identificador de sesión de 2FA requerido.")
		return
	}

	if c.twofaService == nil {
		writeJSONError(w, http.StatusInternalServerError, msgTwofaUnavailable)
		return
	}

	if err := c.twofaService.ResendLoginOTP(r.Context(), req.TwofaNonce); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Se ha reenviado el código de seguridad a tu correo electrónico.",
	})
}

// SetupTwofa initiates the 2FA configuration in user profile
func (c *AuthController) SetupTwofa(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, errUnauthorized, http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.SetupTwofaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, msgInvalidReqFormat)
		return
	}

	req.Method = strings.TrimSpace(req.Method)
	if err := c.validate.Struct(req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "El método de 2FA debe ser 'email' o 'totp'.")
		return
	}

	if c.twofaService == nil {
		writeJSONError(w, http.StatusInternalServerError, msgTwofaUnavailable)
		return
	}

	resp, err := c.twofaService.InitiateSetup(r.Context(), userID, req.Method)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, resp)
}

// EnableTwofa validates the confirmation code and turns on 2FA
func (c *AuthController) EnableTwofa(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, errUnauthorized, http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.EnableTwofaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, msgInvalidReqFormat)
		return
	}

	req.Method = strings.TrimSpace(req.Method)
	req.Code = strings.TrimSpace(req.Code)
	if err := c.validate.Struct(req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Datos inválidos: ingrese un método válido y el código de 6 dígitos.")
		return
	}

	if c.twofaService == nil {
		writeJSONError(w, http.StatusInternalServerError, msgTwofaUnavailable)
		return
	}

	if err := c.twofaService.ConfirmEnable(r.Context(), userID, req.Method, req.Code); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Autenticación de dos factores activada exitosamente.",
	})
}

// DisableTwofa disables 2FA after password or verification code check
func (c *AuthController) DisableTwofa(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, errUnauthorized, http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.DisableTwofaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, msgInvalidReqFormat)
		return
	}

	if c.twofaService == nil {
		writeJSONError(w, http.StatusInternalServerError, msgTwofaUnavailable)
		return
	}

	if err := c.twofaService.Disable(r.Context(), userID, req.Password, req.Code); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	c.clearTrustedDeviceCookie(w)

	writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Autenticación de dos factores desactivada exitosamente.",
	})
}

// RevokeTrustedDevice clears the trusted device cookie
func (c *AuthController) RevokeTrustedDevice(w http.ResponseWriter, r *http.Request) {
	c.clearTrustedDeviceCookie(w)
	writeJSONResponse(w, http.StatusOK, map[string]string{
		"message": "Dispositivo de confianza olvidado exitosamente.",
	})
}
