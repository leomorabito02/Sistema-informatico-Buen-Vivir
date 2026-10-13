package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"auth/internal/dto"
	"auth/internal/services"
	"auth/internal/utils"

	"github.com/go-chi/chi/v5"
)

const (
	errIDRequired    = "ID es requerido"
	errDocIDRequired = "ID de documento es requerido"
)

type ParticipantController struct {
	participantService services.ParticipantService
}

func NewParticipantController(participantService services.ParticipantService) *ParticipantController {
	return &ParticipantController{participantService: participantService}
}

func validateUploadedFile(file multipart.File, header *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".pdf", ".jpg", ".jpeg", ".png", ".webp":
	default:
		return "", fmt.Errorf("extensión no permitida (%s)", ext)
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("error al leer archivo: %w", err)
	}
	if n == 0 {
		return "", errors.New("archivo vacío")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("error al reposicionar puntero de archivo: %w", err)
	}

	detected := http.DetectContentType(buf[:n])
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(detected, ";")[0]))

	switch mediaType {
	case "application/pdf", "image/jpeg", "image/png", "image/webp":
		return mediaType, nil
	default:
		return "", fmt.Errorf("tipo de contenido MIME no permitido (%s)", mediaType)
	}
}

func mapParticipantError(w http.ResponseWriter, err error, action string) {
	errMsg := err.Error()
	errLower := strings.ToLower(errMsg)

	if strings.Contains(errLower, "forbidden") || strings.Contains(errLower, "does not belong") {
		utils.JSONResponse(w, http.StatusForbidden, false, nil, errMsg)
		return
	}

	if strings.Contains(errLower, "not found") || strings.Contains(errLower, "no encontrado") {
		utils.JSONResponse(w, http.StatusNotFound, false, nil, errMsg)
		return
	}

	slog.Error("Participant operation failed", "action", action, "error", errMsg)
	utils.JSONResponse(w, http.StatusInternalServerError, false, nil, "Error interno al procesar la solicitud")
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
		mapParticipantError(w, err, "CreateParticipant")
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
		mapParticipantError(w, err, "GetParticipant")
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, participant, nil)
}

func (c *ParticipantController) GetAllParticipants(w http.ResponseWriter, r *http.Request) {
	participants, err := c.participantService.GetAllParticipants(r.Context())
	if err != nil {
		mapParticipantError(w, err, "GetAllParticipants")
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
		slog.Warn("Failed to decode update participant request", "participant_id", id, "error", err)
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, fmt.Sprintf("Formato de solicitud inválido: %v", err))
		return
	}

	participant, err := c.participantService.UpdateParticipant(r.Context(), id, &req)
	if err != nil {
		if validationErrors := utils.FormatValidationError(err); len(validationErrors) > 0 {
			slog.Warn("Validation errors on update participant", "participant_id", id, "errors", validationErrors)
			utils.JSONResponse(w, http.StatusBadRequest, false, nil, validationErrors)
			return
		}
		mapParticipantError(w, err, "UpdateParticipant")
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
		mapParticipantError(w, err, "DeleteParticipant")
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, map[string]string{"message": "Participante eliminado correctamente"}, nil)
}

func (c *ParticipantController) UploadDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, errIDRequired)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "Formato inválido")
		return
	}
	
	file, header, err := r.FormFile("file")
	if err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "Archivo es requerido")
		return
	}
	defer file.Close()

	validatedContentType, err := validateUploadedFile(file, header)
	if err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, err.Error())
		return
	}

	docName := r.FormValue("name")
	if docName == "" {
		docName = header.Filename
	}

	doc, err := c.participantService.UploadDocument(r.Context(), id, docName, file, validatedContentType, header.Filename)
	if err != nil {
		mapParticipantError(w, err, "UploadDocument")
		return
	}

	utils.JSONResponse(w, http.StatusCreated, true, doc, nil)
}

func (c *ParticipantController) GetDocuments(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, errIDRequired)
		return
	}

	docs, err := c.participantService.GetDocuments(r.Context(), id)
	if err != nil {
		mapParticipantError(w, err, "GetDocuments")
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, docs, nil)
}

func (c *ParticipantController) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, errIDRequired)
		return
	}
	docID := chi.URLParam(r, "docId")
	if docID == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, errDocIDRequired)
		return
	}
	
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "Formato inválido")
		return
	}

	docName := r.FormValue("name")
	file, header, err := r.FormFile("file")
	var contentType, filename string
	if err == nil {
		defer file.Close()
		validatedContentType, valErr := validateUploadedFile(file, header)
		if valErr != nil {
			utils.JSONResponse(w, http.StatusBadRequest, false, nil, valErr.Error())
			return
		}
		contentType = validatedContentType
		filename = header.Filename
	}

	doc, err := c.participantService.UpdateDocument(r.Context(), id, docID, docName, file, contentType, filename)
	if err != nil {
		mapParticipantError(w, err, "UpdateDocument")
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, doc, nil)
}

func (c *ParticipantController) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, errIDRequired)
		return
	}
	docID := chi.URLParam(r, "docId")
	if docID == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, errDocIDRequired)
		return
	}
	
	if err := c.participantService.DeleteDocument(r.Context(), id, docID); err != nil {
		mapParticipantError(w, err, "DeleteDocument")
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, map[string]string{"message": "Documento eliminado"}, nil)
}

func (c *ParticipantController) DownloadDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, errIDRequired)
		return
	}
	docID := chi.URLParam(r, "docId")
	if docID == "" {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, errDocIDRequired)
		return
	}
	
	reader, contentType, filename, err := c.participantService.DownloadDocument(r.Context(), id, docID)
	if err != nil {
		mapParticipantError(w, err, "DownloadDocument")
		return
	}
	defer reader.Close()

	if contentType != "" {
		w.Header().Set(contentTypeHeader, contentType)
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	
	io.Copy(w, reader)
}
