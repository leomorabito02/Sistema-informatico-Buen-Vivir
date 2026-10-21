package reportes

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
)

// Handlers expone el historial (solo administración: el guard de rutas
// lo protege). Con servicio nil responde 503: el historial no está
// configurado en ese entorno, pero el resto del dashboard funciona.
type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers { return &Handlers{svc: svc} }

// Registrar: POST /api/v1/dashboard/reportes, multipart/form-data con
// "metadatos" (JSON) y "archivo" (el PDF).
func (h *Handlers) Registrar(w http.ResponseWriter, r *http.Request) {
	if !h.disponible(w) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, MaxPDFBytes+(1<<20))
	if err := r.ParseMultipartForm(MaxPDFBytes + (1 << 20)); err != nil {
		writeError(w, http.StatusBadRequest, "El pedido tiene que traer el PDF (hasta 10 MB) y sus datos")
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	var meta Metadatos
	if err := json.Unmarshal([]byte(r.FormValue("metadatos")), &meta); err != nil {
		writeError(w, http.StatusBadRequest, "Los datos del reporte no son válidos")
		return
	}
	f, _, err := r.FormFile("archivo")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Falta el PDF del reporte")
		return
	}
	defer func() { _ = f.Close() }()
	pdf, err := io.ReadAll(io.LimitReader(f, MaxPDFBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "No se pudo leer el PDF")
		return
	}

	rep, err := h.svc.Registrar(r.Context(), meta, pdf, r.Header.Get("X-User-ID"))
	if err != nil {
		h.error(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, rep)
}

// Listar: GET /api/v1/dashboard/reportes.
func (h *Handlers) Listar(w http.ResponseWriter, r *http.Request) {
	if !h.disponible(w) {
		return
	}
	reps, err := h.svc.Listar(r.Context())
	if err != nil {
		h.error(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, reps)
}

// Descargar: GET /api/v1/dashboard/reportes/{id}/pdf.
func (h *Handlers) Descargar(w http.ResponseWriter, r *http.Request) {
	if !h.disponible(w) {
		return
	}
	rep, rc, err := h.svc.AbrirPDF(r.Context(), r.PathValue("id"))
	if err != nil {
		h.error(w, r, err)
		return
	}
	defer func() { _ = rc.Close() }()
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="reporte-dashboard-`+rep.FechaGeneracion.Format("2006-01-02")+`.pdf"`)
	w.Header().Set("Cache-Control", "no-store")
	if _, err := io.Copy(w, rc); err != nil {
		log.Printf("reportes: error enviando el PDF %s: %v", rep.ID, err)
	}
}

func (h *Handlers) disponible(w http.ResponseWriter) bool {
	if h == nil || h.svc == nil {
		writeError(w, http.StatusServiceUnavailable, "El historial de reportes no está configurado en este entorno")
		return false
	}
	return true
}

func (h *Handlers) error(w http.ResponseWriter, r *http.Request, err error) {
	var v *ErrValidacion
	switch {
	case errors.As(err, &v):
		writeError(w, http.StatusBadRequest, v.Mensaje)
	case errors.Is(err, ErrNoEncontrado):
		writeError(w, http.StatusNotFound, "El reporte no existe")
	default:
		log.Printf("reportes %s: %v", r.URL.Path, err)
		writeError(w, http.StatusInternalServerError, "No se pudo guardar o leer el reporte. Probá de nuevo en unos minutos.")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
