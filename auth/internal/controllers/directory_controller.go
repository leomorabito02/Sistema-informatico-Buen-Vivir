package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"auth/internal/repositories"
	"auth/internal/services"
	"auth/internal/utils"

	"github.com/go-chi/chi/v5"
)

type DirectoryController struct {
	directoryService services.ParticipantDirectoryService
}

func NewDirectoryController(directoryService services.ParticipantDirectoryService) *DirectoryController {
	return &DirectoryController{directoryService: directoryService}
}

type ParticipantProfileDTO struct {
	ParticipantID   string `json:"participant_id"`
	UserID          string `json:"user_id"`
	Nombre          string `json:"nombre"`
	Apellido        string `json:"apellido"`
	FechaNacimiento string `json:"fecha_nacimiento,omitempty"`
	ApprovalStatus  string `json:"approval_status"`
}

func toProfileDTO(row *repositories.ParticipantProfileRow) ParticipantProfileDTO {
	birthdateStr := ""
	if !row.Birthdate.IsZero() {
		birthdateStr = row.Birthdate.Format("2006-01-02")
	}
	return ParticipantProfileDTO{
		ParticipantID:   row.ParticipantID,
		UserID:          row.UserID,
		Nombre:          row.FirstName,
		Apellido:        row.LastName,
		FechaNacimiento: birthdateStr,
		ApprovalStatus:  row.ApprovalStatus,
	}
}

func (c *DirectoryController) GetParticipantProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if strings.TrimSpace(id) == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "ID de participante es requerido")
		return
	}

	profile, err := c.directoryService.GetProfileByParticipantID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repositories.ErrParticipantProfileNotFound) {
			utils.JSONResponse(w, http.StatusNotFound, false, nil, "Participante no encontrado")
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, "Error al consultar participante")
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, toProfileDTO(profile), nil)
}

func (c *DirectoryController) GetProfileByUserID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	if strings.TrimSpace(userID) == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "ID de usuario es requerido")
		return
	}

	profile, err := c.directoryService.GetProfileByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, repositories.ErrParticipantProfileNotFound) {
			utils.JSONResponse(w, http.StatusNotFound, false, nil, "Participante no encontrado")
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, "Error al consultar participante")
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, toProfileDTO(profile), nil)
}

func (c *DirectoryController) SearchParticipants(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	profiles, err := c.directoryService.SearchParticipants(r.Context(), q, limit)
	if err != nil {
		if errors.Is(err, services.ErrInvalidDirectoryRequest) {
			utils.JSONResponse(w, http.StatusBadRequest, false, nil, "Búsqueda inválida (mínimo 2 caracteres)")
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, "Error al buscar participantes")
		return
	}

	dtos := make([]ParticipantProfileDTO, 0, len(profiles))
	for i := range profiles {
		dtos = append(dtos, toProfileDTO(&profiles[i]))
	}

	utils.JSONResponse(w, http.StatusOK, true, dtos, nil)
}

type UserSummariesRequest struct {
	UserIDs []string `json:"user_ids"`
}

type UserSummaryDTO struct {
	UserID   string `json:"user_id"`
	FullName string `json:"full_name"`
}

func (c *DirectoryController) GetUserSummaries(w http.ResponseWriter, r *http.Request) {
	var req UserSummariesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "Cuerpo de solicitud inválido")
		return
	}

	summaries, err := c.directoryService.GetUserSummaries(r.Context(), req.UserIDs)
	if err != nil {
		if errors.Is(err, services.ErrInvalidDirectoryRequest) {
			utils.JSONResponse(w, http.StatusBadRequest, false, nil, "Demasiados IDs o formato inválido")
			return
		}
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, "Error al consultar usuarios")
		return
	}

	dtos := make([]UserSummaryDTO, 0, len(summaries))
	for _, s := range summaries {
		name := strings.TrimSpace(s.FirstName + " " + s.LastName)
		if name == "" {
			name = s.FirstName
		}
		dtos = append(dtos, UserSummaryDTO{
			UserID:   s.UserID,
			FullName: name,
		})
	}

	utils.JSONResponse(w, http.StatusOK, true, dtos, nil)
}
