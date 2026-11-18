package controllers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"buenvivir/labor-integration/internal/dto"
	"buenvivir/labor-integration/internal/models"
	"buenvivir/labor-integration/internal/repositories"
	"buenvivir/labor-integration/internal/services"

	"github.com/go-chi/chi/v5"
)

const headerUserID = "X-User-ID"

type JobOfferController struct {
	service *services.JobOfferService
	audit   repositories.AuditRepository
}

// WithAudit activa el registro de auditoría (HU-42).
func (c *JobOfferController) WithAudit(repo repositories.AuditRepository) *JobOfferController {
	c.audit = repo
	return c
}

// record deja en la auditoría quién hizo qué sobre la oferta. Si falla
// queda en el log, pero no deshace la operación.
func (c *JobOfferController) record(r *http.Request, accion, ofertaID string) {
	if c.audit == nil {
		return
	}
	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		ip = host
	}
	err := c.audit.Record(r.Context(), repositories.AuditEntry{
		UsuarioID: r.Header.Get(headerUserID),
		Rol:       r.Header.Get(headerUserRole),
		Accion:    accion,
		Recurso:   "oferta",
		RecursoID: ofertaID,
		IP:        ip,
	})
	if err != nil {
		slog.Error("auditoría: no se pudo registrar", slog.String("accion", accion), slog.String("oferta", ofertaID), slog.String("error", err.Error()))
	}
}

// withToken pasa el token del pedido al servicio (cruce de HU-44).
func withToken(r *http.Request) *http.Request {
	tok := ""
	for _, n := range []string{"__Host-access_token", "access_token"} {
		if ck, err := r.Cookie(n); err == nil && ck.Value != "" {
			tok = ck.Value
			break
		}
	}
	if a := r.Header.Get("Authorization"); tok == "" && strings.HasPrefix(a, "Bearer ") {
		tok = strings.TrimSpace(strings.TrimPrefix(a, "Bearer "))
	}
	if tok == "" {
		return r
	}
	return r.WithContext(services.WithBearer(r.Context(), tok))
}

func NewJobOfferController(service *services.JobOfferService) *JobOfferController {
	return &JobOfferController{service: service}
}

func (c *JobOfferController) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateJobOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	role := r.Header.Get(headerUserRole)
	userID := r.Header.Get(headerUserID)

	offer := models.JobOffer{
		CreatorName:    req.CreatorName,
		Title:          req.Title,
		Description:    req.Description,
		Location:       req.Location,
		Modality:       req.Modality,
		JobType:        req.JobType,
		Salary:         req.Salary,
		Schedule:       req.Schedule,
		Requirements:   req.Requirements,
		Deadline:       req.Deadline,
	}

	if role == "role_company" {
		offer.CompanyID = userID
	}

	r = withToken(r)
	if err := c.service.CreateOffer(r.Context(), &offer, role, userID); err != nil {
		if errors.Is(err, services.ErrOfferForbidden) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c.record(r, "crear", offer.ID)

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(offer)
}

func (c *JobOfferController) List(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerUserRole)
	companyID := ""
	if role == "role_company" {
		companyID = r.Header.Get(headerUserID)
	}

	// ?q= busca en título, empresa y ubicación; ?empresa= deja solo las
	// ofertas de esa empresa (HU-45, HU-46).
	q := r.URL.Query().Get("q")
	empresa := r.URL.Query().Get("empresa")
	if len([]rune(q)) > services.MaxBusquedaOfertas || len([]rune(empresa)) > services.MaxBusquedaOfertas {
		http.Error(w, "search too long", http.StatusBadRequest)
		return
	}

	offers, err := c.service.ListOffers(r.Context(), companyID, role)
	if err != nil {
		http.Error(w, "failed to list offers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(services.FiltrarOfertas(offers, q, empresa))
}

func (c *JobOfferController) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdateJobOfferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	offer := models.JobOffer{
		Title:          req.Title,
		Description:    req.Description,
		Location:       req.Location,
		Modality:       req.Modality,
		JobType:        req.JobType,
		Salary:         req.Salary,
		Schedule:       req.Schedule,
		Requirements:   req.Requirements,
		Deadline:       req.Deadline,
		Status:         req.Status,
	}


	if err := c.service.UpdateOffer(r.Context(), id, &offer, r.Header.Get(headerUserRole), r.Header.Get(headerUserID)); err != nil {
		writeOfferError(w, err, "failed to update offer")
		return
	}
	c.record(r, "actualizar", id)

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(offer)
}

func (c *JobOfferController) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := c.service.DeleteOffer(r.Context(), id, r.Header.Get(headerUserRole), r.Header.Get(headerUserID)); err != nil {
		writeOfferError(w, err, "failed to delete offer")
		return
	}
	c.record(r, "baja", id)

	w.WriteHeader(http.StatusNoContent)
}

func (c *JobOfferController) Approve(w http.ResponseWriter, r *http.Request) {
	// Aprobar y rechazar es solo del administrador (HU-41); el voluntario
	// consulta ofertas sin poder aprobarlas (HU-06).
	role := r.Header.Get(headerUserRole)
	if role != "role_admin" && role != "role_dev" {
		http.Error(w, "unauthorized", http.StatusForbidden)
		return
	}
	
	id := chi.URLParam(r, "id")
	userID := r.Header.Get(headerUserID)
	
	r = withToken(r)
	if err := c.service.ApproveOffer(r.Context(), id, userID); err != nil {
		http.Error(w, "failed to approve offer", http.StatusInternalServerError)
		return
	}
	c.record(r, "aprobar", id)
	
	w.WriteHeader(http.StatusOK)
}

func (c *JobOfferController) Reject(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerUserRole)
	if role != "role_admin" && role != "role_dev" {
		http.Error(w, "unauthorized", http.StatusForbidden)
		return
	}
	
	id := chi.URLParam(r, "id")
	userID := r.Header.Get(headerUserID)
	
	if err := c.service.RejectOffer(r.Context(), id, userID); err != nil {
		http.Error(w, "failed to reject offer", http.StatusInternalServerError)
		return
	}
	c.record(r, "rechazar", id)
	
	w.WriteHeader(http.StatusOK)
}

// writeOfferError: 403 sin permisos, 404 si la oferta no existe y 500 sin
// exponer el detalle para lo demás.
func writeOfferError(w http.ResponseWriter, err error, msg string) {
	switch {
	case errors.Is(err, services.ErrOfferForbidden):
		http.Error(w, "forbidden", http.StatusForbidden)
	case errors.Is(err, services.ErrOfferInvalid):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, repositories.ErrOfferNotFound):
		http.Error(w, "offer not found", http.StatusNotFound)
	default:
		slog.Error(msg, slog.String("error", err.Error()))
		http.Error(w, msg, http.StatusInternalServerError)
	}
}

// Auditoria - GET /api/v1/labor-integration/auditoria?desde=&hasta=&accion=
// Solo administración (HU-42), solo lectura.
func (c *JobOfferController) Auditoria(w http.ResponseWriter, r *http.Request) {
	role := r.Header.Get(headerUserRole)
	if role != "role_admin" && role != "role_dev" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if c.audit == nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
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
	rows, err := c.audit.List(r.Context(), f)
	if err != nil {
		slog.Error("auditoría: lectura fallida", slog.String("error", err.Error()))
		http.Error(w, "no se pudo leer la auditoría", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(rows)
}
