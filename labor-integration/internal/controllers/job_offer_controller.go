package controllers

import (
	"encoding/json"
	"net/http"

	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/services"

	"github.com/go-chi/chi/v5"
)

const headerUserID = "X-User-ID"

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

	offer.CompanyID = r.Header.Get(headerUserID)

	if err := c.service.CreateOffer(r.Context(), &offer); err != nil {
		http.Error(w, "failed to create offer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(offer)
}

func (c *JobOfferController) List(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get("X-User-Role")
	companyID := ""
	if role == "role_company" {
		companyID = r.Header.Get(headerUserID)
	}

	offers, err := c.service.ListOffers(r.Context(), companyID)
	if err != nil {
		http.Error(w, "failed to list offers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(offers)
}

func (c *JobOfferController) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var offer models.JobOffer
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	offer.CompanyID = r.Header.Get(headerUserID)

	if err := c.service.UpdateOffer(r.Context(), id, &offer); err != nil {
		http.Error(w, "failed to update offer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(offer)
}

func (c *JobOfferController) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := c.service.DeleteOffer(r.Context(), id); err != nil {
		http.Error(w, "failed to delete offer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
