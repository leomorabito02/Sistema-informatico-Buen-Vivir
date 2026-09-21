package routes

import (
	"net/http"

	"buenvivir/rxresume/internal/controllers"
)

const errMethodNotAllowed = "method not allowed"

// RegisterRoutes registers all HTTP handler functions for the RxResume API.
func RegisterRoutes(mux *http.ServeMux, ctrl *controllers.ResumeController) {
	mux.HandleFunc("GET /health", ctrl.HealthCheck)
	
	mux.HandleFunc("GET /api/v1/rxresume/me", ctrl.GetMyResume)
	mux.HandleFunc("PUT /api/v1/rxresume/me", ctrl.UpdateMyResume)
	
	mux.HandleFunc("POST /api/v1/rxresume/me/parse-pdf", ctrl.ParseMyPDF)
	mux.HandleFunc("POST /api/v1/rxresume/me/picture", ctrl.UploadMyPicture)
	mux.HandleFunc("POST /api/v1/rxresume/me/generate", ctrl.GenerateMyResume)
	
	mux.HandleFunc("GET /api/v1/rxresume/participants", ctrl.GetAllParticipantsResumes)
	
	mux.HandleFunc("GET /api/v1/rxresume/user", ctrl.GetUserResumeByAdmin)
	mux.HandleFunc("PUT /api/v1/rxresume/user", ctrl.UpdateUserResumeByAdmin)

	mux.HandleFunc("GET /api/v1/rxresume/public/{id}/pdf", ctrl.DownloadPublicPDF)
	mux.HandleFunc("GET /api/v1/rxresume/public/{id}", ctrl.GetPublicResume)
}
