package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"

	"buenvivir/campaigns/internal/config"
	"buenvivir/campaigns/internal/controllers"
	"buenvivir/campaigns/internal/events"
	"buenvivir/campaigns/internal/events/producers"
	"buenvivir/campaigns/internal/middlewares"
	"buenvivir/campaigns/internal/outbox"
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

	relay := startRelay(ctx, cfg, firestoreClient)
	var writer *outbox.Writer
	if relay != nil {
		writer = outbox.NewWriter(firestoreClient, events.ProducerCampanas, relay.Notify)
	}
	campanasProducer := producers.NewCampanasProducer(writer, cfg.Events.TopicCampanas)

	service := services.NewCampaignService(
		repositories.NewCampaignRepository(firestoreClient),
		services.WithCampanasProducer(campanasProducer),
	)

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

func startRelay(ctx context.Context, cfg *config.Config, client *firestore.Client) *outbox.Relay {
	if !cfg.Events.PublishEnabled() {
		slog.Info("PUBSUB_PROJECT_ID vacío: campaigns no publica eventos")
		return nil
	}
	publisher, err := outbox.NewPubSubPublisher(ctx, cfg.Events.ProjectID)
	if err != nil {
		slog.Error("no se pudo inicializar Pub/Sub", slog.String("error", err.Error()))
		os.Exit(1)
	}
	relay := outbox.NewRelay(client, publisher, slog.Default())
	go relay.Run(ctx)
	return relay
}
