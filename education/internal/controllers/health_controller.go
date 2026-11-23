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

// Health responde GET /api/salud con el ambiente y el Environment de Contentful cargados.
func (c *HealthController) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":                "UP",
		"service":               "education-service",
		"appEnv":                string(c.cfg.AppEnv),
		"contentfulEnvironment": string(c.cfg.ContentfulEnvironment),
	})
}
