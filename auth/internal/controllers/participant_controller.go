package controllers

import (
	"encoding/json"
	"fmt"
	"io"
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

func (c *ParticipantController) UploadDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
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

	docName := r.FormValue("name")
	if docName == "" {
		docName = header.Filename
	}

	contentType := header.Header.Get(contentTypeHeader)
	doc, err := c.participantService.UploadDocument(r.Context(), id, docName, file, contentType, header.Filename)
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusCreated, true, doc, nil)
}

func (c *ParticipantController) GetDocuments(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	docs, err := c.participantService.GetDocuments(r.Context(), id)
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, docs, nil)
}

func (c *ParticipantController) UpdateDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	docID := chi.URLParam(r, "docId")
	
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.JSONResponse(w, http.StatusBadRequest, false, nil, "Formato inválido")
		return
	}

	docName := r.FormValue("name")
	file, header, err := r.FormFile("file")
	var contentType, filename string
	if err == nil {
		defer file.Close()
		contentType = header.Header.Get(contentTypeHeader)
		filename = header.Filename
	}

	doc, err := c.participantService.UpdateDocument(r.Context(), id, docID, docName, file, contentType, filename)
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, doc, nil)
}

func (c *ParticipantController) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	docID := chi.URLParam(r, "docId")
	
	if err := c.participantService.DeleteDocument(r.Context(), id, docID); err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}

	utils.JSONResponse(w, http.StatusOK, true, map[string]string{"message": "Documento eliminado"}, nil)
}

func (c *ParticipantController) DownloadDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	docID := chi.URLParam(r, "docId")
	
	reader, contentType, filename, err := c.participantService.DownloadDocument(r.Context(), id, docID)
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, false, nil, err.Error())
		return
	}
	defer reader.Close()

	if contentType != "" {
		w.Header().Set(contentTypeHeader, contentType)
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	
	io.Copy(w, reader)
}
