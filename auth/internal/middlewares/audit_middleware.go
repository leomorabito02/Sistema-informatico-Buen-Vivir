package middlewares

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"auth/internal/models"

	"github.com/go-chi/chi/v5"
)

// AuditWriter es lo único que el middleware necesita del repositorio.
type AuditWriter interface {
	Create(ctx context.Context, entry *models.AuditLog) error
}

// auditActor lo completa AuthMiddleware con el usuario del JWT, para que
// AuditTrail (que corre por fuera) sepa quién hizo el pedido.
type auditActor struct {
	userID string
	role   string
}

type auditActorKey struct{}

func auditActorFrom(ctx context.Context) *auditActor {
	a, _ := ctx.Value(auditActorKey{}).(*auditActor)
	return a
}

// Acciones del log de auditoría de Autenticación.
const (
	AuditCrear          = "crear"
	AuditActualizar     = "actualizar"
	AuditBaja           = "baja"
	AuditAprobar        = "aprobar"
	AuditRechazar       = "rechazar"
	AuditTokenRechazado = "token_rechazado" // HU-03
	AuditAccesoDenegado = "acceso_denegado" // HU-06
)

// camposSensibles no se listan ni por nombre.
var camposSensibles = map[string]bool{"password": true, "new_password": true, "old_password": true, "current_password": true, "token": true, "code": true, "secret": true}

const maxAuditBody = 1 << 20

// AuditTrail registra en el log de auditoría (HU-10, HU-03, HU-06):
//   - cada alta, edición o baja que terminó bien (2xx), con los nombres
//     de los campos enviados (nunca los valores);
//   - cada token presentado y rechazado (401) y cada acceso denegado
//     por rol (403).
//
// Va en el grupo de rutas protegidas, por fuera de AuthMiddleware. El
// registro se escribe antes de terminar el pedido (en Cloud Run la CPU se
// corta al terminar), con un tope de 3 segundos; si falla, queda en el
// log técnico.
func AuditTrail(writer AuditWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if writer == nil || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			mutacion := r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch || r.Method == http.MethodDelete

			var body []byte
			if mutacion && strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") && r.Body != nil {
				b, err := io.ReadAll(io.LimitReader(r.Body, maxAuditBody))
				if err == nil {
					body = b
					r.Body = io.NopCloser(bytes.NewReader(b))
				}
			}

			actor := &auditActor{}
			r = r.WithContext(context.WithValue(r.Context(), auditActorKey{}, actor))
			rec := &auditRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)

			entry := buildAuditEntry(r, rec, actor, mutacion, body)
			if entry == nil {
				return
			}
			ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 3*time.Second)
			defer cancel()
			if err := writer.Create(ctx, entry); err != nil {
				slog.Error("auditoría: no se pudo registrar", slog.String("accion", entry.Accion), slog.String("recurso", entry.Recurso), slog.String("error", err.Error()))
			}
		})
	}
}

func buildAuditEntry(r *http.Request, rec *auditRecorder, actor *auditActor, mutacion bool, body []byte) *models.AuditLog {
	pattern := ""
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		pattern = rctx.RoutePattern()
	}
	entry := &models.AuditLog{
		UsuarioID: actor.userID,
		Rol:       actor.role,
		Recurso:   auditRecurso(pattern, r.URL.Path),
		RecursoID: auditRecursoID(r, rec, actor),
		IP:        clientIP(r),
		Fecha:     time.Now().UTC(),
	}
	switch {
	case rec.status == http.StatusUnauthorized:
		if !tokenPresented(r) {
			return nil // sin sesión: no es un rechazo de token
		}
		entry.Accion, entry.Resultado = AuditTokenRechazado, "denegado"
		entry.Campos = r.Method + " " + firstNonEmpty(pattern, r.URL.Path)
	case rec.status == http.StatusForbidden:
		entry.Accion, entry.Resultado = AuditAccesoDenegado, "denegado"
		entry.Campos = r.Method + " " + firstNonEmpty(pattern, r.URL.Path)
	case mutacion && rec.status >= 200 && rec.status < 300:
		campos, approval := jsonKeys(body)
		entry.Accion, entry.Resultado = auditAccion(r.Method, approval), "ok"
		entry.Campos = strings.Join(campos, ",")
	default:
		return nil
	}
	return entry
}

