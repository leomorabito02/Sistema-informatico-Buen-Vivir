package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"buenvivir/dashboard/internal/auth"
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
// API Gateway de Docker (X-User-Role) después de validar el token con auth.
// Se usa cuando no hay JWT_SECRET (ver NewGuard).
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		permitir(w, r, next, r.Header.Get("X-User-Role"))
	})
}

// NewGuard elige cómo saber quién pide. Con validador (JWT_SECRET
// configurado, obligatorio en Google Cloud, donde el gateway no pone
// X-User-Role) valida el token por su cuenta y reemplaza X-User-ID y
// X-User-Role con lo que dice el token: así nadie puede mandarlos
// inventados, e Inserción Laboral, que confía en esos headers, recibe los
// verdaderos. Sin validador, confía en el gateway de Docker.
func NewGuard(v *auth.Validator) func(http.Handler) http.Handler {
	if v == nil {
		return RequireAdmin
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tok := auth.TokenDelPedido(r)
			if tok == "" {
				writeError(w, http.StatusUnauthorized, "Iniciá sesión para ver el dashboard.")
				return
			}
			id, err := v.Validar(tok)
			if err != nil {
				log.Printf("dashboard: token inválido en %s", r.URL.Path)
				writeError(w, http.StatusUnauthorized, "Tu sesión expiró o no es válida. Volvé a iniciar sesión.")
				return
			}
			r.Header.Set("X-User-ID", id.UserID)
			r.Header.Set("X-User-Role", id.Rol)
			permitir(w, r, next, id.Rol)
		})
	}
}

func permitir(w http.ResponseWriter, r *http.Request, next http.Handler, rol string) {
	switch rol {
	case "role_admin", "role_dev":
		next.ServeHTTP(w, r)
	case "":
		writeError(w, http.StatusUnauthorized, "Iniciá sesión para ver el dashboard.")
	default:
		writeError(w, http.StatusForbidden, "El dashboard es solo para la administración del centro.")
	}
}

func (c *DashboardController) Areas(w http.ResponseWriter, r *http.Request) {
	responder(w, r, c.service.Areas)
}

// Resumen - GET /api/v1/dashboard/resumen?desde=AAAA-MM-DD&hasta=AAAA-MM-DD&modulo=todos&area=<id>
// Indicadores y gráficas ya calculados (HU-52 a HU-54). modulo y area son
// opcionales (por defecto, todos y todas las áreas).
func (c *DashboardController) Resumen(w http.ResponseWriter, r *http.Request) {
	f, msg := leerFiltros(r)
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	writeJSON(w, http.StatusOK, services.Resumen(r.Context(), c.service, clients.FromRequest(r), f))
}

func leerFiltros(r *http.Request) (services.Filtros, string) {
	q := r.URL.Query()
	desde, err1 := time.ParseInLocation("2006-01-02", q.Get("desde"), services.Zona)
	hasta, err2 := time.ParseInLocation("2006-01-02", q.Get("hasta"), services.Zona)
	if err1 != nil || err2 != nil {
		return services.Filtros{}, "Las fechas «desde» y «hasta» tienen que tener formato AAAA-MM-DD."
	}
	if desde.After(hasta) {
		return services.Filtros{}, "La fecha «desde» no puede ser posterior a «hasta»."
	}
	if services.MesesEntre(desde, hasta) > services.MaxMesesPeriodo {
		return services.Filtros{}, fmt.Sprintf("El período no puede superar los %d años.", services.MaxMesesPeriodo/12)
	}
	modulo := q.Get("modulo")
	if modulo == "" {
		modulo = services.ModuloTodos
	}
	if !services.ModuloValido(modulo) {
		return services.Filtros{}, "Módulo desconocido."
	}
	return services.Filtros{Desde: desde, Hasta: hasta, Modulo: modulo, AreaID: q.Get("area")}, ""
}

func (c *DashboardController) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "UP", "service": "dashboard-service"})
}

// responder llama a la fuente con las credenciales de quien pidió y
// traduce los errores: 401/403 del servicio de origen se respetan; el
// resto es 502 (el origen falló), y el front marca ese indicador como
// "No disponible" sin romper los demás.
func responder[T any](w http.ResponseWriter, r *http.Request, fuente func(context.Context, clients.Credentials) (T, error)) {
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
