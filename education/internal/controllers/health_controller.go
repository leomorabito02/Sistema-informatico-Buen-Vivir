package controllers

import (
	"net/http"

	"buenvivir/education/internal/config"
)

type HealthController struct {
	cfg *config.Config
}

func NewHealthController(cfg *config.Config) *HealthController {
	return &HealthController{cfg: cfg}
}

// Health responde GET /api/salud, informando qué ambiente y qué
// Contentful Environment tiene cargados esta instancia — útil para
// confirmar rápido, sin leer logs, contra qué está hablando el
// servicio (clave para no confundir testing con producción).
func (c *HealthController) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":                "UP",
		"service":               "education-service",
		"appEnv":                string(c.cfg.AppEnv),
		"contentfulEnvironment": string(c.cfg.ContentfulEnvironment),
	})
}
