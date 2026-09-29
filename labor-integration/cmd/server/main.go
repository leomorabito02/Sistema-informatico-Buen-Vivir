package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"buenvivir/labor-integration/internal/config"
	"buenvivir/labor-integration/internal/controllers"
	"buenvivir/labor-integration/internal/repositories"
	"buenvivir/labor-integration/internal/routes"
	"buenvivir/labor-integration/internal/services"
)

func main() {
	ctx := context.Background()

	firestoreClient, err := config.NewFirestoreClient(ctx)
	if err != nil {
		log.Fatalf("failed to initialize Firestore: %v", err)
	}
	defer firestoreClient.Close()

	// gRPC Client
	authTarget := os.Getenv("AUTH_SERVICE_GRPC_TARGET")
	if authTarget == "" {
		authTarget = "auth-service:50051" // Default
	}
	authServiceClient, err := repositories.NewAuthServiceClient(authTarget)
	if err != nil {
		log.Fatalf("failed to initialize auth service client: %v", err)
	}

	// Repositories
	jobOfferRepo := repositories.NewJobOfferRepository(firestoreClient)
	jobApplicationRepo := repositories.NewJobApplicationRepository(firestoreClient)

	// Services
	jobOfferService := services.NewJobOfferService(jobOfferRepo)
	jobApplicationService := services.NewJobApplicationService(jobApplicationRepo, authServiceClient)

	// Controllers
	jobOfferController := controllers.NewJobOfferController(jobOfferService)
	jobApplicationController := controllers.NewJobApplicationController(jobApplicationService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	routes.RegisterRoutes(r, routes.Controllers{
		JobOffer:       jobOfferController,
		JobApplication: jobApplicationController,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Labor Integration Service listening on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
