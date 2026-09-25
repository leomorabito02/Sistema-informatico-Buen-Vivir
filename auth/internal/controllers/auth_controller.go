package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/go-chi/chi/v5"

	"auth/internal/dto"
	"auth/internal/middlewares"
	"auth/internal/services"
	"auth/internal/utils"
)

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
	errUnauthorized   = "unauthorized"
	cookieHostAccess  = "__Host-access_token"
	cookieHostRefresh = "__Host-refresh_token"
	msgInvalidReqBody = "Cuerpo de solicitud inválido"
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

	c.setAuthCookies(w, resp.Token, resp.RefreshToken)
	writeJSONResponse(w, http.StatusCreated, resp)
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

	status := c.authService.GetApprovalStatus(r.Context(), roleID, userID)

	writeJSONResponse(w, http.StatusOK, map[string]any{
		"user_id":         userID,
		"role_id":         roleID,
		"approval_status": status,
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
