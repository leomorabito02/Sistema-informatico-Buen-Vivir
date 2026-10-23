package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"buenvivir/notifications/api/proto/notificationpb"
	"buenvivir/notifications/internal/clients"
	"buenvivir/notifications/internal/config"
	"buenvivir/notifications/internal/controllers"
	"buenvivir/notifications/internal/database"
	"buenvivir/notifications/internal/repositories"
	"buenvivir/notifications/internal/routes"
	"buenvivir/notifications/internal/services"
	"buenvivir/notifications/internal/templates"

	"google.golang.org/grpc"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	db := setupDatabase(cfg.DatabaseURL)
	repo := repositories.NewNotificationRepository(db)

	var serviceOpts []services.NotificationServiceOption

	// 1. Resend Email Client
	if cfg.ResendAPIKey != "" {
		emailSvc, err := services.NewDualResendEmailService(
			cfg.ResendAPIKey,
			cfg.ResendFromNotifications,
			cfg.ResendFromAlerts,
		)
		if err != nil {
			log.Printf("Warning: failed to initialize Resend email service: %v", err)
		} else {
			log.Println("Resend email service initialized successfully (dual-from support enabled).")
			serviceOpts = append(serviceOpts, services.WithEmailService(emailSvc))
		}
	} else {
		log.Println("Notice: RESEND_API_KEY is not configured; email dispatching will be disabled.")
	}

	// 2. Firebase FCM Push Client via Google Cloud ADC
	pushSvc, err := services.NewFirebasePushService(context.Background())
	if err != nil {
		log.Printf("Notice: Firebase Push Service not active (will skip push until ADC is available): %v", err)
	} else {
		log.Println("Firebase Cloud Messaging (FCM) push service initialized via Google Cloud ADC.")
		serviceOpts = append(serviceOpts, services.WithPushService(pushSvc))
	}

	// 3. Auth gRPC Client for Device Tokens
	authClient, err := clients.NewAuthClient(cfg.AuthGRPCURL)
	if err != nil {
		log.Printf("Warning: failed to initialize Auth gRPC client for %s: %v", cfg.AuthGRPCURL, err)
	} else {
		log.Printf("Auth gRPC client connected to %s", cfg.AuthGRPCURL)
		serviceOpts = append(serviceOpts, services.WithAuthClient(authClient))
	}

	// 4. HTML Template Engine (//go:embed)
	tmplEngine, err := templates.NewTemplateEngine()
	if err != nil {
		log.Printf("Warning: failed to initialize embedded template engine: %v", err)
	} else {
		log.Println("Embedded HTML template engine initialized successfully.")
		serviceOpts = append(serviceOpts, services.WithTemplateEngine(tmplEngine))
	}

	service := services.NewNotificationService(repo, serviceOpts...)
	controller := controllers.NewNotificationController(service)
	grpcController := controllers.NewGrpcNotificationController(service)

	startGRPCServer(cfg.GRPCPort, grpcController)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, controller)

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
