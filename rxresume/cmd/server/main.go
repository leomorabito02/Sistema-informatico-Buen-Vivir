package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"buenvivir/rxresume/internal/config"
	"buenvivir/rxresume/internal/controllers"
	"buenvivir/rxresume/internal/models"
	"buenvivir/rxresume/internal/renderer"
	"buenvivir/rxresume/internal/repositories"
	"buenvivir/rxresume/internal/routes"
	"buenvivir/rxresume/internal/services"
	"buenvivir/rxresume/internal/storage"

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
	storageService, err := storage.NewObjectStorageService(ctx, cfg.GCPProjectID, cfg.GCPBucketName, cfg.ApplicationCreds)
	if err != nil {
		log.Printf("Warning: Storage service initialization warning: %v", err)
	}

	rendererClient := renderer.NewClient(cfg.RendererURL)
	repo := repositories.NewResumeRepository(db)

	service := services.NewResumeService(repo, storageService, rendererClient)
	controller := controllers.NewResumeController(service)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, controller)

	// Sin JWT_SECRET la identidad sale de los headers X-User-* que pone el
	// gateway nginx (solo válido en Docker local). En producción no arranca.
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		if strings.EqualFold(cfg.AppEnv, "production") {
			log.Fatalf("invalid configuration: JWT_SECRET es obligatorio en producción")
		}
		log.Printf("Warning: JWT_SECRET vacío: la identidad sale de los headers X-User-* del gateway (solo válido en local)")
	}
	handler := routes.Security(mux, cfg.JWTSecret, cfg.CORSAllowedOrigins)

	log.Printf("Servicio RxResume escuchando en el puerto :%s (conectado a cv-renderer en %s)", cfg.Port, cfg.RendererURL)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("Error al iniciar el servidor rxresume: %v", err)
	}
}
