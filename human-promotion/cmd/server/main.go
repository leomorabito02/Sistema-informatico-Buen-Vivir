package main

import (
	"log"
	"net/http"

	"buenvivir/human-promotion/internal/config"
	"buenvivir/human-promotion/internal/controllers"
	"buenvivir/human-promotion/internal/repositories"
	"buenvivir/human-promotion/internal/routes"
	"buenvivir/human-promotion/internal/services"
)

func main() {
	cfg := config.LoadConfig()

	repo := repositories.NewPromotionRepository()
	service := services.NewPromotionService(repo)
	controller := controllers.NewPromotionController(service)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, controller)

	log.Printf("Servicio Human Promotion escuchando en :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
