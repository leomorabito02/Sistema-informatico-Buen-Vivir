package controllers

import (
	"encoding/json"
	"net/http"

	"buenvivir/education/internal/repositories"
	"buenvivir/education/internal/services"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func writeErrorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func decodeJSONBody(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

// writeError traduce un error de las capas de service/repository a una
// respuesta HTTP apropiada: 400 si es de validación de negocio, 404 si
// Contentful no encontró la entry, 409 si hubo conflicto de versión
// (edición concurrente), y 502 para cualquier otra falla de Contentful.
func writeError(w http.ResponseWriter, err error) {
	switch {
	case services.IsValidationError(err):
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
	case repositories.IsNotFound(err):
		writeErrorJSON(w, http.StatusNotFound, "recurso no encontrado")
	case repositories.IsConflict(err):
		writeErrorJSON(w, http.StatusConflict, "el recurso fue modificado por otra edición, volvé a intentar")
	default:
		writeErrorJSON(w, http.StatusBadGateway, "error hablando con Contentful: "+err.Error())
	}
}
