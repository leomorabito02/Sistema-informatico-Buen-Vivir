package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"buenvivir/campaigns/internal/repositories"
	"buenvivir/campaigns/internal/services"
)

// maxBodyBytes acota el cuerpo de las requests: el texto más largo (5000
// caracteres) y la URL de la imagen entran holgados.
const maxBodyBytes = 64 << 10

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

// decodeJSONBody rechaza campos desconocidos y cuerpos demasiado grandes.
func decodeJSONBody(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	defer func() { _ = r.Body.Close() }()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// writeError traduce un error de service/repository a HTTP: 400
// validación, 404 no encontrada y 500 para todo lo demás. El detalle de
// un error interno solo va al log.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case services.IsValidationError(err):
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, repositories.ErrNotFound):
		writeErrorJSON(w, http.StatusNotFound, "La campaña no existe")
	default:
		slog.Error("error interno",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("error", err.Error()),
		)
		writeErrorJSON(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
