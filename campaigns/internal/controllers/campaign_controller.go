package controllers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"buenvivir/campaigns/internal/middlewares"
	"buenvivir/campaigns/internal/models"
	"buenvivir/campaigns/internal/services"
)

type CampaignController struct {
	service services.CampaignService
}

func NewCampaignController(service services.CampaignService) *CampaignController {
	return &CampaignController{service: service}
}

// respuestaConMensaje es la respuesta de las escrituras: la campaña como
// quedó y el mensaje de éxito que pide la HU.
type respuestaConMensaje struct {
	Mensaje string          `json:"mensaje"`
	Campana models.Campaign `json:"campana"`
}

// ListPublished: GET /api/v1/campaigns/publicas?q= — sin autenticación
// (HU-50), solo publicadas, con búsqueda por título (HU-51).
func (c *CampaignController) ListPublished(w http.ResponseWriter, r *http.Request) {
	campaigns, err := c.service.ListPublished(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	// El público no necesita saber qué usuario interno creó la campaña.
	for i := range campaigns {
		campaigns[i].CreadoPor = ""
	}
	// HU-50: con ?limit= se devuelve una página {items, total}; sin
	// limit, la lista completa como antes (compatibilidad).
	if raw := r.URL.Query().Get("limit"); raw != "" {
		page, ok := paginar(campaigns, raw, r.URL.Query().Get("offset"))
		if !ok {
			writeErrorJSON(w, http.StatusBadRequest, "limit y offset deben ser números (limit entre 1 y 100)")
			return
		}
		writeJSON(w, http.StatusOK, page)
		return
	}
	writeJSON(w, http.StatusOK, campaigns)
}

// PaginaCampanas es la respuesta paginada del listado público (HU-50).
type PaginaCampanas struct {
	Items []models.Campaign `json:"items"`
	Total int               `json:"total"`
}

func paginar(all []models.Campaign, rawLimit, rawOffset string) (PaginaCampanas, bool) {
	limit, err := strconv.Atoi(rawLimit)
	if err != nil || limit < 1 || limit > 100 {
		return PaginaCampanas{}, false
	}
	offset := 0
	if rawOffset != "" {
		if offset, err = strconv.Atoi(rawOffset); err != nil || offset < 0 {
			return PaginaCampanas{}, false
		}
	}
	page := PaginaCampanas{Items: []models.Campaign{}, Total: len(all)}
	if offset < len(all) {
		end := offset + limit
		if end > len(all) {
			end = len(all)
		}
		page.Items = all[offset:end]
	}
	return page, true
}

// List: GET /api/v1/campaigns?q=&estado= — gestión (borradores y
// publicadas).
func (c *CampaignController) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	campaigns, err := c.service.List(r.Context(), q.Get("q"), q.Get("estado"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, campaigns)
}

// Get: GET /api/v1/campaigns/{id} — gestión.
func (c *CampaignController) Get(w http.ResponseWriter, r *http.Request) {
	campaign, err := c.service.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, campaign)
}

// Create: POST /api/v1/campaigns (HU-47).
func (c *CampaignController) Create(w http.ResponseWriter, r *http.Request) {
	var in models.CampaignInput
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "El cuerpo de la solicitud no es válido")
		return
	}
	campaign, err := c.service.Create(r.Context(), in, middlewares.UserIDFromContext(r.Context()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, respuestaConMensaje{Mensaje: services.MsgCampanaCreada, Campana: campaign})
}

// Update: PUT /api/v1/campaigns/{id} (HU-48).
func (c *CampaignController) Update(w http.ResponseWriter, r *http.Request) {
	var in models.CampaignInput
	if err := decodeJSONBody(w, r, &in); err != nil {
		writeErrorJSON(w, http.StatusBadRequest, "El cuerpo de la solicitud no es válido")
		return
	}
	campaign, err := c.service.Update(r.Context(), chi.URLParam(r, "id"), in)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, respuestaConMensaje{Mensaje: services.MsgCampanaEditada, Campana: campaign})
}

// Publish: POST /api/v1/campaigns/{id}/publicar (HU-49).
func (c *CampaignController) Publish(w http.ResponseWriter, r *http.Request) {
	campaign, err := c.service.Publish(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, respuestaConMensaje{Mensaje: services.MsgCampanaPublicada, Campana: campaign})
}

// Delete: DELETE /api/v1/campaigns/{id} — baja lógica (auditoría
// 2026-11-11, Problema 5). Responde el mensaje de éxito; 404 si no
// existe o ya estaba dada de baja.
func (c *CampaignController) Delete(w http.ResponseWriter, r *http.Request) {
	err := c.service.Delete(r.Context(), chi.URLParam(r, "id"), middlewares.UserIDFromContext(r.Context()))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"mensaje": services.MsgCampanaEliminada})
}

// HealthCheck: GET /health.
func (c *CampaignController) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "UP", "service": "campaigns-service"})
}
