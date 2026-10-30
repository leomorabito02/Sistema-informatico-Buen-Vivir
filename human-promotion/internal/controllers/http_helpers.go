package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"buenvivir/human-promotion/internal/repositories"
	"buenvivir/human-promotion/internal/services"
)

// maxBodyBytes acota el cuerpo de las requests: los textos más largos
// (situación y avances, 5000 caracteres cada uno) entran holgados.
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

// writeError traduce un error de las capas de service/repository a una
// respuesta HTTP: 400 validación, 404 no encontrado, 409 ficha
// duplicada y 500 para todo lo demás. El detalle de un error interno
// solo va al log, nunca a la respuesta.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case services.IsValidationError(err):
		writeErrorJSON(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, repositories.ErrNotFound):
		writeErrorJSON(w, http.StatusNotFound, "recurso no encontrado")
	case errors.Is(err, repositories.ErrAlreadyExists):
		writeErrorJSON(w, http.StatusConflict, "El participante ya tiene una ficha de proceso personal")
	default:
		slog.Error("error interno",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("error", err.Error()),
		)
		writeErrorJSON(w, http.StatusInternalServerError, "error interno del servidor")
	}
}
