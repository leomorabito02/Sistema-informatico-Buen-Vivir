package controllers

import (
	"net/http"
	"time"

	"auth/internal/repositories"
)

// RegistroAuditoria es una fila de GET /api/v1/auth/auditoria. Mismo
// formato que el de Educación y Promoción Humana, más el rol, los
// campos tocados (solo nombres) y el resultado (ok o denegado).
type RegistroAuditoria struct {
	ID            uint      `json:"id"`
	UsuarioID     string    `json:"usuario_id"`
	UsuarioNombre string    `json:"usuario_nombre,omitempty"`
	Rol           string    `json:"rol,omitempty"`
	Accion        string    `json:"accion"`
	Recurso       string    `json:"recurso"`
	RecursoID     string    `json:"recurso_id"`
	Campos        string    `json:"campos,omitempty"`
	Resultado     string    `json:"resultado"`
	Fecha         time.Time `json:"fecha"`
	IP            string    `json:"ip,omitempty"`
}

// AuditController expone el log de auditoría de Autenticación en solo
// lectura y solo para administración (HU-10).
type AuditController struct {
	repo repositories.AuditRepository
}

func NewAuditController(repo repositories.AuditRepository) *AuditController {
	return &AuditController{repo: repo}
}

var zonaAR = func() *time.Location {
	if loc, err := time.LoadLocation("America/Argentina/Cordoba"); err == nil {
		return loc
	}
	return time.FixedZone("ART", -3*3600)
}()

// List - GET /api/v1/auth/auditoria?desde=YYYY-MM-DD&hasta=YYYY-MM-DD&accion=&recurso=
func (c *AuditController) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := repositories.AuditFilter{Accion: q.Get("accion"), Recurso: q.Get("recurso")}
	if d := q.Get("desde"); d != "" {
		t, err := time.ParseInLocation("2006-01-02", d, zonaAR)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "desde debe tener formato AAAA-MM-DD")
			return
		}
		f.Desde = t
	}
	if h := q.Get("hasta"); h != "" {
		t, err := time.ParseInLocation("2006-01-02", h, zonaAR)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "hasta debe tener formato AAAA-MM-DD")
			return
		}
		f.Hasta = t.AddDate(0, 0, 1) // inclusive
	}
	if !f.Desde.IsZero() && !f.Hasta.IsZero() && f.Hasta.Before(f.Desde) {
		writeJSONError(w, http.StatusBadRequest, "el rango de fechas es inválido")
		return
	}

	rows, err := c.repo.List(r.Context(), f)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "no se pudo leer la auditoría")
		return
	}
	ids := make([]string, 0, len(rows))
	seen := map[string]bool{}
	for _, row := range rows {
		if row.UsuarioID != "" && !seen[row.UsuarioID] {
			seen[row.UsuarioID] = true
			ids = append(ids, row.UsuarioID)
		}
	}
	names, _ := c.repo.UserNames(r.Context(), ids)

	out := make([]RegistroAuditoria, 0, len(rows))
	for _, row := range rows {
		out = append(out, RegistroAuditoria{
			ID: row.ID, UsuarioID: row.UsuarioID, UsuarioNombre: names[row.UsuarioID], Rol: row.Rol,
			Accion: row.Accion, Recurso: row.Recurso, RecursoID: row.RecursoID, Campos: row.Campos,
			Resultado: row.Resultado, Fecha: row.Fecha, IP: row.IP,
		})
	}
	writeJSONResponse(w, http.StatusOK, out)
}
