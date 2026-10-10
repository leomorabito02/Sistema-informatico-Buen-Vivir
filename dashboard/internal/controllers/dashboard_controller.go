package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"buenvivir/dashboard/internal/clients"
	"buenvivir/dashboard/internal/services"
)

type DashboardController struct {
	service services.DashboardService
}

func NewDashboardController(service services.DashboardService) *DashboardController {
	return &DashboardController{service: service}
}

// RequireAdmin deja pasar solo a role_admin / role_dev. El rol lo pone el
// API Gateway (X-User-Role) después de validar el token con auth.
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("X-User-Role") {
		case "role_admin", "role_dev":
			next.ServeHTTP(w, r)
		case "":
			writeError(w, http.StatusUnauthorized, "Iniciá sesión para ver el dashboard.")
		default:
			writeError(w, http.StatusForbidden, "El dashboard es solo para la administración del centro.")
		}
	})
}

func (c *DashboardController) Participantes(w http.ResponseWriter, r *http.Request) {
	responder(w, r, c.service.Participantes)
}

func (c *DashboardController) Cursos(w http.ResponseWriter, r *http.Request) {
	responder(w, r, c.service.Cursos)
}

func (c *DashboardController) Habilidades(w http.ResponseWriter, r *http.Request) {
	responder(w, r, c.service.Habilidades)
}

func (c *DashboardController) Postulaciones(w http.ResponseWriter, r *http.Request) {
	responder(w, r, c.service.Postulaciones)
}

func (c *DashboardController) Areas(w http.ResponseWriter, r *http.Request) {
	responder(w, r, c.service.Areas)
}

func (c *DashboardController) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "UP", "service": "dashboard-service"})
}

// responder llama a la fuente con las credenciales de quien pidió y
// traduce los errores: 401/403 del servicio de origen se respetan; el
// resto es 502 (el origen falló), y el front marca ese indicador como
// "No disponible" sin romper los demás.
func responder[T any](w http.ResponseWriter, r *http.Request, fuente func(context.Context, clients.Credentials) ([]T, error)) {
	data, err := fuente(r.Context(), clients.FromRequest(r))
	if err != nil {
		var up *clients.UpstreamError
		if errors.As(err, &up) && (up.Status == http.StatusUnauthorized || up.Status == http.StatusForbidden) {
			writeError(w, up.Status, "No tenés permiso para ver estos datos.")
			return
		}
		log.Printf("dashboard %s: %v", r.URL.Path, err)
		writeError(w, http.StatusBadGateway, "El servicio de origen no respondió. Probá de nuevo en unos minutos.")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
