package controllers

import (
	"encoding/json"
	"log/slog"
	"net"
	"net/http"
	"time"

	"buenvivir/rxresume/internal/models"
	"buenvivir/rxresume/internal/repositories"
)

// Acciones auditadas sobre el CV de otra persona.
const (
	AccionVer         = "ver"
	AccionEditar      = "editar"
	AccionCambiarFoto = "cambiar_foto"
)

// WithAudit activa el registro de auditoría (sin esto, no se registra).
func (c *ResumeController) WithAudit(repo repositories.AuditRepository) *ResumeController {
	c.audit = repo
	return c
}

// registrar agrega una fila después de una acción de gestión exitosa. Si
// falla no corta el pedido: queda en el log técnico.
func (c *ResumeController) registrar(r *http.Request, accion, usuarioAfectado string) {
	if c.audit == nil {
		return
	}
	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		ip = host
	}
	err := c.audit.Record(r.Context(), models.AuditEntry{
		UsuarioID: r.Header.Get(headerXUserID),
		Rol:       r.Header.Get(headerXUserRole),
		Accion:    accion,
		Recurso:   "cv",
		RecursoID: usuarioAfectado,
		IP:        ip,
	})
	if err != nil {
		slog.Error("auditoría: no se pudo registrar", slog.String("accion", accion), slog.String("error", err.Error()))
	}
}

// Auditoria: GET /api/v1/rxresume/auditoria?desde=AAAA-MM-DD&hasta=&accion=
// Solo administración, solo lectura.
func (c *ResumeController) Auditoria(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerXUserRole)
	if role != "role_admin" && role != "role_dev" {
		http.Error(w, errInsufficientPerms, http.StatusForbidden)
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
			http.Error(w, "desde debe tener formato AAAA-MM-DD", http.StatusBadRequest)
			return
		}
		f.Desde = t
	}
	if h := q.Get("hasta"); h != "" {
		t, err := time.ParseInLocation("2006-01-02", h, loc)
		if err != nil {
			http.Error(w, "hasta debe tener formato AAAA-MM-DD", http.StatusBadRequest)
			return
		}
		f.Hasta = t.AddDate(0, 0, 1)
	}
	rows := []models.AuditEntry{}
	if c.audit != nil {
		rows, err = c.audit.List(r.Context(), f)
		if err != nil {
			slog.Error("auditoría: lectura fallida", slog.String("error", err.Error()))
			http.Error(w, "no se pudo leer la auditoría", http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set(headerContentType, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(rows)
}
