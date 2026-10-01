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
		r.Put("/offers/{id}", ctrls.JobOffer.Update)
		r.Delete("/offers/{id}", ctrls.JobOffer.Delete)
		r.Get("/offers/{id}/applications", ctrls.JobApplication.ListByOffer)
		r.Get("/applications/me", ctrls.JobApplication.ListMine)
		r.Post("/applications", ctrls.JobApplication.Apply)
		r.Put("/applications/{id}/status", ctrls.JobApplication.UpdateStatus)
	})
}
