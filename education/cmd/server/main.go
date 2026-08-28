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
	defer firestoreClient.Close()

	courseRepo := repositories.NewCourseRepository(contentfulClient)
	trainingAreaRepo := repositories.NewTrainingAreaRepository(contentfulClient)
	materialRepo := repositories.NewMaterialRepository(contentfulClient)
	activityRepo := repositories.NewActivityRepository(firestoreClient)
	enrollmentRepo := repositories.NewEnrollmentRepository(firestoreClient)
	attendanceRepo := repositories.NewAttendanceRepository(firestoreClient)

	jwtService := services.NewJWTService()
	courseService := services.NewCourseService(courseRepo)
	trainingAreaService := services.NewTrainingAreaService(trainingAreaRepo)
	materialService := services.NewMaterialService(materialRepo)
	activityService := services.NewActivityService(activityRepo)
	enrollmentService := services.NewEnrollmentService(enrollmentRepo)
	attendanceService := services.NewAttendanceService(attendanceRepo)
	progressService := services.NewProgressService(activityRepo, enrollmentRepo, attendanceRepo)

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

	r := chi.NewRouter()
	middlewares.SetupMiddlewares(r)
	r.Use(middlewares.RateLimiter())

	routes.RegisterRoutes(r, jwtService, courseController, trainingAreaController, materialController, healthController, docenteController)

	log.Printf("Servicio Education escuchando en :%s (appEnv=%s, contentfulEnvironment=%s)", cfg.Port, cfg.AppEnv, cfg.ContentfulEnvironment)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("error al iniciar el servidor: %v", err)
	}
}
