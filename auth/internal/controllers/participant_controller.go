package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"auth/internal/dto"
	"auth/internal/services"
	"auth/internal/utils"

	"github.com/go-chi/chi/v5"
)

const errIDRequired = "ID es requerido"

type ParticipantController struct {
	participantService services.ParticipantService
}

func NewParticipantController(participantService services.ParticipantService) *ParticipantController {
	return &ParticipantController{participantService: participantService}
}

func (c *ParticipantController) CreateParticipant(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, err.Error())
		return
	}

	participant, err := c.participantService.CreateParticipant(r.Context(), &req)
	if err != nil {
		if validationErrors := utils.FormatValidationError(err); len(validationErrors) > 0 {
			utils.JSONResponse(w, http.StatusBadRequest, false, nil, validationErrors)
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusCreated, true, participant, nil)
}

func (c *ParticipantController) GetParticipant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, errIDRequired)
		return
	}

	participant, err := c.participantService.GetParticipantByID(r.Context(), id)
	if err != nil {
		utils.JSONResponse(w, http.StatusNotFound, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, participant, nil)
}

func (c *ParticipantController) GetAllParticipants(w http.ResponseWriter, r *http.Request) {
	participants, err := c.participantService.GetAllParticipants(r.Context())
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, participants, nil)
}

func (c *ParticipantController) UpdateParticipant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, errIDRequired)
		return
	}

	var req dto.UpdateParticipantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("UpdateParticipant Decode Error: %v\n", err)
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, fmt.Sprintf("Formato de solicitud inválido: %v", err))
		return
	}

	fmt.Printf("UpdateParticipant Decoded Request: %+v\n", req)
	participant, err := c.participantService.UpdateParticipant(r.Context(), id, &req)
	if err != nil {
		if validationErrors := utils.FormatValidationError(err); len(validationErrors) > 0 {
			fmt.Printf("UpdateParticipant Validation Errors: %v\n", validationErrors)
			utils.JSONResponse(w, http.StatusBadRequest, false, nil, validationErrors)
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, participant, nil)
}

func (c *ParticipantController) DeleteParticipant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, errIDRequired)
		return
	}

	if err := c.participantService.DeleteParticipant(r.Context(), id); err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, map[string]string{"message": "Participante eliminado correctamente"}, nil)
}
