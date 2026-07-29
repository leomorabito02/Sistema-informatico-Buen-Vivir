package routes

import (
	"net/http"

	"buenvivir/auth/internal/controllers"
)

func RegisterRoutes(mux *http.ServeMux, authCtrl *controllers.AuthController) {
	mux.HandleFunc("/health", authCtrl.HealthCheck)
	mux.HandleFunc("/api/v1/auth/login", authCtrl.Login)
}
