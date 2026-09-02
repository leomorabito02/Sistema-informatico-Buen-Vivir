package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"auth/internal/dto"
	"auth/internal/middlewares"
	"auth/internal/services"
	"auth/internal/utils"
)

// AuthController representa el Controlador (C de MVC)
type AuthController struct {
	authService services.AuthService
	validate    *validator.Validate
}

func NewAuthController(authService services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
		validate:    utils.GetValidator(),
	}
}

func (c *AuthController) setAuthCookies(w http.ResponseWriter, token, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "__Host-access_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   900, // 15 minutes
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "__Host-refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   7 * 24 * 3600, // 7 days
	})
}

func (c *AuthController) clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "__Host-access_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "__Host-refresh_token",
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "El formato de la solicitud de registro es inválido."})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	req.Phone = strings.TrimSpace(req.Phone)

	if err := c.validate.Struct(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Datos de registro inválidos: revise nombre, apellido, email, teléfono y contraseña (mínimo 8 caracteres, con mayúscula, minúscula y carácter especial)."})
		return
	}

	resp, err := c.authService.Register(r.Context(), &req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	c.setAuthCookies(w, resp.Token, resp.RefreshToken)
	w.WriteHeader(http.StatusCreated)
	// #nosec G117
	_ = json.NewEncoder(w).Encode(resp)
}

// CheckCredentials handles step 1 of login: validates email + password and returns available accounts/roles.
func (c *AuthController) CheckCredentials(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max body
	var req dto.CheckCredentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "El formato de la solicitud de inicio de sesión es inválido."})
		return
	}

	req.Email = strings.TrimSpace(req.Email)

	if err := c.validate.Struct(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Debe ingresar un correo electrónico válido y su contraseña."})
		return
	}

	resp, err := c.authService.CheckCredentials(r.Context(), &req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// CheckGoogleCredentials handles step 1 of Google login: validates Firebase token and returns available accounts/roles.
func (c *AuthController) CheckGoogleCredentials(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max body
	var req dto.GoogleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "El formato de la solicitud de Google es inválido."})
		return
	}

	resp, err := c.authService.CheckGoogleCredentials(r.Context(), &req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Login handles step 2: given a user_id and role_id (no password), issues the JWT token.
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max body
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "El formato de la solicitud de selección de rol es inválido."})
		return
	}

	if err := c.validate.Struct(req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "El usuario y el rol son obligatorios."})
		return
	}

	resp, err := c.authService.Login(r.Context(), &req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if resp.UserStatus == "blocked" {
		w.WriteHeader(http.StatusOK)
		// #nosec G117
		_ = json.NewEncoder(w).Encode(resp)
		return
	}
	c.setAuthCookies(w, resp.Token, resp.RefreshToken)
	w.WriteHeader(http.StatusOK)
	// #nosec G117
	_ = json.NewEncoder(w).Encode(resp)
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	c.clearAuthCookies(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Logout exitoso"})
}

func (c *AuthController) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	roleID, _ := r.Context().Value(middlewares.RoleIDKey).(string)

	status := c.authService.GetApprovalStatus(r.Context(), roleID, userID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// #nosec G117
	if err := json.NewEncoder(w).Encode(map[string]any{
		"user_id":         userID,
		"role_id":         roleID,
		"approval_status": status,
	}); err != nil {
		// el header ya se envió, así que solo podés loguear, no reenviar status
		http.Error(w, "error encoding response: %v", http.StatusInternalServerError)
	}
}

func (c *AuthController) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	roleID, ok := r.Context().Value(middlewares.RoleIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max body
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Cuerpo de solicitud inválido", http.StatusBadRequest)
		return
	}

	if err := c.authService.UpdateMyProfile(r.Context(), roleID, userID, data); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Perfil actualizado correctamente"})
}

func (c *AuthController) GetMyProfileDetails(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	roleID, ok := r.Context().Value(middlewares.RoleIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := c.authService.GetMyProfile(r.Context(), roleID, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(profile)
}

func (c *AuthController) UpdateMyPassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB max body
	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Cuerpo de solicitud inválido", http.StatusBadRequest)
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Las contraseñas son obligatorias"})
		return
	}

	if err := c.authService.UpdateMyPassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Contraseña actualizada correctamente"})
}

func (c *AuthController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// #nosec G117
	if err := json.NewEncoder(w).Encode(map[string]string{
		"status":  "UP",
		"service": "auth-service",
	}); err != nil {
		http.Error(w, "error encoding response: %v", http.StatusInternalServerError)
	}
}

func (c *AuthController) LinkGoogleAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.LinkGoogleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Cuerpo de solicitud inválido"})
		return
	}

	if err := c.authService.LinkGoogleAccount(r.Context(), userID, &req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Cuenta de Google vinculada correctamente"})
}

func (c *AuthController) UnlinkGoogleAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if err := c.authService.UnlinkGoogleAccount(r.Context(), userID); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Cuenta de Google desvinculada correctamente"})
}
