package controllers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/mail"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"buenvivir/notifications/internal/services"
)

// Formulario de "Contacto e información institucional" (pedido del
// 17/11/2026). Antes el formulario del frontend no enviaba nada: solo
// mostraba "enviado". Ahora la consulta se manda por email al centro.
//
// No se guarda en la base: la consulta tiene datos personales (nombre y
// email) y su único destino es la casilla del centro. El email sale con
// Reply-To del remitente para que el centro conteste directo.

const (
	DefaultContactEmail = "ceibuenvivir7@gmail.com"

	contactMaxNombre  = 100
	contactMaxAsunto  = 120
	contactMaxMensaje = 3000
	contactMinMensaje = 10

	// Anti-spam: por IP, como máximo contactMaxPorVentana consultas cada
	// contactVentana.
	contactMaxPorVentana = 5
	contactVentana       = 10 * time.Minute
)

// ContactRequest es el cuerpo de POST /api/v1/notifications/contacto.
type ContactRequest struct {
	Nombre           string `json:"nombre"`
	Email            string `json:"email"`
	Asunto           string `json:"asunto"`
	Mensaje          string `json:"mensaje"`
	AceptaPrivacidad bool   `json:"acepta_privacidad"`
}

// ContactController recibe las consultas públicas y las reenvía por email.
type ContactController struct {
	email   services.EmailService
	destino string

	mu      sync.Mutex
	envios  map[string][]time.Time
	nowFunc func() time.Time
}

// NewContactController arma el controlador. Sin servicio de email
// (RESEND_API_KEY vacío) responde 503 en vez de simular el envío.
func NewContactController(email services.EmailService, destino string) *ContactController {
	destino = strings.TrimSpace(destino)
	if destino == "" {
		destino = DefaultContactEmail
	}
	return &ContactController{email: email, destino: destino, envios: map[string][]time.Time{}, nowFunc: time.Now}
}

type contactFieldErrors map[string]string

// validar devuelve los errores por campo, con el mismo criterio que el
// formulario del frontend.
func (r *ContactRequest) validar() contactFieldErrors {
	r.Nombre = strings.TrimSpace(r.Nombre)
	r.Email = strings.TrimSpace(r.Email)
	r.Asunto = strings.TrimSpace(r.Asunto)
	r.Mensaje = strings.TrimSpace(r.Mensaje)

	errs := contactFieldErrors{}
	if n := utf8.RuneCountInString(r.Nombre); n < 2 || n > contactMaxNombre {
		errs["nombre"] = "Ingresá tu nombre (entre 2 y 100 caracteres)."
	}
	if addr, err := mail.ParseAddress(r.Email); err != nil || addr.Address != r.Email || strings.ContainsAny(r.Email, "\r\n") {
		errs["email"] = "Ingresá un email válido."
	}
	if n := utf8.RuneCountInString(r.Asunto); n == 0 || n > contactMaxAsunto || strings.ContainsAny(r.Asunto, "\r\n") {
		errs["asunto"] = "Elegí el motivo de la consulta."
	}
	if n := utf8.RuneCountInString(r.Mensaje); n < contactMinMensaje || n > contactMaxMensaje {
		errs["mensaje"] = "El mensaje debe tener entre 10 y 3000 caracteres."
	}
	if !r.AceptaPrivacidad {
		errs["acepta_privacidad"] = "Tenés que aceptar la política de privacidad."
	}
	return errs
}

// permitir aplica el límite por IP. Devuelve false si se pasó del cupo.
func (c *ContactController) permitir(ip string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	ahora := c.nowFunc()
	desde := ahora.Add(-contactVentana)
	recientes := c.envios[ip][:0]
	for _, t := range c.envios[ip] {
		if t.After(desde) {
			recientes = append(recientes, t)
		}
	}
	if len(recientes) >= contactMaxPorVentana {
		c.envios[ip] = recientes
		return false
	}
	c.envios[ip] = append(recientes, ahora)
	return true
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// Enviar - POST /api/v1/notifications/contacto (público).
// 202: enviada. 400: campos inválidos ({"errores": {campo: mensaje}}).
// 429: demasiadas consultas seguidas. 503: el email no está configurado o
// falló el envío.
func (c *ContactController) Enviar(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	var req ContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondContact(w, http.StatusBadRequest, map[string]any{"error": "Los datos enviados no son válidos."})
		return
	}
	if errs := req.validar(); len(errs) > 0 {
		respondContact(w, http.StatusBadRequest, map[string]any{"error": "Revisá los campos marcados.", "errores": errs})
		return
	}
	if !c.permitir(clientIP(r)) {
		respondContact(w, http.StatusTooManyRequests, map[string]any{"error": "Enviaste varias consultas seguidas. Probá de nuevo en unos minutos."})
		return
	}
	if c.email == nil {
		slog.Error("consulta de contacto sin servicio de email configurado")
		respondContact(w, http.StatusServiceUnavailable, map[string]any{"error": "No pudimos enviar la consulta. Escribinos a " + c.destino + "."})
		return
	}

	texto := fmt.Sprintf(
		"Nueva consulta desde el formulario de contacto de Buen Vivir.\n\n"+
			"Nombre: %s\nEmail: %s\nMotivo: %s\n\nMensaje:\n%s\n\n"+
			"Respondé este correo para contestarle directamente.",
		req.Nombre, req.Email, req.Asunto, req.Mensaje,
	)
	_, err := c.email.SendEmail(r.Context(), services.EmailMessage{
		To:         []string{c.destino},
		Subject:    "Consulta web: " + req.Asunto,
		Text:       texto,
		ReplyTo:    req.Email,
		SenderType: services.EmailSenderNotification,
		Category:   "contacto",
	})
	if err != nil {
		// No se loguean nombre, email ni mensaje (datos personales).
		slog.Error("no se pudo enviar la consulta de contacto", slog.String("error", err.Error()))
		respondContact(w, http.StatusServiceUnavailable, map[string]any{"error": "No pudimos enviar la consulta. Escribinos a " + c.destino + "."})
		return
	}
	slog.Info("consulta de contacto enviada", slog.String("resultado", "enviada"))
	respondContact(w, http.StatusAccepted, map[string]any{"enviada": true})
}

func respondContact(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
