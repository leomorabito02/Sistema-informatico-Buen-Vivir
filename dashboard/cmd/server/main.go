package main

import (
	"log"
	"net/http"
	"time"

	"buenvivir/dashboard/internal/clients"
	"buenvivir/dashboard/internal/config"
	"buenvivir/dashboard/internal/controllers"
	"buenvivir/dashboard/internal/routes"
	"buenvivir/dashboard/internal/services"
)

func main() {
	cfg := config.LoadConfig()

	service := services.NewDashboardService(
		clients.NewUpstream("autenticación", cfg.AuthURL, cfg.Timeout),
		clients.NewUpstream("educación", cfg.EducationURL, cfg.Timeout),
		clients.NewUpstream("inserción laboral", cfg.LaborURL, cfg.Timeout),
	)
	controller := controllers.NewDashboardController(service)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, controller)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		// Cursos con muchas actividades pueden tardar: margen sobre el
		// timeout de cada consulta a un servicio de origen.
		WriteTimeout: 60 * time.Second,
	}
	log.Printf("Servicio Dashboard escuchando en :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}
