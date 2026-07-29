package routes

import (
	"net/http"

	"buenvivir/education/internal/controllers"
)

func RegisterRoutes(mux *http.ServeMux, ctrl *controllers.EducationController) {
	mux.HandleFunc("/health", ctrl.HealthCheck)
	mux.HandleFunc("/api/v1/education/courses", ctrl.GetCourses)
}
