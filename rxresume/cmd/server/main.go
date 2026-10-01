package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"buenvivir/rxresume/internal/client"
	"buenvivir/rxresume/internal/config"
	"buenvivir/rxresume/internal/controllers"
	"buenvivir/rxresume/internal/models"
	"buenvivir/rxresume/internal/parser"
	"buenvivir/rxresume/internal/repositories"
	"buenvivir/rxresume/internal/routes"
	"buenvivir/rxresume/internal/services"
	"buenvivir/rxresume/internal/storage"
	"buenvivir/rxresume/internal/worker"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.LoadConfig()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: Database connection failed (%v). Running with fallback memory/storage.", err)
	} else {
		if err := db.AutoMigrate(&models.Resume{}); err != nil {
			log.Printf("Warning: AutoMigrate failed: %v", err)
		}
	}

	ctx := context.Background()
	storageService, err := storage.NewObjectStorageService(ctx, cfg.GCPProjectID, cfg.GCPBucketName, cfg.GCPEmulatorHost, cfg.ApplicationCreds)
	if err != nil {
		log.Printf("Warning: Storage service initialization warning: %v", err)
	}

	rxClient := client.NewReactiveResumeClient(cfg.ReactiveResumeURL, cfg.FrontendURL, cfg.ReactiveResumeSecretKey)
	pdfParser := parser.NewOfflinePDFParser()
	repo := repositories.NewResumeRepository(db)

	if cfg.PubSubEmulatorHost != "" {
		os.Setenv("PUBSUB_EMULATOR_HOST", cfg.PubSubEmulatorHost)
	}

	syncQueue, err := worker.NewPubSubSyncQueue(ctx, cfg.GCPProjectID, cfg.PubSubSyncTopic, cfg.PubSubSyncSubscription, repo, rxClient, storageService)
	if err != nil {
		log.Fatalf("Warning: Failed to initialize Pub/Sub queue: %v", err)
	}
	syncQueue.Start(ctx)

	service := services.NewResumeService(repo, storageService, pdfParser, rxClient, syncQueue)
	controller := controllers.NewResumeController(service)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, controller)

	log.Printf("Servicio RxResume escuchando en el puerto :%s (conectado a %s)", cfg.Port, cfg.ReactiveResumeURL)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor rxresume: %v", err)
	}
}
