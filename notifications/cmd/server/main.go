package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"buenvivir/notifications/api/proto/notificationpb"
	"buenvivir/notifications/internal/audience"
	"buenvivir/notifications/internal/clients"
	"buenvivir/notifications/internal/config"
	"buenvivir/notifications/internal/consumers"
	"buenvivir/notifications/internal/controllers"
	"buenvivir/notifications/internal/database"
	"buenvivir/notifications/internal/delivery"
	"buenvivir/notifications/internal/events"
	"buenvivir/notifications/internal/repositories"
	"buenvivir/notifications/internal/routes"
	"buenvivir/notifications/internal/rules"
	"buenvivir/notifications/internal/services"
	"buenvivir/notifications/internal/store"
	"buenvivir/notifications/internal/templates"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	db := setupDatabase(cfg.DatabaseURL)
	repo := repositories.NewNotificationRepository(db)

	var serviceOpts []services.NotificationServiceOption
	var emailSvc services.EmailService
	var pushSvc services.PushService
	var authClient clients.AuthClient
	var tmplEngine templates.TemplateEngine

	// 1. Resend Email Client
	if cfg.ResendAPIKey != "" {
		svc, err := services.NewDualResendEmailService(
			cfg.ResendAPIKey,
			cfg.ResendFromNotifications,
			cfg.ResendFromAlerts,
		)
		if err != nil {
			log.Printf("Warning: failed to initialize Resend email service: %v", err)
		} else {
			log.Println("Resend email service initialized successfully (dual-from support enabled).")
			emailSvc = svc
			serviceOpts = append(serviceOpts, services.WithEmailService(svc))
		}
	} else {
		log.Println("Notice: RESEND_API_KEY is not configured; email dispatching will be disabled.")
	}

	// 2. Firebase FCM Push Client via Google Cloud ADC
	fcmSvc, err := services.NewFirebasePushService(context.Background())
	if err != nil {
		log.Printf("Notice: Firebase Push Service not active (will skip push until ADC is available): %v", err)
	} else {
		log.Println("Firebase Cloud Messaging (FCM) push service initialized via Google Cloud ADC.")
		pushSvc = fcmSvc
		serviceOpts = append(serviceOpts, services.WithPushService(fcmSvc))
	}

	// 3. Auth gRPC Client for Device Tokens
	authCl, err := clients.NewAuthClient(cfg.AuthGRPCURL)
	if err != nil {
		log.Printf("Warning: failed to initialize Auth gRPC client for %s: %v", cfg.AuthGRPCURL, err)
	} else {
		log.Printf("Auth gRPC client connected to %s", cfg.AuthGRPCURL)
		authClient = authCl
		serviceOpts = append(serviceOpts, services.WithAuthClient(authCl))
	}

	// 4. HTML Template Engine (//go:embed)
	tmplEng, err := templates.NewTemplateEngine()
	if err != nil {
		log.Printf("Warning: failed to initialize embedded template engine: %v", err)
	} else {
		log.Println("Embedded HTML template engine initialized successfully.")
		tmplEngine = tmplEng
		serviceOpts = append(serviceOpts, services.WithTemplateEngine(tmplEng))
	}

	service := services.NewNotificationService(repo, serviceOpts...)
	controller := controllers.NewNotificationController(service)
	grpcController := controllers.NewGrpcNotificationController(service)

	startGRPCServer(cfg.GRPCPort, grpcController)

	// Delivery Worker
	deliveryWorker := delivery.NewWorker(repo, emailSvc, pushSvc, authClient, tmplEngine, nil)
	deliveryWorker.FrontendURL = cfg.FrontendURL
	go deliveryWorker.Run(context.Background())

	// Pub/Sub Push Verifier and Processed Store
	isProd := cfg.Environment == "production"
	pushVerifier, err := events.NewPushVerifier(cfg.PubSubPushAuth, cfg.PubSubPushAudience, cfg.PubSubPushServiceAccounts, isProd)
	if err != nil {
		log.Printf("Warning: push verifier initialization: %v", err)
		pushVerifier = events.NoVerifier{}
	}
	processedStore := store.NewProcessedStore(db)

	// Rule Registry and Audience Resolver
	registry := rules.NewRegistry()
	registry.Register(&rules.InscripcionCreadaRule{})
	registry.Register(&rules.ContenidoDocenteSubidoRule{})
	registry.Register(&rules.ContenidoDocenteModeradoRule{})
	registry.Register(&rules.CursoCompletadoRule{})
	registry.Register(&rules.HabilidadOtorgadaRule{})
	registry.Register(&rules.CursoEliminadoRule{})
	registry.Register(&rules.CursoCreadoRule{})
	registry.Register(&rules.ActividadRecordatorioRule{})
	registry.Register(&rules.FichaCreadaRule{})
	registry.Register(&rules.FichaActualizadaRule{})
	registry.Register(&rules.HitoRegistradoRule{})
	registry.Register(&rules.HitoActualizadoRule{})
	registry.Register(&rules.ParticipanteEliminadoRule{})
	registry.Register(&rules.OfertaCreadaRule{})
	registry.Register(&rules.OfertaRevisadaRule{})
	registry.Register(&rules.PostulacionCreadaRule{})
	registry.Register(&rules.PostulacionDecididaRule{})
	registry.Register(&rules.CampanaCreadaRule{})
	registry.Register(&rules.UsuarioCreadoRule{})
	registry.Register(&rules.EmailSolicitadoRule{})

	audienceResolver := audience.NewResolver(authClient)

	eduConsumer := consumers.NewEducacionConsumer(pushVerifier, processedStore, registry, audienceResolver, repo, deliveryWorker, nil)
	phConsumer := consumers.NewPromocionHumanaConsumer(pushVerifier, processedStore, registry, audienceResolver, repo, deliveryWorker, nil)
	usuariosConsumer := consumers.NewUsuariosConsumer(pushVerifier, processedStore, registry, audienceResolver, repo, deliveryWorker, nil)
	laborConsumer := consumers.NewInsercionLaboralConsumer(pushVerifier, processedStore, registry, audienceResolver, repo, deliveryWorker, nil)
	campanasConsumer := consumers.NewCampanasConsumer(pushVerifier, processedStore, registry, audienceResolver, repo, deliveryWorker, nil)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, controller)
	routes.RegisterAsyncRoutes(mux, routes.AsyncEndpoints{
		Educacion:        eduConsumer,
		PromocionHumana:  phConsumer,
		Usuarios:         usuariosConsumer,
		InsercionLaboral: laborConsumer,
		Campanas:         campanasConsumer,
		DeliveryFlush:    deliveryWorker.FlushHandler(pushVerifier),
	})

	log.Printf("Servicio Notifications escuchando en :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}

func startGRPCServer(port string, grpcController *controllers.GrpcNotificationController) {
	go func() {
		lis, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Printf("Error: failed to listen on gRPC port %s: %v", port, err)
			return
		}
		grpcServer := grpc.NewServer()
		notificationpb.RegisterNotificationIngestionServiceServer(grpcServer, grpcController)
		log.Printf("Servicio Notifications (gRPC) escuchando en :%s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Error serving gRPC: %v", err)
		}
	}()
}

func setupDatabase(dbURL string) *gorm.DB {
	if dbURL == "" {
		log.Println("Notice: DATABASE_URL_NOTIFICATIONS is not set. Running in-memory repository mode.")
		return nil
	}

	db, err := database.InitDatabase(dbURL)
	if err != nil {
		log.Printf("Warning: failed to connect to database: %v. Running in-memory repository mode.", err)
		return nil
	}

	log.Println("Database connection pool initialized successfully.")
	return db
}
