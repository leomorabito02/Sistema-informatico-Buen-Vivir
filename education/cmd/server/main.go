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

	firestoreClient, err := repositories.NewFirestoreClient(ctx)
	if err != nil {
		log.Fatalf("no se pudo inicializar Firestore: %v", err)
	}
	defer func() { _ = firestoreClient.Close() }()

	courseRepo := repositories.NewCourseRepository(contentfulClient)
	trainingAreaRepo := repositories.NewTrainingAreaRepository(contentfulClient)
	materialRepo := repositories.NewMaterialRepository(contentfulClient)
	activityRepo := repositories.NewActivityRepository(firestoreClient)
	enrollmentRepo := repositories.NewEnrollmentRepository(firestoreClient)
	attendanceRepo := repositories.NewAttendanceRepository(firestoreClient)
	generalAttendanceRepo := repositories.NewGeneralAttendanceRepository(firestoreClient)
	authServiceClient := repositories.NewAuthServiceClient(cfg.AuthServiceURL)

	jwtService := services.NewJWTService()
	kvSyncService := services.NewKVSyncService(courseRepo)
	kvSyncService.SyncCourseStatsAsync()

	courseService := services.NewCourseService(courseRepo, kvSyncService)
	trainingAreaService := services.NewTrainingAreaService(trainingAreaRepo)
	materialService := services.NewMaterialService(materialRepo)
	activityService := services.NewActivityService(activityRepo)
	enrollmentService := services.NewEnrollmentService(enrollmentRepo, courseRepo, authServiceClient)
	attendanceService := services.NewAttendanceService(attendanceRepo, enrollmentRepo, authServiceClient)
	progressService := services.NewProgressService(activityRepo, enrollmentRepo, attendanceRepo)
	generalAttendanceService := services.NewGeneralAttendanceService(generalAttendanceRepo, authServiceClient)

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
	)
	enrollmentController := controllers.NewEnrollmentController(enrollmentService)
	attendanceController := controllers.NewAttendanceController(activityService, attendanceService)
	generalAttendanceController := controllers.NewGeneralAttendanceController(generalAttendanceService)

	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r)
	r.Use(middlewares.RateLimiter())

	routes.RegisterRoutes(
		r, jwtService,
		courseController, trainingAreaController, materialController, healthController, docenteController,
		enrollmentController, attendanceController, generalAttendanceController,
	)

	log.Printf("Servicio Education escuchando en :%s (appEnv=%s, contentfulEnvironment=%s)", cfg.Port, cfg.AppEnv, cfg.ContentfulEnvironment)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("error al iniciar el servidor: %v", err)
	}
}
