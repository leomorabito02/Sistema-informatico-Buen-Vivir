package controllers

import (
	"encoding/json"
	"net/http"

	"auth/internal/middlewares"
	"auth/internal/repositories"
)

type DeviceController struct {
	repo repositories.DeviceRepository
}

func NewDeviceController(repo repositories.DeviceRepository) *DeviceController {
	return &DeviceController{repo: repo}
}

func (c *DeviceController) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	var req struct {
		Token     string `json:"token"`
		UserAgent string `json:"userAgent"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		http.Error(w, "token requerido", http.StatusBadRequest)
		return
	}

	if err := c.repo.RegisterDevice(r.Context(), userID, req.Token, req.UserAgent); err != nil {
		http.Error(w, "no se pudo registrar el dispositivo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "dispositivo registrado"})
}
