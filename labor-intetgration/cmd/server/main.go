package main

import (
	"log"
	"net/http"

	"buenvivir/labor-integration/internal/config"
	"buenvivir/labor-integration/internal/controllers"
	"buenvivir/labor-integration/internal/repositories"
	"buenvivir/labor-integration/internal/routes"
	"buenvivir/labor-integration/internal/services"
)

func main() {
	cfg := config.LoadConfig()

	repo := repositories.NewJobRepository()
	service := services.NewJobService(repo)
	controller := controllers.NewJobController(service)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, controller)

	log.Printf("Servicio Labor Integration escuchando en :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
