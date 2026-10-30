package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"

	"buenvivir/campaigns/internal/config"
	"buenvivir/campaigns/internal/controllers"
	"buenvivir/campaigns/internal/middlewares"
	"buenvivir/campaigns/internal/repositories"
	"buenvivir/campaigns/internal/routes"
	"buenvivir/campaigns/internal/services"
)

func main() {
	// Logs en JSON a stdout: Cloud Run los manda tal cual a Cloud Logging.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.LoadConfig()
	ctx := context.Background()

	jwtService := services.NewJWTService()

	firestoreClient, err := repositories.NewFirestoreClient(ctx)
	if err != nil {
		slog.Error("no se pudo inicializar Firestore", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() { _ = firestoreClient.Close() }()

	service := services.NewCampaignService(repositories.NewCampaignRepository(firestoreClient))

	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r, cfg.CORSAllowedOrigins)
	routes.RegisterRoutes(r, jwtService, controllers.NewCampaignController(service))

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
	slog.Info("Servicio Campaigns escuchando", slog.String("port", cfg.Port))
	if err := server.ListenAndServe(); err != nil {
		slog.Error("error al iniciar el servidor", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
