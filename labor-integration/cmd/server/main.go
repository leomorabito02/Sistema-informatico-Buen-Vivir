package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"

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
	cfg := config.LoadConfig()
	// Sin JWT_SECRET el servicio confía en los headers X-User-* (los pone
	// el gateway nginx en local). En producción eso no es seguro: el API
	// Gateway de GCP no los pisa (auditoría 2026-11-11).
	if cfg.JWTSecret == "" && strings.EqualFold(cfg.AppEnv, "production") {
		log.Fatalf("invalid configuration: JWT_SECRET es obligatorio en producción")
	}
	if cfg.JWTSecret == "" {
		log.Printf("Warning: JWT_SECRET vacío: la identidad sale de los headers X-User-* del gateway (solo válido en local)")
	}

	googleCredentials, err := config.GoogleCredentialsJSON()
	if err != nil {
		log.Fatalf("invalid Google credentials: %v", err)
	}

	firestoreClient, err := config.NewFirestoreClient(ctx)
	if err != nil {
		log.Fatalf("failed to initialize Firestore: %v", err)
	}
	defer func() { _ = firestoreClient.Close() }()

	// Auth Client (HTTP in Cloud Run via AUTH_SERVICE_URL, gRPC in local Docker)
	authTarget := cfg.AuthServiceURL
	if authTarget == "" {
		authTarget = cfg.AuthServiceGRPCTarget
	}
	if authTarget == "" {
		log.Fatalf("invalid configuration: missing AUTH_SERVICE_URL or AUTH_SERVICE_GRPC_TARGET")
	}
	var authIDToken func(context.Context) (string, error)
	if cfg.UpstreamIDToken && cfg.AuthServiceURL != "" {
		authIDToken = idTokenPara(ctx, cfg.AuthServiceURL)
	}
	authServiceClient, err := repositories.NewAuthServiceClient(
		authTarget,
		repositories.WithAuthIDToken(authIDToken),
		repositories.WithAuthJWTSecret(cfg.JWTSecret),
	)
	if err != nil {
		log.Fatalf("failed to initialize auth service client: %v", err)
	}
	defer func() { _ = authServiceClient.Close() }()

	// Repositories
	jobOfferRepo := repositories.NewJobOfferRepository(firestoreClient)
	jobApplicationRepo := repositories.NewJobApplicationRepository(firestoreClient)

	// Mensajería asíncrona (docs/eventos/README.md).
	relay := startRelay(ctx, cfg, firestoreClient, googleCredentials)
	var writer *outbox.Writer
	if relay != nil {
		writer = outbox.NewWriter(firestoreClient, events.ProducerInsercionLaboral, relay.Notify)
	}
	insercionLaboralProducer := producers.NewInsercionLaboral(writer, cfg.Events.TopicInsercionLaboral)

	// Education Client (HTTP in Cloud Run via EDUCATION_SERVICE_URL, gRPC in local Docker)
	eduTarget := cfg.EducationServiceURL
	if eduTarget == "" {
		eduTarget = cfg.EducationServiceGRPCTarget
	}
	var eduIDToken func(context.Context) (string, error)
	if cfg.UpstreamIDToken && cfg.EducationServiceURL != "" {
		eduIDToken = idTokenPara(ctx, cfg.EducationServiceURL)
	}

	// Services
	jobOfferService := services.NewJobOfferService(
		jobOfferRepo,
		jobApplicationRepo,
		authServiceClient,
		services.WithOfferProducer(insercionLaboralProducer),
		services.WithSkillsClient(repositories.NewSkillsClient(eduTarget, eduIDToken)),
	)
	jobApplicationService := services.NewJobApplicationService(
		jobApplicationRepo,
		jobOfferRepo,
		authServiceClient,
		services.WithApplicationProducer(insercionLaboralProducer),
	)

	// Controllers
	jobOfferController := controllers.NewJobOfferController(jobOfferService).
		WithAudit(repositories.NewAuditRepository(firestoreClient))
	jobApplicationController := controllers.NewJobApplicationController(jobApplicationService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	routes.RegisterRoutes(r, routes.Controllers{
		JobOffer:       jobOfferController,
		JobApplication: jobApplicationController,
	})

	log.Printf("Labor Integration Service listening on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func startRelay(ctx context.Context, cfg *config.Config, client *firestore.Client, googleCredentials []byte) *outbox.Relay {
	if !cfg.Events.PublishEnabled() {
		log.Println("PUBSUB_PROJECT_ID vacío: labor-integration no publica eventos")
		return nil
	}
	var opts []option.ClientOption
	if googleCredentials != nil {
		opts = append(opts, option.WithCredentialsJSON(googleCredentials)) //nolint:staticcheck // inline credentials injected in Cloud Run CD
	}
	publisher, err := outbox.NewPubSubPublisher(ctx, cfg.Events.ProjectID, opts...)
	if err != nil {
		log.Fatalf("no se pudo inicializar Pub/Sub: %v", err)
	}
	relay := outbox.NewRelay(client, publisher, slog.Default())
	go relay.Run(ctx)
	return relay
}

func idTokenPara(ctx context.Context, audiencia string) func(context.Context) (string, error) {
	var (
		mu sync.Mutex
		ts oauth2.TokenSource
	)
	return func(context.Context) (string, error) {
		mu.Lock()
		defer mu.Unlock()
		if ts == nil {
			s, err := idtoken.NewTokenSource(ctx, audiencia)
			if err != nil {
				return "", err
			}
			ts = s
		}
		tok, err := ts.Token()
		if err != nil {
			return "", err
		}
		return tok.AccessToken, nil
	}
}
