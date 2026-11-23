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
	defer func() { _ = r.Body.Close() }()
	return json.NewDecoder(r.Body).Decode(dst)
}

// writeError traduce errores de service/repository a HTTP: 400 validación, 403
// ownership, 404 no encontrado, 409 conflicto de versión, 422 rechazo de Contentful,
// 503 IA no disponible y 502 el resto.
func writeError(w http.ResponseWriter, err error) {
	switch {
	case services.IsValidationError(err):
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
	case services.IsForbiddenError(err):
		writeErrorJSON(w, http.StatusForbidden, err.Error())
	case services.IsUnavailableError(err):
		writeErrorJSON(w, http.StatusServiceUnavailable, err.Error())
	case repositories.IsNotFound(err):
		writeErrorJSON(w, http.StatusNotFound, "recurso no encontrado")
	case repositories.IsConflict(err):
		writeErrorJSON(w, http.StatusConflict, "el recurso fue modificado por otra edición, volvé a intentar")
	default:
		if msg, ok := repositories.ValidationMessage(err); ok {
			writeErrorJSON(w, http.StatusUnprocessableEntity, msg)
			return
		}
		writeErrorJSON(w, http.StatusBadGateway, "error de infraestructura: "+err.Error())
	}
}
