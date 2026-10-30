package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/models"
	"buenvivir/education/internal/services"
)

// CommentController expone los comentarios de cursos y campañas
// (HU-24). Leer y comentar es público (el contenido se ve sin login);
// borrar es la moderación del administrador.
type CommentController struct {
	service services.CommentService
}

func NewCommentController(service services.CommentService) *CommentController {
	return &CommentController{service: service}
}

// List responde GET /api/comentarios?entidadTipo=curso|campana&entidadId=...
func (c *CommentController) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	list, err := c.service.List(r.Context(), q.Get("entidadTipo"), q.Get("entidadId"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// Create responde POST /api/comentarios con body
// {entidadTipo, entidadId, nombre, apellido, texto}.
func (c *CommentController) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var in models.ComentarioInput
	if err := decodeJSONBody(r, &in); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "body inválido")
		return
	}
	usuarioID, _ := r.Context().Value(middlewares.UserIDKey).(string)
	created, err := c.service.Create(r.Context(), in, usuarioID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// Delete responde DELETE /api/comentarios/{id} (solo administrador).
func (c *CommentController) Delete(w http.ResponseWriter, r *http.Request) {
	if err := c.service.Delete(r.Context(), chi.URLParam(r, "id")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
