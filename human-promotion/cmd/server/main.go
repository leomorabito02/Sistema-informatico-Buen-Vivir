package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
	"google.golang.org/api/option"
	"sync"

	"buenvivir/human-promotion/internal/config"
	"buenvivir/human-promotion/internal/controllers"
	"buenvivir/human-promotion/internal/crypto"
	"buenvivir/human-promotion/internal/events"
	"buenvivir/human-promotion/internal/events/consumers"
	"buenvivir/human-promotion/internal/events/producers"
	"buenvivir/human-promotion/internal/middlewares"
	"buenvivir/human-promotion/internal/outbox"
	"buenvivir/human-promotion/internal/repositories"
	"buenvivir/human-promotion/internal/routes"
	"buenvivir/human-promotion/internal/services"
)

func main() {
	// Logs en JSON a stdout: Cloud Run los manda tal cual a Cloud Logging.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg := config.LoadConfig()
	ctx := context.Background()

	key, err := crypto.LoadKey(ctx, cfg.EncryptionKeySecret, cfg.EncryptionKeyBase64)
	if err != nil {
		fatal("no se pudo obtener la clave de cifrado", err)
	}
	fieldCipher, err := crypto.NewAESGCM(key)
	if err != nil {
		fatal("clave de cifrado inválida", err)
	}

	// Antes de NewFirestoreClient, que limpia la variable de credenciales.
	googleCredentials, err := repositories.GoogleCredentialsJSON()
	if err != nil {
		fatal("credenciales de Google inválidas", err)
	}

	firestoreClient, err := repositories.NewFirestoreClient(ctx)
	if err != nil {
		fatal("no se pudo inicializar Firestore", err)
	}
	defer func() { _ = firestoreClient.Close() }()

	authTarget := cfg.AuthServiceURL
	if authTarget == "" {
		authTarget = cfg.AuthServiceGRPCTarget
	}
	if authTarget == "" {
		fatal("configuración inválida", errors.New("falta AUTH_SERVICE_URL o AUTH_SERVICE_GRPC_TARGET"))
	}
	var authIDToken func(context.Context) (string, error)
	if cfg.UpstreamIDToken && cfg.AuthServiceURL != "" {
		authIDToken = idTokenPara(ctx, cfg.AuthServiceURL)
	}
	authDirectory, err := repositories.NewAuthDirectory(authTarget, authIDToken)
	if err != nil {
		fatal("no se pudo crear el cliente de auth-service", err)
	}
	defer func() { _ = authDirectory.Close() }()

	// Mensajería asíncrona (docs/eventos/README.md).
	relay := startRelay(ctx, cfg, firestoreClient, googleCredentials)
	var writer *outbox.Writer
	if relay != nil {
		writer = outbox.NewWriter(firestoreClient, events.ProducerPromocionHumana, relay.Notify)
	}
	eventProducer := producers.NewPromocionHumana(writer, cfg.Events.TopicPromocion)

	fichaRepo := repositories.NewFichaRepository(firestoreClient, eventProducer)
	hitoRepo := repositories.NewHitoRepository(firestoreClient, eventProducer)

	jwtService := services.NewJWTService()
	eduTarget := cfg.EducationServiceURL
	if eduTarget == "" {
		eduTarget = cfg.EducationServiceGRPCTarget
	}
	var eduIDToken func(context.Context) (string, error)
	if cfg.UpstreamIDToken && cfg.EducationServiceURL != "" {
		eduIDToken = idTokenPara(ctx, cfg.EducationServiceURL)
	}
	fichaService := services.NewFichaService(fichaRepo, fieldCipher, authDirectory, repositories.NewEducationClient(eduTarget, eduIDToken))
	hitoService := services.NewHitoService(hitoRepo, fichaRepo, fieldCipher, authDirectory)
	actividadService := services.NewActividadService(repositories.NewLineaTiempoRepository(firestoreClient), fichaRepo)
	participanteService := services.NewParticipanteService(authDirectory, fichaRepo, hitoRepo, fieldCipher, actividadService)
	pdfService := services.NewPDFService(fichaService, hitoService, actividadService)

	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r, cfg.CORSAllowedOrigins)
	routes.RegisterRoutes(r, jwtService, routes.Controllers{
		Ficha:        controllers.NewFichaController(fichaService, pdfService),
		Hito:         controllers.NewHitoController(hitoService),
		Participante: controllers.NewParticipanteController(participanteService),
		Indicadores:  controllers.NewIndicadoresController(repositories.NewIndicadoresRepository(firestoreClient)),
		Actividad:    controllers.NewActividadController(actividadService),
		Auditoria:    controllers.NewAuditoriaController(services.NewAuditoriaService(repositories.NewAuditoriaRepository(firestoreClient), authDirectory)),
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           routes.WithAsync(r, asyncEndpoints(cfg, firestoreClient, fichaRepo, relay)),
		ReadHeaderTimeout: 10 * time.Second,
	}

	slog.Info("Servicio Human Promotion escuchando", slog.String("port", cfg.Port))
	if err := server.ListenAndServe(); err != nil {
		fatal("error al iniciar el servidor", err)
	}
}

func fatal(msg string, err error) {
	slog.Error(msg, slog.String("error", err.Error()))
	os.Exit(1)
}

// startRelay arranca el publicador del outbox si hay proyecto de Pub/Sub
// configurado; si no, el servicio no publica eventos (devuelve nil).
func startRelay(ctx context.Context, cfg *config.Config, client *firestore.Client, googleCredentials []byte) *outbox.Relay {
	if !cfg.Events.PublishEnabled() {
		slog.Info("PUBSUB_PROJECT_ID vacío: human-promotion no publica eventos")
		return nil
	}
	var opts []option.ClientOption
	if googleCredentials != nil {
		opts = append(opts, option.WithCredentialsJSON(googleCredentials)) //nolint:staticcheck // mismo criterio que NewFirestoreClient (credenciales inline)
	}
	publisher, err := outbox.NewPubSubPublisher(ctx, cfg.Events.ProjectID, opts...)
	if err != nil {
		fatal("no se pudo inicializar Pub/Sub", err)
	}
	relay := outbox.NewRelay(client, publisher, slog.Default())
	go relay.Run(ctx)
	return relay
}

// asyncEndpoints expone los push de Pub/Sub solo si se puede verificar
// quién los manda.
func asyncEndpoints(cfg *config.Config, client *firestore.Client, fichas repositories.FichaRepository, relay *outbox.Relay) routes.AsyncEndpoints {
	var a routes.AsyncEndpoints
	if !cfg.Events.ConsumeEnabled() {
		return a
	}
	verifier, err := events.NewPushVerifier(cfg.Events.PushAuth, cfg.Events.PushAudience, cfg.Events.PushServiceAccounts, cfg.Events.Production)
	if err != nil {
		fatal("configuración de Pub/Sub inválida", err)
	}
	store := outbox.NewFirestoreProcessedStore(client)
	a.Educacion = consumers.NewEducacion(verifier, store, repositories.NewLineaTiempoRepository(client), slog.Default())
	a.Usuarios = consumers.NewUsuarios(verifier, store, fichas, slog.Default())
	if relay != nil {
		a.OutboxFlush = relay.FlushHandler(verifier)
	}
	return a
}

// idTokenPara devuelve una función que da un token de identidad de
// Google para el servicio de esa URL (X-Serverless-Authorization).
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
