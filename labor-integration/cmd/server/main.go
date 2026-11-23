package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

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
	"buenvivir/labor-integration/internal/observability"
	"buenvivir/labor-integration/internal/outbox"
	"buenvivir/labor-integration/internal/repositories"
	"buenvivir/labor-integration/internal/routes"
	"buenvivir/labor-integration/internal/services"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()
	observability.InitLogger("labor-integration", cfg.AppEnv)
	metrics := observability.NewRegistry("labor-integration")
	validateJWTConfig(cfg)

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
	authTarget, authIDToken := resolveServiceEndpoint(ctx, cfg.AuthServiceURL, cfg.AuthServiceGRPCTarget, cfg.UpstreamIDToken)
	if authTarget == "" {
		log.Fatalf("invalid configuration: missing AUTH_SERVICE_URL or AUTH_SERVICE_GRPC_TARGET")
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

	jobOfferRepo := repositories.NewJobOfferRepository(firestoreClient)
	jobApplicationRepo := repositories.NewJobApplicationRepository(firestoreClient)

	// Mensajería asíncrona (docs/eventos/README.md).
	relayCtx, cancelRelay := context.WithCancel(ctx)
	relay := startRelay(relayCtx, cfg, firestoreClient, googleCredentials)
	var writer *outbox.Writer
	if relay != nil {
		writer = outbox.NewWriter(firestoreClient, events.ProducerInsercionLaboral, relay.NotifyInline)
	}
	insercionLaboralProducer := producers.NewInsercionLaboral(writer, cfg.Events.TopicInsercionLaboral)

	// Education Client (HTTP in Cloud Run via EDUCATION_SERVICE_URL, gRPC in local Docker)
	eduTarget, eduIDToken := resolveServiceEndpoint(ctx, cfg.EducationServiceURL, cfg.EducationServiceGRPCTarget, cfg.UpstreamIDToken)

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

	jobOfferController := controllers.NewJobOfferController(jobOfferService).
		WithAudit(repositories.NewAuditRepository(firestoreClient))
	jobApplicationController := controllers.NewJobApplicationController(jobApplicationService)

	r := chi.NewRouter()
	r.Use(observability.HTTPMiddleware(metrics, chiRoutePattern))
	r.Use(middleware.Recoverer)

	routes.RegisterRoutes(r, routes.Controllers{
		JobOffer:       jobOfferController,
		JobApplication: jobApplicationController,
	})
	if observability.MetricsEnabled() {
		r.Handle("/metrics", metrics.Handler())
	}

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}
	slog.Info("Labor Integration Service listening", slog.String("port", cfg.Port))
	if err := observability.ListenAndServeGraceful(ctx, server, func(shutdownCtx context.Context) {
		cancelRelay()
		relay.Shutdown(shutdownCtx)
	}); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func validateJWTConfig(cfg *config.Config) {
	// Sin JWT_SECRET el servicio confía en los headers X-User-* (los pone nginx en local).
	// En producción no es seguro: el API Gateway de GCP no los pisa.
	if cfg.JWTSecret == "" && strings.EqualFold(cfg.AppEnv, "production") {
		log.Fatalf("invalid configuration: JWT_SECRET es obligatorio en producción")
	}
	if cfg.JWTSecret == "" {
		log.Printf("Warning: JWT_SECRET vacío: la identidad sale de los headers X-User-* del gateway (solo válido en local)")
	}
}

func resolveServiceEndpoint(ctx context.Context, httpURL, grpcTarget string, useIDToken bool) (string, func(context.Context) (string, error)) {
	target := httpURL
	if target == "" {
		target = grpcTarget
	}
	if useIDToken && httpURL != "" {
		return target, idTokenPara(ctx, httpURL)
	}
	return target, nil
}

func chiRoutePattern(req *http.Request) string {
	if rc := chi.RouteContext(req.Context()); rc != nil {
		return rc.RoutePattern()
	}
	return ""
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
