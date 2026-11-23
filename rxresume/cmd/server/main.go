package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"buenvivir/rxresume/internal/config"
	"buenvivir/rxresume/internal/controllers"
	"buenvivir/rxresume/internal/models"
	"buenvivir/rxresume/internal/observability"
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
	if strings.TrimSpace(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_BASE64")) != "" || strings.HasPrefix(cfg.ApplicationCreds, "{") {
		_ = os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
	}
	observability.InitLogger("rxresume", cfg.AppEnv)
	metrics := observability.NewRegistry("rxresume")

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: Database connection failed (%v). Running with fallback memory/storage.", err)
	} else {
		if err := db.AutoMigrate(&models.Resume{}, &models.AuditEntry{}); err != nil {
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
	var auditRepo repositories.AuditRepository
	if db != nil {
		auditRepo = repositories.NewAuditRepository(db)
	}
	controller := controllers.NewResumeController(service).WithAudit(auditRepo)

	mux := http.NewServeMux()
	if observability.MetricsEnabled() {
		mux.Handle("GET /metrics", metrics.Handler())
	}
	routes.RegisterRoutes(mux, controller)

	// Sin JWT_SECRET la identidad sale de los headers X-User-* que pone el
	// gateway nginx (solo válido en Docker local). En producción no arranca.
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		if strings.EqualFold(cfg.AppEnv, "production") {
			log.Fatalf("invalid configuration: JWT_SECRET es obligatorio en producción")
		}
		slog.Warn("JWT_SECRET vacío: la identidad sale de los headers X-User-* del gateway (solo válido en local)")
	}
	handler := observability.HTTPMiddleware(metrics, nil)(routes.Security(mux, cfg.JWTSecret, cfg.CORSAllowedOrigins))

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
	slog.Info("Servicio RxResume escuchando", slog.String("port", cfg.Port), slog.String("renderer_url", cfg.RendererURL))
	if err := observability.ListenAndServeGraceful(ctx, srv); err != nil {
		log.Fatalf("Error al iniciar el servidor rxresume: %v", err)
	}
}
