package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"buenvivir/labor-integration/internal/config"
	"buenvivir/labor-integration/internal/controllers"
	"buenvivir/labor-integration/internal/events"
	"buenvivir/labor-integration/internal/events/producers"
	"buenvivir/labor-integration/internal/outbox"
	"buenvivir/labor-integration/internal/repositories"
	"buenvivir/labor-integration/internal/routes"
	"buenvivir/labor-integration/internal/services"
)

func main() {
	ctx := context.Background()

	firestoreClient, err := config.NewFirestoreClient(ctx)
	if err != nil {
		log.Fatalf("failed to initialize Firestore: %v", err)
	}
	defer firestoreClient.Close()

	// gRPC Client
	authTarget := os.Getenv("AUTH_SERVICE_GRPC_TARGET")
	if authTarget == "" {
		authTarget = "auth-service:50051" // Default
	}
	authServiceClient, err := repositories.NewAuthServiceClient(authTarget)
	if err != nil {
		log.Fatalf("failed to initialize auth service client: %v", err)
	}

	// Repositories
	jobOfferRepo := repositories.NewJobOfferRepository(firestoreClient)
	jobApplicationRepo := repositories.NewJobApplicationRepository(firestoreClient)

	// Configuration
	cfg := config.LoadConfig()

	// Mensajería asíncrona (docs/eventos/README.md).
	relay := startRelay(ctx, cfg, firestoreClient)
	var writer *outbox.Writer
	if relay != nil {
		writer = outbox.NewWriter(firestoreClient, events.ProducerInsercionLaboral, relay.Notify)
	}
	insercionLaboralProducer := producers.NewInsercionLaboral(writer, cfg.Events.TopicInsercionLaboral)

	// Services
	jobOfferService := services.NewJobOfferService(
		jobOfferRepo,
		jobApplicationRepo,
		authServiceClient,
		services.WithOfferProducer(insercionLaboralProducer),
	)
	jobApplicationService := services.NewJobApplicationService(
		jobApplicationRepo,
		jobOfferRepo,
		authServiceClient,
		services.WithApplicationProducer(insercionLaboralProducer),
	)

	// Controllers
	jobOfferController := controllers.NewJobOfferController(jobOfferService)
	jobApplicationController := controllers.NewJobApplicationController(jobApplicationService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	routes.RegisterRoutes(r, routes.Controllers{
		JobOffer:       jobOfferController,
		JobApplication: jobApplicationController,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Labor Integration Service listening on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func startRelay(ctx context.Context, cfg *config.Config, client *firestore.Client) *outbox.Relay {
	if !cfg.Events.PublishEnabled() {
		log.Println("PUBSUB_PROJECT_ID vacío: labor-integration no publica eventos")
		return nil
	}
	publisher, err := outbox.NewPubSubPublisher(ctx, cfg.Events.ProjectID)
	if err != nil {
		log.Fatalf("no se pudo inicializar Pub/Sub: %v", err)
	}
	relay := outbox.NewRelay(client, publisher, slog.Default())
	go relay.Run(ctx)
	return relay
}
