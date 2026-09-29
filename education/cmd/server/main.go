package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/config"
	"buenvivir/education/internal/controllers"
	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/repositories"
	"buenvivir/education/internal/routes"
	"buenvivir/education/internal/services"
)

func main() {
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

	courseRepo := repositories.NewCourseRepository(contentfulClient)
	trainingAreaRepo := repositories.NewTrainingAreaRepository(contentfulClient)
	materialRepo := repositories.NewMaterialRepository(contentfulClient)
	activityRepo := repositories.NewActivityRepository(firestoreClient)
	enrollmentRepo := repositories.NewEnrollmentRepository(firestoreClient)
	attendanceRepo := repositories.NewAttendanceRepository(firestoreClient)
	submissionRepo := repositories.NewSubmissionRepository(firestoreClient)
	generalAttendanceRepo := repositories.NewGeneralAttendanceRepository(firestoreClient)
	authServiceClient := repositories.NewAuthServiceClient(cfg.AuthServiceURL)

	jwtService := services.NewJWTService()
	kvSyncService := services.NewKVSyncService(courseRepo)
	kvSyncService.SyncCourseStatsAsync()

	courseService := services.NewCourseService(courseRepo, kvSyncService)
	trainingAreaService := services.NewTrainingAreaService(trainingAreaRepo)
	materialService := services.NewMaterialService(materialRepo)
	activityService := services.NewActivityService(activityRepo, fileStorage)
	enrollmentService := services.NewEnrollmentService(enrollmentRepo, courseRepo, authServiceClient)
	attendanceService := services.NewAttendanceService(attendanceRepo)
	progressService := services.NewProgressService(activityRepo, enrollmentRepo, attendanceRepo)
	generalAttendanceService := services.NewGeneralAttendanceService(generalAttendanceRepo, authServiceClient)
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

	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r)
	r.Use(middlewares.RateLimiter())

	routes.RegisterRoutes(
		r, jwtService,
		courseController, trainingAreaController, materialController, healthController, docenteController,
		enrollmentController, attendanceController, generalAttendanceController, activityController,
		submissionController,
	)

	log.Printf("Servicio Education escuchando en :%s (appEnv=%s, contentfulEnvironment=%s)", cfg.Port, cfg.AppEnv, cfg.ContentfulEnvironment)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("error al iniciar el servidor: %v", err)
	}
}
