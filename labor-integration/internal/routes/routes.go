package routes

import (
	"github.com/go-chi/chi/v5"

	"buenvivir/labor-integration/internal/controllers"
)

type Controllers struct {
	JobOffer       *controllers.JobOfferController
	JobApplication *controllers.JobApplicationController
}

func RegisterRoutes(r chi.Router, ctrls Controllers) {
	r.Route("/api/v1/labor-integration", func(r chi.Router) {
		r.Get("/offers", ctrls.JobOffer.List)
		r.Post("/offers", ctrls.JobOffer.Create)
		r.Post("/applications", ctrls.JobApplication.Apply)
	})
}
