package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"buenvivir/rxresume/internal/dto"
	"buenvivir/rxresume/internal/services"
)

const (
	headerXUserID     = "X-User-ID"
	headerXUserRole   = "X-User-Role"
	headerContentType     = "Content-Type"
	contentTypeJSON       = "application/json"
	errMissingUserID      = "missing user identity header X-User-ID"
	errInvalidJSON        = "invalid JSON payload"
	errParseMultipartForm = "failed to parse multipart form"
	errFileFieldRequired  = "file field is required"
	errInsufficientPerms  = "insufficient permissions"
	errMissingParamUserID = "parameter user_id is required"
)

type ResumeController struct {
	service services.ResumeService
}

func NewResumeController(service services.ResumeService) *ResumeController {
	return &ResumeController{service: service}
}

// GetMyResume handles fetching the CV for the currently authenticated participant.
func (c *ResumeController) GetMyResume(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(headerXUserID)
	if userID == "" {
		http.Error(w, errMissingUserID, http.StatusUnauthorized)
		return
	}

	resume, err := c.service.GetResumeByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(resume)
}

// GetMyPDFStatus returns whether the PDF is ready in GCS for the logged-in participant.
func (c *ResumeController) GetMyPDFStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(headerXUserID)
	if userID == "" {
		http.Error(w, errMissingUserID, http.StatusUnauthorized)
		return
	}
	ready := c.service.GetPDFStatus(r.Context(), userID)
	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(map[string]bool{"ready": ready})
}

// UpdateMyResume updates the resume content of the logged-in participant.
func (c *ResumeController) UpdateMyResume(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(headerXUserID)
	if userID == "" {
		http.Error(w, errMissingUserID, http.StatusUnauthorized)
		return
	}

	var req dto.UpdateResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, errInvalidJSON, http.StatusBadRequest)
		return
	}
	req.Sanitize()
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resume, err := c.service.UpdateResume(r.Context(), userID, &req.ResumeData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(resume)
}

// ParseMyPDF handles uploading and parsing a PDF resume for the current participant.
func (c *ResumeController) ParseMyPDF(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(headerXUserID)
	if userID == "" {
		http.Error(w, errMissingUserID, http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, errParseMultipartForm, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, errFileFieldRequired, http.StatusBadRequest)
		return
	}
	defer file.Close()

	resume, err := c.service.ParseAndUploadPDF(r.Context(), userID, header.Filename, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(resume)
}

// UploadMyPicture handles uploading a profile picture for the current participant.
func (c *ResumeController) UploadMyPicture(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(headerXUserID)
	if userID == "" {
		http.Error(w, errMissingUserID, http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, errParseMultipartForm, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, errFileFieldRequired, http.StatusBadRequest)
		return
	}
	defer file.Close()

	resume, err := c.service.UploadPicture(r.Context(), userID, header.Filename, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(resume)
}

// DownloadPublicMedia streams a media file (like a profile picture) from storage to the response.
func (c *ResumeController) DownloadPublicMedia(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "missing path parameter", http.StatusBadRequest)
		return
	}

	// Basic security check: only allow pictures/ prefix
	if !strings.HasPrefix(path, "pictures/") {
		http.Error(w, "unauthorized media path", http.StatusForbidden)
		return
	}

	rc, err := c.service.DownloadMedia(r.Context(), path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer rc.Close()

	ext := strings.ToLower(filepath.Ext(path))
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	
	_, err = io.Copy(w, rc)
	if err != nil {
		http.Error(w, "error streaming media", http.StatusInternalServerError)
	}
}

// GetAllParticipantsResumes returns resumes of all participants for authorized volunteers and admins.
func (c *ResumeController) GetAllParticipantsResumes(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerXUserRole)
	if !isStaffRole(role) {
		http.Error(w, "insufficient permissions: only volunteers and admins can list all resumes", http.StatusForbidden)
		return
	}

	resumes, err := c.service.GetAllResumes(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(resumes)
}

// GetUserResumeByAdmin allows volunteers and administrators to view any participant's resume.
func (c *ResumeController) GetUserResumeByAdmin(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerXUserRole)
	if !isStaffRole(role) {
		http.Error(w, errInsufficientPerms, http.StatusForbidden)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, errMissingParamUserID, http.StatusBadRequest)
		return
	}

	resume, err := c.service.GetResumeByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(resume)
}

// GetUserPDFStatusByAdmin allows volunteers and administrators to check PDF status.
func (c *ResumeController) GetUserPDFStatusByAdmin(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerXUserRole)
	if !isStaffRole(role) {
		http.Error(w, errInsufficientPerms, http.StatusForbidden)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, errMissingParamUserID, http.StatusBadRequest)
		return
	}

	ready := c.service.GetPDFStatus(r.Context(), userID)
	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(map[string]bool{"ready": ready})
}

// UpdateUserResumeByAdmin allows volunteers and administrators to update any participant's resume.
func (c *ResumeController) UpdateUserResumeByAdmin(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerXUserRole)
	if !isStaffRole(role) {
		http.Error(w, errInsufficientPerms, http.StatusForbidden)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, errMissingParamUserID, http.StatusBadRequest)
		return
	}

	var req dto.UpdateResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, errInvalidJSON, http.StatusBadRequest)
		return
	}
	req.Sanitize()
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resume, err := c.service.UpdateResume(r.Context(), userID, &req.ResumeData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(resume)
}

// UploadUserPictureByAdmin allows volunteers and administrators to upload a picture for any participant.
func (c *ResumeController) UploadUserPictureByAdmin(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerXUserRole)
	if !isStaffRole(role) {
		http.Error(w, errInsufficientPerms, http.StatusForbidden)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, errMissingParamUserID, http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, errParseMultipartForm, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, errFileFieldRequired, http.StatusBadRequest)
		return
	}
	defer file.Close()

	resume, err := c.service.UploadPicture(r.Context(), userID, header.Filename, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(resume)
}

// GenerateMyResume explicitly triggers synchronization with Reactive Resume using profile data.
func (c *ResumeController) GenerateMyResume(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get(headerXUserID)
	if userID == "" {
		http.Error(w, errMissingUserID, http.StatusUnauthorized)
		return
	}

	var req dto.GenerateResumeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, errInvalidJSON, http.StatusBadRequest)
		return
	}
	req.Sanitize()
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resume, err := c.service.GenerateResume(r.Context(), userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(resume)
}

// HealthCheck verifies service operational status.
func (c *ResumeController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "UP", "service": "rxresume-service"})
}

func isStaffRole(role string) bool {
	r := strings.ToLower(role)
	return strings.Contains(r, "admin") || strings.Contains(r, "volunteer") || strings.Contains(r, "voluntario")
}

// GetPublicResume retrieves a resume by user_id for public viewing without auth checks.
func (c *ResumeController) GetPublicResume(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "missing user ID", http.StatusBadRequest)
		return
	}

	resume, err := c.service.GetResumeByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(resume)
}

// DownloadPublicPDF retrieves the PDF from Reactive Resume backend and serves it.
func (c *ResumeController) DownloadPublicPDF(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "missing user ID", http.StatusBadRequest)
		return
	}

	pdfBytes, err := c.service.ExportPDF(r.Context(), userID)
	if err != nil {
		if errors.Is(err, services.ErrPDFNotReady) {
			w.Header().Set(headerContentType, contentTypeJSON)
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"ready":   false,
				"message": "El PDF está siendo generado. Por favor, espere.",
			})
			return
		}
		http.Error(w, "error generating PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(headerContentType, "application/pdf")
	w.Header().Set("Content-Disposition", "inline; filename=\"resume.pdf\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}
