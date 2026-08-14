package routes

import (
	"net/http"

	"auth/internal/controllers"
)

// RegisterRoutes is a legacy helper kept for compatibility; main.go uses chi directly.
func RegisterRoutes(mux *http.ServeMux, authCtrl *controllers.AuthController) {
	mux.HandleFunc("/health", authCtrl.HealthCheck)
	mux.HandleFunc("/api/v1/auth/check-credentials", authCtrl.CheckCredentials)
	mux.HandleFunc("/api/v1/auth/google-check-credentials", authCtrl.CheckGoogleCredentials)
	mux.HandleFunc("/api/v1/auth/login", authCtrl.Login)
	mux.HandleFunc("/api/v1/auth/register", authCtrl.Register)
	mux.HandleFunc("/api/v1/auth/logout", authCtrl.Logout)
}
