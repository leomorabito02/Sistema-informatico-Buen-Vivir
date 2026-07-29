package main

import (
	"log"
	"net/http"

	"buenvivir/auth/internal/config"
	"buenvivir/auth/internal/controllers"
	"buenvivir/auth/internal/repositories"
	"buenvivir/auth/internal/routes"
	"buenvivir/auth/internal/services"
)

func main() {
	cfg := config.LoadConfig()

	// Repositorios, Servicios y Controladores
	userRepo := repositories.NewUserRepository()
	authService := services.NewAuthService(userRepo, cfg.JWTSecret)
	authController := controllers.NewAuthController(authService)

	// Router HTTP
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, authController)

	log.Printf("Servicio Auth escuchando en el puerto :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
