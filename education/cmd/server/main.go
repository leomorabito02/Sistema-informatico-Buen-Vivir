package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/go-chi/chi/v5"
	"google.golang.org/api/option"

	"buenvivir/education/internal/config"
	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/events"
	"buenvivir/education/internal/events/consumers"
	"buenvivir/education/internal/events/producers"
	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/outbox"
	"buenvivir/education/internal/repositories"
	"buenvivir/education/internal/routes"
	"buenvivir/education/internal/services"
)

func main() {
	// Logs estructurados en JSON (Cloud Run los manda a Cloud Logging).
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("configuración inválida: %v", err)
	}

	ctx := context.Background()

	contentfulClient := repositories.NewContentfulClient(cfg)

	// Se resuelven antes de NewFirestoreClient, que limpia la variable
	// de entorno de credenciales inline (ver GoogleCredentialsJSON).
	googleCredentials, err := repositories.GoogleCredentialsJSON()
	if err != nil {
		log.Fatalf("credenciales de Google inválidas: %v", err)
	}

	firestoreClient, err := repositories.NewFirestoreClient(ctx)
	if err != nil {
		log.Fatalf("no se pudo inicializar Firestore: %v", err)
	}
	defer func() { _ = firestoreClient.Close() }()

	fileStorage, storageClient, err := repositories.NewGCSFileStorage(ctx, cfg.EntregasBucket, googleCredentials)
	if err != nil {
		log.Fatalf("no se pudo inicializar el bucket de entregas: %v", err)
	}
	if storageClient != nil {
		defer func() { _ = storageClient.Close() }()
	} else {
		log.Printf("ENTREGAS_BUCKET vacío: las entregas se aceptan solo con texto, sin adjuntos")
	}

	// Mensajería asíncrona (docs/eventos/README.md): el productor escribe
	// en el outbox y el relay publica en Pub/Sub.
	relay := startRelay(ctx, cfg, firestoreClient, googleCredentials)
	var writer *outbox.Writer
	if relay != nil {
		writer = outbox.NewWriter(firestoreClient, events.ProducerEducacion, relay.Notify)
	}
	eventProducer := producers.NewEducacion(writer, cfg.Events.TopicEducacion, cfg.Events.TopicContenido)

	courseRepo := repositories.NewCourseRepository(contentfulClient)
	trainingAreaRepo := repositories.NewTrainingAreaRepository(contentfulClient)
	materialRepo := repositories.NewMaterialRepository(contentfulClient)
	activityRepo := repositories.NewActivityRepository(firestoreClient)
	enrollmentRepo := repositories.NewEnrollmentRepository(firestoreClient, eventProducer)
	attendanceRepo := repositories.NewAttendanceRepository(firestoreClient)
	submissionRepo := repositories.NewSubmissionRepository(firestoreClient, eventProducer)
	generalAttendanceRepo := repositories.NewGeneralAttendanceRepository(firestoreClient, eventProducer)
	skillRepo := repositories.NewSkillRepository(firestoreClient, eventProducer)
	authServiceClient := repositories.NewAuthServiceClient(cfg.AuthServiceURL)

	jwtService := services.NewJWTService()
	kvSyncService := services.NewKVSyncService(courseRepo)
	kvSyncService.SyncCourseStatsAsync()

	courseService := services.NewCourseServiceWithEvents(courseRepo, eventProducer, kvSyncService)
	trainingAreaService := services.NewTrainingAreaService(trainingAreaRepo)
	materialService := services.NewMaterialService(materialRepo)
	activityService := services.NewActivityService(activityRepo, fileStorage)
	enrollmentService := services.NewEnrollmentService(enrollmentRepo, courseRepo, authServiceClient)
	attendanceService := services.NewAttendanceService(attendanceRepo)
	progressService := services.NewProgressService(activityRepo, enrollmentRepo, attendanceRepo)
	generalAttendanceService := services.NewGeneralAttendanceService(generalAttendanceRepo, authServiceClient)
	skillService := services.NewSkillService(skillRepo, submissionRepo)
	skillSuggestionService := services.NewSkillSuggestionService(skillRepo, skillProviders(cfg, googleCredentials), cfg.SkillsAI.Timeout, cfg.SkillsAI.MaxRetries)
	submissionService := services.NewSubmissionService(submissionRepo, activityRepo, enrollmentRepo, courseRepo, fileStorage, authServiceClient)

	courseController := controllers.NewCourseController(courseService)
	trainingAreaController := controllers.NewTrainingAreaController(trainingAreaService)
	materialController := controllers.NewMaterialController(materialService)
	healthController := controllers.NewHealthController(cfg)
	docenteController := controllers.NewDocenteController(
		courseService,
		activityService,
		enrollmentService,
		attendanceService,
		progressService,
		submissionService,
	)
	enrollmentController := controllers.NewEnrollmentController(enrollmentService)
	attendanceController := controllers.NewAttendanceController(activityService, attendanceService)
	generalAttendanceController := controllers.NewGeneralAttendanceController(generalAttendanceService)
	activityController := controllers.NewActivityController(courseService, activityService, submissionService)
	submissionController := controllers.NewSubmissionController(courseService, activityService, submissionService)
	skillController := controllers.NewSkillController(courseService, activityService, skillService, skillSuggestionService, submissionService)

	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r)
	r.Use(middlewares.RateLimiter())

	routes.RegisterRoutes(
		r, jwtService,
		routes.Controllers{
			Course:            courseController,
			TrainingArea:      trainingAreaController,
			Material:          materialController,
			Health:            healthController,
			Docente:           docenteController,
			Enrollment:        enrollmentController,
			Attendance:        attendanceController,
			GeneralAttendance: generalAttendanceController,
			Activity:          activityController,
			Submission:        submissionController,
			Skill:             skillController,
		},
	)

	handler := routes.WithAsync(r, asyncEndpoints(cfg, firestoreClient, enrollmentRepo, contentfulClient, eventProducer, relay))

	log.Printf("Servicio Education escuchando en :%s (appEnv=%s, contentfulEnvironment=%s)", cfg.Port, cfg.AppEnv, cfg.ContentfulEnvironment)
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("error al iniciar el servidor: %v", err)
	}
}

