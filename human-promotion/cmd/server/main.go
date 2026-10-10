package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"

	"buenvivir/human-promotion/internal/config"
	"buenvivir/human-promotion/internal/controllers"
	"buenvivir/human-promotion/internal/crypto"
	"buenvivir/human-promotion/internal/middlewares"
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

	firestoreClient, err := repositories.NewFirestoreClient(ctx)
	if err != nil {
		fatal("no se pudo inicializar Firestore", err)
	}
	defer func() { _ = firestoreClient.Close() }()

	if cfg.AuthServiceGRPCTarget == "" {
		fatal("configuración inválida", errors.New("falta AUTH_SERVICE_GRPC_TARGET"))
	}
	authDirectory, err := repositories.NewAuthDirectory(cfg.AuthServiceGRPCTarget)
	if err != nil {
		fatal("no se pudo crear el cliente de auth-service", err)
	}
	defer func() { _ = authDirectory.Close() }()

	fichaRepo := repositories.NewFichaRepository(firestoreClient)
	hitoRepo := repositories.NewHitoRepository(firestoreClient)

	jwtService := services.NewJWTService()
	fichaService := services.NewFichaService(fichaRepo, fieldCipher, authDirectory, repositories.NewEducationClient(cfg.EducationServiceURL))
	hitoService := services.NewHitoService(hitoRepo, fichaRepo, fieldCipher, authDirectory)
	participanteService := services.NewParticipanteService(authDirectory, fichaRepo, hitoRepo, fieldCipher)
	pdfService := services.NewPDFService(fichaService, hitoService)

	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r, cfg.CORSAllowedOrigins)
	routes.RegisterRoutes(r, jwtService, routes.Controllers{
		Ficha:        controllers.NewFichaController(fichaService, pdfService),
		Hito:         controllers.NewHitoController(hitoService),
		Participante: controllers.NewParticipanteController(participanteService),
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
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
