package main

import (
	"log"
	"net/http"

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

	contentfulClient := repositories.NewContentfulClient(cfg)

	courseRepo := repositories.NewCourseRepository(contentfulClient)
	trainingAreaRepo := repositories.NewTrainingAreaRepository(contentfulClient)
	materialRepo := repositories.NewMaterialRepository(contentfulClient)

	courseService := services.NewCourseService(courseRepo)
	trainingAreaService := services.NewTrainingAreaService(trainingAreaRepo)
	materialService := services.NewMaterialService(materialRepo)

	courseController := controllers.NewCourseController(courseService)
	trainingAreaController := controllers.NewTrainingAreaController(trainingAreaService)
	materialController := controllers.NewMaterialController(materialService)
	healthController := controllers.NewHealthController(cfg)

	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, courseController, trainingAreaController, materialController, healthController)

	handler := middlewares.Chain(mux,
		middlewares.Recoverer,
		middlewares.Logging,
		middlewares.SecurityHeaders,
		middlewares.NewCORS(cfg.CORSAllowedOrigins),
	)

	log.Printf("Servicio Education escuchando en :%s (appEnv=%s, contentfulEnvironment=%s)", cfg.Port, cfg.AppEnv, cfg.ContentfulEnvironment)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("error al iniciar el servidor: %v", err)
	}
}
