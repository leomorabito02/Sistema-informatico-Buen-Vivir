package controllers

import (
	"io"
	"mime"
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

// maxContenidoUploadBytes es el límite duro del multipart de un material
// (un archivo de 10 MB más los campos de texto).
const maxContenidoUploadBytes = 11 << 20

// TeacherContentController expone el contenido de docentes externos con
// moderación (HU-23) y la biblioteca pública con registro de descargas
// (HU-25).
type TeacherContentController struct {
	service services.TeacherContentService
}

func NewTeacherContentController(service services.TeacherContentService) *TeacherContentController {
	return &TeacherContentController{service: service}
}

func actor(r *http.Request) (userID, roleID string) {
	userID, _ = r.Context().Value(middlewares.UserIDKey).(string)
	roleID, _ = r.Context().Value(middlewares.RoleIDKey).(string)
	return userID, roleID
}

// leerMaterial lee el multipart: "titulo", "descripcion" y el archivo en
// "archivo" (opcional en la edición).
func leerMaterial(w http.ResponseWriter, r *http.Request) (models.ContenidoInput, *models.ArchivoSubido, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxContenidoUploadBytes)
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "no se pudo leer el material (¿el archivo pesa más de 10 MB?)")
		return models.ContenidoInput{}, nil, false
	}
	in := models.ContenidoInput{Titulo: r.FormValue("titulo"), Descripcion: r.FormValue("descripcion")}
	files := r.MultipartForm.File["archivo"]
	if len(files) == 0 {
		return in, nil, true
	}
	fh := files[0]
	f, err := fh.Open()
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "no se pudo leer el archivo "+fh.Filename)
		return in, nil, false
	}
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(f)
	if err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "no se pudo leer el archivo "+fh.Filename)
		return in, nil, false
	}
	return in, &models.ArchivoSubido{
		Nombre: fh.Filename,
		Tipo:   uploadContentType(fh.Filename, fh.Header.Get("Content-Type")),
		Data:   data,
	}, true
}

// Upload responde POST /api/contenidos-docentes (docente).
func (c *TeacherContentController) Upload(w http.ResponseWriter, r *http.Request) {
	in, archivo, ok := leerMaterial(w, r)
	if !ok {
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()
	userID, _ := actor(r)
	created, err := c.service.Upload(r.Context(), userID, in, archivo)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// ListMine responde GET /api/contenidos-docentes/mios (docente).
func (c *TeacherContentController) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, _ := actor(r)
	list, err := c.service.ListMine(r.Context(), userID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// Update responde PUT /api/contenidos-docentes/{id} (el dueño, pendiente).
func (c *TeacherContentController) Update(w http.ResponseWriter, r *http.Request) {
	in, archivo, ok := leerMaterial(w, r)
	if !ok {
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()
	userID, _ := actor(r)
	updated, err := c.service.Update(r.Context(), userID, chi.URLParam(r, "id"), in, archivo)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// Delete responde DELETE /api/contenidos-docentes/{id}.
func (c *TeacherContentController) Delete(w http.ResponseWriter, r *http.Request) {
	userID, roleID := actor(r)
	if err := c.service.Delete(r.Context(), userID, roleID, chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// List responde GET /api/contenidos-docentes?estado= (administrador).
func (c *TeacherContentController) List(w http.ResponseWriter, r *http.Request) {
	list, err := c.service.List(r.Context(), r.URL.Query().Get("estado"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// Moderate responde POST /api/contenidos-docentes/{id}/moderacion con
// body {estado: "aprobado"|"rechazado", motivo?} (administrador).
func (c *TeacherContentController) Moderate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Estado string `json:"estado"`
		Motivo string `json:"motivo"`
	}
	if err := decodeJSONBody(r, &body); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido")
		return
	}
	userID, _ := actor(r)
	out, err := c.service.Moderate(r.Context(), userID, chi.URLParam(r, "id"), body.Estado, body.Motivo)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// ListDownloads responde GET /api/biblioteca/descargas (administrador).
func (c *TeacherContentController) ListDownloads(w http.ResponseWriter, r *http.Request) {
	list, err := c.service.ListDownloads(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// ListPublic responde GET /api/biblioteca (público).
func (c *TeacherContentController) ListPublic(w http.ResponseWriter, r *http.Request) {
	list, err := c.service.ListPublic(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// PublicDownload responde POST /api/biblioteca/{id}/descargas (público)
// con body {nombre, apellido, email, telefono}: registra la descarga y
// devuelve el archivo.
func (c *TeacherContentController) PublicDownload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 8<<10)
	var in models.DescargaInput
	if err := decodeJSONBody(r, &in); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido")
		return
	}
	contenido, rc, err := c.service.PublicDownload(r.Context(), chi.URLParam(r, "id"), in)
	if err != nil {
		writeError(w, err)
		return
	}
	writeFile(w, contenido, rc)
}

// File responde GET /api/contenidos-docentes/{id}/archivo (con sesión).
func (c *TeacherContentController) File(w http.ResponseWriter, r *http.Request) {
	userID, roleID := actor(r)
	contenido, rc, err := c.service.File(r.Context(), userID, roleID, chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeFile(w, contenido, rc)
}

func writeFile(w http.ResponseWriter, c *models.ContenidoDocente, rc io.ReadCloser) {
	defer func() { _ = rc.Close() }()
	w.Header().Set("Content-Type", c.ArchivoTipo)
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": c.ArchivoNombre}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, rc)
}
