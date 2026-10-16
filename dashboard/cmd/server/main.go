package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"buenvivir/dashboard/internal/clients"
	"buenvivir/dashboard/internal/config"
	"buenvivir/dashboard/internal/controllers"
	"buenvivir/dashboard/internal/events"
	"buenvivir/dashboard/internal/events/consumers"
	"buenvivir/dashboard/internal/routes"
	"buenvivir/dashboard/internal/services"
)

func main() {
	// Logs en JSON (Cloud Run los manda a Cloud Logging).
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	cfg := config.LoadConfig()

	service := services.NewCachedDashboardService(services.NewDashboardService(
		clients.NewUpstream("autenticación", cfg.AuthURL, cfg.Timeout),
		clients.NewUpstream("educación", cfg.EducationURL, cfg.Timeout),
		clients.NewUpstream("inserción laboral", cfg.LaborURL, cfg.Timeout),
	), cfg.CacheTTL)
	controller := controllers.NewDashboardController(service)

	var eventsHandler http.Handler
	if cfg.ConsumeEnabled() {
		verifier, err := events.NewPushVerifier(cfg.PushAuth, cfg.PushAudience, cfg.PushServiceAccounts, cfg.Production)
		if err != nil {
			log.Fatalf("configuración de Pub/Sub inválida: %v", err)
		}
		eventsHandler = events.CorrelationMiddleware(consumers.NewDashboard(verifier, events.NewMemoryStore(1000), service, slog.Default()))
	}

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, controller, eventsHandler)

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