// skillProviders arma la lista de proveedores de IA para sugerir
// habilidades, en orden de preferencia: Gemini (principal) y OpenAI
// (respaldo). Los que no están configurados se omiten; si falla la
// configuración de Gemini el servicio arranca igual (con OpenAI, o sin
// IA), porque sugerir habilidades es una ayuda, no algo imprescindible.
func skillProviders(cfg *config.Config, googleCredentials []byte) []services.SugeridorHabilidades {
	var providers []services.SugeridorHabilidades
	ai := cfg.SkillsAI
	if ai.GeminiEnabled() {
		gemini, err := repositories.NewGeminiClient(ai.VertexProjectID, ai.VertexLocation, ai.GeminiModel, googleCredentials)
		if err != nil {
			log.Printf("Gemini deshabilitado: %v", err)
		} else {
			providers = append(providers, gemini)
		}
	}
	if ai.OpenAIEnabled() {
		providers = append(providers, repositories.NewOpenAIClient(ai.OpenAIAPIKey, ai.OpenAIModel))
	}
	names := make([]string, len(providers))
	for i, p := range providers {
		names[i] = p.Nombre()
	}
	log.Printf("sugerencia de habilidades con IA: proveedores %v", names)
	return providers
}

// startRelay arranca el publicador del outbox si hay proyecto de Pub/Sub
// configurado (PUBSUB_PROJECT_ID). Sin él, el servicio no publica
// eventos y devuelve nil.
func startRelay(ctx context.Context, cfg *config.Config, client *firestore.Client, googleCredentials []byte) *outbox.Relay {
	if !cfg.Events.PublishEnabled() {
		log.Printf("PUBSUB_PROJECT_ID vacío: education no publica eventos")
		return nil
	}
	var opts []option.ClientOption
	if googleCredentials != nil {
		opts = append(opts, option.WithCredentialsJSON(googleCredentials)) //nolint:staticcheck // mismo criterio que NewFirestoreClient (credenciales inline)
	}
	publisher, err := outbox.NewPubSubPublisher(ctx, cfg.Events.ProjectID, opts...)
	if err != nil {
		log.Fatalf("no se pudo inicializar Pub/Sub: %v", err)
	}
	relay := outbox.NewRelay(client, publisher, slog.Default())
	go relay.Run(ctx)
	return relay
}

// asyncEndpoints arma los endpoints de eventos según la configuración:
// los push de Pub/Sub solo si se puede verificar quién los manda, y el
// webhook de Contentful solo si tiene secreto y hay dónde publicar.
func asyncEndpoints(cfg *config.Config, client *firestore.Client, enrollments repositories.EnrollmentRepository, contentful *repositories.ContentfulClient, producer *producers.Educacion, relay *outbox.Relay) routes.AsyncEndpoints {
	var a routes.AsyncEndpoints
	if cfg.Events.ContentfulWebhookSecret != "" && producer.Enabled() {
		a.ContentfulWebhook = http.HandlerFunc(controllers.NewContentfulWebhookController(cfg.Events.ContentfulWebhookSecret, producer).Receive)
	}
	if !cfg.Events.ConsumeEnabled() {
		return a
	}
	verifier, err := events.NewPushVerifier(cfg.Events.PushAuth, cfg.Events.PushAudience, cfg.Events.PushServiceAccounts, cfg.AppEnv == config.AppEnvProduction)
	if err != nil {
		log.Fatalf("configuración de Pub/Sub inválida: %v", err)
	}
	store := outbox.NewFirestoreProcessedStore(client)
	a.Usuarios = consumers.NewUsuarios(verifier, store, enrollments, slog.Default())
	a.Contenido = consumers.NewContenido(verifier, store, contentful, slog.Default())
	if relay != nil {
		a.OutboxFlush = relay.FlushHandler(verifier)
	}
	return a
}
