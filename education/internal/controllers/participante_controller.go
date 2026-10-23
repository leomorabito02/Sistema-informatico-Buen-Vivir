package controllers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"buenvivir/education/internal/middlewares"
	"buenvivir/education/internal/services"
)

// ParticipanteController expone, para admin y voluntario, el buscador de
// participantes de los selectores de inscripción y el historial de
// asistencia de un participante (HU-19).
type ParticipanteController struct {
	service services.ParticipanteService
}

func NewParticipanteController(service services.ParticipanteService) *ParticipanteController {
	return &ParticipanteController{service: service}
}

// Buscar responde GET /api/participantes?q=texto&cursoId=id. Busca en el
// backend (antes el navegador bajaba el directorio entero y filtraba).
// Con cursoId deja afuera a los ya inscriptos en ese curso.
func (c *ParticipanteController) Buscar(w http.ResponseWriter, r *http.Request) {
	token, _ := r.Context().Value(middlewares.TokenKey).(string)
	q := r.URL.Query()
	out, err := c.service.Buscar(r.Context(), q.Get("q"), q.Get("cursoId"), token)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// HistorialAsistencia responde GET /api/participants/{id}/asistencia.
func (c *ParticipanteController) HistorialAsistencia(w http.ResponseWriter, r *http.Request) {
	out, err := c.service.HistorialAsistencia(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}
