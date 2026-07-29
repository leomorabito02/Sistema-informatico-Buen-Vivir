package main

import (
	"log"
	"net/http"

	"buenvivir/campaigns/internal/config"
	"buenvivir/campaigns/internal/controllers"
	"buenvivir/campaigns/internal/repositories"
	"buenvivir/campaigns/internal/routes"
	"buenvivir/campaigns/internal/services"
)

func main() {
	cfg := config.LoadConfig()

	repo := repositories.NewCampaignRepository()
	service := services.NewCampaignService(repo)
	controller := controllers.NewCampaignController(service)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, controller)

	log.Printf("Servicio Campaigns escuchando en :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
