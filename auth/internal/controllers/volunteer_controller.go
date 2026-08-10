package controllers

import (
	"encoding/json"
	"net/http"

	"auth/internal/dto"
	"auth/internal/services"
	"auth/internal/utils"

	"github.com/go-chi/chi/v5"
)

type VolunteerController struct {
	volunteerService services.VolunteerService
}

func NewVolunteerController(volunteerService services.VolunteerService) *VolunteerController {
	return &VolunteerController{volunteerService: volunteerService}
}

func (c *VolunteerController) CreateVolunteer(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateVolunteerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "Formato de solicitud inválido")
		return
	}

	volunteer, err := c.volunteerService.CreateVolunteer(r.Context(), &req)
	if err != nil {
		if validationErrors := utils.FormatValidationError(err); len(validationErrors) > 0 {
			utils.JSONResponse(w, http.StatusBadRequest, false, nil, validationErrors)
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusCreated, true, volunteer, nil)
}

func (c *VolunteerController) GetVolunteer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "ID es requerido")
		return
	}

	volunteer, err := c.volunteerService.GetVolunteerByID(r.Context(), id)
	if err != nil {
		utils.JSONResponse(w, http.StatusNotFound, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, volunteer, nil)
}

func (c *VolunteerController) GetAllVolunteers(w http.ResponseWriter, r *http.Request) {
	volunteers, err := c.volunteerService.GetAllVolunteers(r.Context())
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, volunteers, nil)
}

func (c *VolunteerController) UpdateVolunteer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "ID es requerido")
		return
	}

	var req dto.UpdateVolunteerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "Formato de solicitud inválido")
		return
	}

	volunteer, err := c.volunteerService.UpdateVolunteer(r.Context(), id, &req)
	if err != nil {
		if validationErrors := utils.FormatValidationError(err); len(validationErrors) > 0 {
			utils.JSONResponse(w, http.StatusBadRequest, false, nil, validationErrors)
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, volunteer, nil)
}

func (c *VolunteerController) DeleteVolunteer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "ID es requerido")
		return
	}

	if err := c.volunteerService.DeleteVolunteer(r.Context(), id); err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, map[string]string{"message": "Voluntario eliminado correctamente"}, nil)
}
