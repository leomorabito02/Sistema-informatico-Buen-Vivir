package controllers

import (
	"log/slog"
	"net"
	"net/http"
	"time"

	"buenvivir/campaigns/internal/middlewares"
	"buenvivir/campaigns/internal/repositories"
)

// Acciones auditadas de Campañas (D8, 22/11/2026).
const (
	AccionCrear    = "crear"
	AccionEditar   = "editar"
	AccionPublicar = "publicar"
	AccionBaja     = "baja"
)

// WithAudit activa el registro de auditoría (sin esto, no se registra).
func (c *CampaignController) WithAudit(repo repositories.AuditRepository) *CampaignController {
	c.audit = repo
	return c
}

// registrar agrega una fila de auditoría después de una escritura exitosa.
// Si falla no corta el pedido: queda en el log técnico.
func (c *CampaignController) registrar(r *http.Request, accion, campanaID string) {
	if c.audit == nil {
		return
	}
	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		ip = host
	}
	err := c.audit.Record(r.Context(), repositories.AuditEntry{
		UsuarioID: middlewares.UserIDFromContext(r.Context()),
		Rol:       middlewares.RoleFromContext(r.Context()),
		Accion:    accion,
		Recurso:   "campana",
		RecursoID: campanaID,
		IP:        ip,
	})
	if err != nil {
		slog.Error("auditoría: no se pudo registrar", slog.String("accion", accion), slog.String("campana", campanaID), slog.String("error", err.Error()))
	}
}

// Auditoria: GET /api/v1/campaigns/auditoria?desde=AAAA-MM-DD&hasta=&accion=
// Solo administración, solo lectura (mismo contrato que la de Inserción
// Laboral).
func (c *CampaignController) Auditoria(w http.ResponseWriter, r *http.Request) {
	if c.audit == nil {
		writeJSON(w, http.StatusOK, []repositories.AuditEntry{})
		return
	}
	loc, err := time.LoadLocation("America/Argentina/Cordoba")
	if err != nil {
		loc = time.FixedZone("ART", -3*3600)
	}
	q := r.URL.Query()
	f := repositories.AuditFilter{Accion: q.Get("accion")}
	if d := q.Get("desde"); d != "" {
		t, err := time.ParseInLocation("2006-01-02", d, loc)
		if err != nil {
			writeErrorJSON(w, http.StatusBadRequest, "desde debe tener formato AAAA-MM-DD")
			return
		}
		f.Desde = t
	}
	if h := q.Get("hasta"); h != "" {
		t, err := time.ParseInLocation("2006-01-02", h, loc)
		if err != nil {
			writeErrorJSON(w, http.StatusBadRequest, "hasta debe tener formato AAAA-MM-DD")
			return
		}
		f.Hasta = t.AddDate(0, 0, 1)
	}
	rows, err := c.audit.List(r.Context(), f)
	if err != nil {
		slog.Error("auditoría: lectura fallida", slog.String("error", err.Error()))
		writeErrorJSON(w, http.StatusInternalServerError, "No se pudo leer la auditoría")
		return
	}
	writeJSON(w, http.StatusOK, rows)
}
