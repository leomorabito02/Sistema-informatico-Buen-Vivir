package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"google.golang.org/api/option"

	"buenvivir/campaigns/internal/config"
	"buenvivir/campaigns/internal/controllers"
	"buenvivir/campaigns/internal/events"
	"buenvivir/campaigns/internal/events/producers"
	"buenvivir/campaigns/internal/middlewares"
	"buenvivir/campaigns/internal/observability"
	"buenvivir/campaigns/internal/outbox"
	"buenvivir/campaigns/internal/repositories"
	"buenvivir/campaigns/internal/routes"
	"buenvivir/campaigns/internal/services"
)

func main() {
	// Logs en JSON a stdout: Cloud Run los manda tal cual a Cloud Logging.
	observability.InitLogger("campaigns", os.Getenv("APP_ENV"))
	metrics := observability.NewRegistry("campaigns")

	cfg := config.LoadConfig()
	ctx := context.Background()

	jwtService := services.NewJWTService()

	googleCredentials, err := repositories.GoogleCredentialsJSON()
	if err != nil {
		slog.Error("credenciales de Google inválidas", slog.String("error", err.Error()))
		os.Exit(1)
	}

	firestoreClient, err := repositories.NewFirestoreClient(ctx)
	if err != nil {
		slog.Error("no se pudo inicializar Firestore", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer func() { _ = firestoreClient.Close() }()

	relayCtx, cancelRelay := context.WithCancel(ctx)
	relay := startRelay(relayCtx, cfg, firestoreClient, googleCredentials)
	var writer *outbox.Writer
	if relay != nil {
		writer = outbox.NewWriter(firestoreClient, events.ProducerCampanas, relay.NotifyInline)
	}
	campanasProducer := producers.NewCampanasProducer(writer, cfg.Events.TopicCampanas)

	service := services.NewCampaignService(
		repositories.NewCampaignRepository(firestoreClient),
		services.WithCampanasProducer(campanasProducer),
	)

	r := chi.NewRouter()
	r.Use(observability.HTTPMiddleware(metrics, func(req *http.Request) string {
		if rc := chi.RouteContext(req.Context()); rc != nil {
			return rc.RoutePattern()
		}
		return ""
	}))
	middlewares.SetupMiddlewares(r, cfg.CORSAllowedOrigins)
	if observability.MetricsEnabled() {
		r.Handle("/metrics", metrics.Handler())
	}
	routes.RegisterRoutes(r, jwtService, controllers.NewCampaignController(service).WithAudit(repositories.NewAuditRepository(firestoreClient)))

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
	slog.Info("Servicio Campaigns escuchando", slog.String("port", cfg.Port))
	if err := observability.ListenAndServeGraceful(ctx, server, func(shutdownCtx context.Context) {
		cancelRelay()
		relay.Shutdown(shutdownCtx)
	}); err != nil {
		slog.Error("error al iniciar el servidor", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func startRelay(ctx context.Context, cfg *config.Config, client *firestore.Client, googleCredentials []byte) *outbox.Relay {
	if !cfg.Events.PublishEnabled() {
		slog.Info("PUBSUB_PROJECT_ID vacío: campaigns no publica eventos")
		return nil
	}
	var opts []option.ClientOption
	if googleCredentials != nil && strings.TrimSpace(os.Getenv("PUBSUB_EMULATOR_HOST")) == "" {
		opts = append(opts, option.WithCredentialsJSON(googleCredentials)) //nolint:staticcheck // inline credentials injected in Cloud Run CD
	}
	publisher, err := outbox.NewPubSubPublisher(ctx, cfg.Events.ProjectID, opts...)
	if err != nil {
		slog.Error("no se pudo inicializar Pub/Sub", slog.String("error", err.Error()))
		os.Exit(1)
	}
	relay := outbox.NewRelay(client, publisher, slog.Default())
	go relay.Run(ctx)
	return relay
}
