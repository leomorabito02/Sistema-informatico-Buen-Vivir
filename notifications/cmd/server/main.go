package main

import (
	"log"
	"net/http"

	"buenvivir/notifications/internal/config"
	"buenvivir/notifications/internal/controllers"
	"buenvivir/notifications/internal/repositories"
	"buenvivir/notifications/internal/routes"
	"buenvivir/notifications/internal/services"
)

func main() {
	cfg := config.LoadConfig()

	repo := repositories.NewNotificationRepository()
	service := services.NewNotificationService(repo)
	controller := controllers.NewNotificationController(service)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, controller)

	log.Printf("Servicio Notifications escuchando en :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
