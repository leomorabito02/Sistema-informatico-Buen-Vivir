package controllers

import (
	"encoding/json"
	"net/http"

	"buenvivir/labor-integration/internal/dto"
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
	var req dto.CreateJobOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	role := r.Header.Get(headerUserRole)
	userID := r.Header.Get(headerUserID)

	offer := models.JobOffer{
		CreatorName:    req.CreatorName,
		Title:          req.Title,
		Description:    req.Description,
		Location:       req.Location,
		Modality:       req.Modality,
		JobType:        req.JobType,
		Salary:         req.Salary,
		Schedule:       req.Schedule,
		Requirements:   req.Requirements,
		Deadline:       req.Deadline,
	}

	if role == "role_company" {
		offer.CompanyID = userID
	}

	if err := c.service.CreateOffer(r.Context(), &offer, role, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(offer)
}

func (c *JobOfferController) List(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerUserRole)
	companyID := ""
	if role == "role_company" {
		companyID = r.Header.Get(headerUserID)
	}

	offers, err := c.service.ListOffers(r.Context(), companyID, role)
	if err != nil {
		http.Error(w, "failed to list offers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(offers)
}

func (c *JobOfferController) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdateJobOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	offer := models.JobOffer{
		Title:          req.Title,
		Description:    req.Description,
		Location:       req.Location,
		Modality:       req.Modality,
		JobType:        req.JobType,
		Salary:         req.Salary,
		Schedule:       req.Schedule,
		Requirements:   req.Requirements,
		Deadline:       req.Deadline,
		Status:         req.Status,
	}


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

func (c *JobOfferController) Approve(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerUserRole)
	if role != "role_admin" && role != "role_volunteer" {
		http.Error(w, "unauthorized", http.StatusForbidden)
		return
	}
	
	id := chi.URLParam(r, "id")
	userID := r.Header.Get(headerUserID)
	
	if err := c.service.ApproveOffer(r.Context(), id, userID); err != nil {
		http.Error(w, "failed to approve offer", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

func (c *JobOfferController) Reject(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerUserRole)
	if role != "role_admin" && role != "role_volunteer" {
		http.Error(w, "unauthorized", http.StatusForbidden)
		return
	}
	
	id := chi.URLParam(r, "id")
	userID := r.Header.Get(headerUserID)
	
	if err := c.service.RejectOffer(r.Context(), id, userID); err != nil {
		http.Error(w, "failed to reject offer", http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}
