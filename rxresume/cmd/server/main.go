package main

import (
	"log"
	"net/http"

	"buenvivir/rxresume/internal/config"
	"buenvivir/rxresume/internal/controllers"
	"buenvivir/rxresume/internal/repositories"
	"buenvivir/rxresume/internal/routes"
	"buenvivir/rxresume/internal/services"
)

func main() {
	cfg := config.LoadConfig()

	repo := repositories.NewResumeRepository()
	service := services.NewResumeService(repo)
	controller := controllers.NewResumeController(service)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, controller)

	log.Printf("Servicio RxResume escuchando en :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
