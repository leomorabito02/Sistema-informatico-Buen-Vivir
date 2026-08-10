package routes

import (
	"net/http"

	"auth/internal/controllers"
)

func RegisterRoutes(mux *http.ServeMux, authCtrl *controllers.AuthController) {
	mux.HandleFunc("/health", authCtrl.HealthCheck)
	mux.HandleFunc("/api/v1/auth/login", authCtrl.Login)
	mux.HandleFunc("/api/v1/auth/register", authCtrl.Register)
	mux.HandleFunc("/api/v1/auth/google-login", authCtrl.LoginWithGoogle)
}
