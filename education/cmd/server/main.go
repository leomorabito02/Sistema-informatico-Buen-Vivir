package main

import (
	"log"
	"net/http"

	"buenvivir/education/internal/config"
	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/repositories"
	"buenvivir/education/internal/routes"
	"buenvivir/education/internal/services"
)

func main() {
	cfg := config.LoadConfig()

	repo := repositories.NewEducationRepository()
	service := services.NewEducationService(repo)
	controller := controllers.NewEducationController(service)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, controller)

	log.Printf("Servicio Education escuchando en :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
