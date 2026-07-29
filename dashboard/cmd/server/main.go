package main

import (
	"log"
	"net/http"

	"buenvivir/dashboard/internal/config"
	"buenvivir/dashboard/internal/controllers"
	"buenvivir/dashboard/internal/repositories"
	"buenvivir/dashboard/internal/routes"
	"buenvivir/dashboard/internal/services"
)

func main() {
	cfg := config.LoadConfig()

	repo := repositories.NewDashboardRepository()
	service := services.NewDashboardService(repo)
	controller := controllers.NewDashboardController(service)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, controller)

	log.Printf("Servicio Dashboard escuchando en :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
