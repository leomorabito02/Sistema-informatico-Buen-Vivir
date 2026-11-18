package controllers

import "net/http"

// HealthCheck - GET /health
func HealthCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "UP", "service": "human-promotion-service"})
}
