package controllers

import (
	"encoding/json"
	"net/http"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/services"
)

type JobOfferController struct {
	service *services.JobOfferService
}

func NewJobOfferController(service *services.JobOfferService) *JobOfferController {
	return &JobOfferController{service: service}
}

func (c *JobOfferController) Create(w http.ResponseWriter, r *http.Request) {
	var offer models.JobOffer
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if err := c.service.CreateOffer(r.Context(), &offer); err != nil {
		http.Error(w, "failed to create offer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(offer)
}

func (c *JobOfferController) List(w http.ResponseWriter, r *http.Request) {
	offers, err := c.service.ListOffers(r.Context())
	if err != nil {
		http.Error(w, "failed to list offers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(offers)
}