func auditAccion(method, approval string) string {
	switch method {
	case http.MethodPost:
		return AuditCrear
	case http.MethodDelete:
		return AuditBaja
	}
	switch strings.ToLower(approval) {
	case "approved", "accepted", "aprobado":
		return AuditAprobar
	case "rejected", "rechazado":
		return AuditRechazar
	}
	return AuditActualizar
}

// auditRecurso traduce la ruta a un nombre de recurso legible.
func auditRecurso(pattern, path string) string {
	p := strings.TrimPrefix(firstNonEmpty(pattern, path), "/api/v1/")
	switch {
	case strings.HasPrefix(p, "participants/{id}/documents"):
		return "documento_participante"
	case strings.HasPrefix(p, "participants"):
		return "perfil_participante"
	case strings.HasPrefix(p, "users/me/devices"):
		return "dispositivo"
	case strings.HasPrefix(p, "users"):
		return "usuario"
	case strings.HasPrefix(p, "companies"):
		return "empresa"
	case strings.HasPrefix(p, "teachers"):
		return "docente"
	case strings.HasPrefix(p, "volunteers"):
		return "voluntario"
	case strings.HasPrefix(p, "roles"):
		return "rol"
	case strings.HasPrefix(p, "permissions"):
		return "permiso"
	case strings.HasPrefix(p, "profile/me/password"):
		return "password"
	case strings.HasPrefix(p, "profile/me/2fa"), strings.Contains(p, "google"):
		return "seguridad_cuenta"
	case strings.HasPrefix(p, "profile"):
		return "perfil_propio"
	}
	return strings.SplitN(p, "/", 2)[0]
}

func auditRecursoID(r *http.Request, rec *auditRecorder, actor *auditActor) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		if id := rctx.URLParam("docId"); id != "" {
			return id
		}
		if id := rctx.URLParam("id"); id != "" {
			return id
		}
	}
	if strings.HasPrefix(r.URL.Path, "/api/v1/profile/me") {
		return actor.userID
	}
	// Alta: el id viene en la respuesta.
	var m map[string]any
	if json.Unmarshal(rec.body.Bytes(), &m) != nil {
		return ""
	}
	if d, ok := m["data"].(map[string]any); ok {
		m = d
	}
	for _, k := range []string{"user_id", "participant_id", "document_id", "id"} {
		if v, ok := m[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// jsonKeys devuelve los nombres de campos del cuerpo (sin los sensibles)
// y el valor de approval_status si vino, para distinguir aprobar de
// rechazar. Los valores no se guardan.
func jsonKeys(body []byte) ([]string, string) {
	var m map[string]json.RawMessage
	if len(body) == 0 || json.Unmarshal(body, &m) != nil {
		return nil, ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		if !camposSensibles[strings.ToLower(k)] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var approval string
	for _, k := range []string{"approval_status", "estado_aprobacion"} {
		if raw, ok := m[k]; ok {
			_ = json.Unmarshal(raw, &approval)
		}
	}
	return keys, approval
}

func tokenPresented(r *http.Request) bool {
	if r.Header.Get("Authorization") != "" {
		return true
	}
	for _, n := range []string{"__Host-access_token", "access_token"} {
		if c, err := r.Cookie(n); err == nil && c.Value != "" {
			return true
		}
	}
	return false
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// auditRecorder guarda el status y hasta 64 KB de la respuesta (para
// sacar el id de un alta).
type auditRecorder struct {
	http.ResponseWriter
	status int
	body   bytes.Buffer
}

func (a *auditRecorder) WriteHeader(code int) {
	a.status = code
	a.ResponseWriter.WriteHeader(code)
}

func (a *auditRecorder) Write(b []byte) (int, error) {
	if a.body.Len() < 64<<10 {
		a.body.Write(b)
	}
	return a.ResponseWriter.Write(b)
}

func (a *auditRecorder) Flush() {
	if f, ok := a.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
