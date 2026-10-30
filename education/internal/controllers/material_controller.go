package controllers

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

// maxUploadBytes es el límite duro del multipart (12 MB) — el límite
// fino por tipo (5 MB imágenes / 10 MB docs) lo aplica el service.
const maxUploadBytes = 12 << 20

type MaterialController struct {
	service services.MaterialService
}

func NewMaterialController(service services.MaterialService) *MaterialController {
	return &MaterialController{service: service}
}

// List responde GET /api/materiales
func (c *MaterialController) List(w http.ResponseWriter, r *http.Request) {
	materials, err := c.service.ListMaterials(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, materials)
}

// Create responde POST /api/materiales
func (c *MaterialController) Create(w http.ResponseWriter, r *http.Request) {
	var input models.MaterialInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	material, err := c.service.CreateMaterial(r.Context(), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, material)
}

// Upload responde POST /api/materiales/upload (multipart/form-data):
// campos "file" (el binario), "title" y "type". Crea un material con el
// archivo subido a Contentful como Asset.
func (c *MaterialController) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	// #nosec G120 -- request body is bounded by http.MaxBytesReader above
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "el archivo supera el límite o el formulario es inválido")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "falta el archivo (campo 'file')")
		return
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "no se pudo leer el archivo")
		return
	}

	contentType := header.Header.Get("Content-Type")
	material, err := c.service.UploadMaterial(r.Context(), services.MaterialUpload{
		Title:       r.FormValue("title"),
		Type:        r.FormValue("type"),
		FileName:    header.Filename,
		ContentType: contentType,
		Data:        data,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, material)
}

// Update responde PUT /api/materiales/{id}
func (c *MaterialController) Update(w http.ResponseWriter, r *http.Request) {
	var input models.MaterialInput
	if err := decodeJSONBody(r, &input); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido: "+err.Error())
		return
	}

	material, err := c.service.UpdateMaterial(r.Context(), chi.URLParam(r, "id"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, material)
}

// Delete responde DELETE /api/materiales/{id}
func (c *MaterialController) Delete(w http.ResponseWriter, r *http.Request) {
	if err := c.service.DeleteMaterial(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
