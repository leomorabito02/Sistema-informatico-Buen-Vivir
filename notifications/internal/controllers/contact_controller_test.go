package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"buenvivir/notifications/internal/services"
)

type emailFalso struct {
	enviados []services.EmailMessage
	err      error
}

func (e *emailFalso) SendEmail(_ context.Context, msg services.EmailMessage) (string, error) {
	if e.err != nil {
		return "", e.err
	}
	e.enviados = append(e.enviados, msg)
	return "id-1", nil
}

const consultaValida = `{"nombre":"Ana Pérez","email":"ana@example.com","asunto":"Cursos y talleres","mensaje":"Quiero saber cuándo empieza el taller de huerta.","acepta_privacidad":true}`

func postContacto(c *ContactController, body, ip string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/notifications/contacto", strings.NewReader(body))
	req.RemoteAddr = ip + ":1234"
	rec := httptest.NewRecorder()
	c.Enviar(rec, req)
	return rec
}

func TestContacto_EnviaEmailAlCentro(t *testing.T) {
	email := &emailFalso{}
	c := NewContactController(email, "")

	rec := postContacto(c, consultaValida, "10.0.0.1")
	if rec.Code != http.StatusAccepted {
		t.Fatalf("esperaba 202, fue %d: %s", rec.Code, rec.Body.String())
	}
	if len(email.enviados) != 1 {
		t.Fatalf("esperaba 1 email, hubo %d", len(email.enviados))
	}
	m := email.enviados[0]
	if m.To[0] != DefaultContactEmail || m.ReplyTo != "ana@example.com" {
		t.Fatalf("destino/reply-to incorrectos: %+v", m)
	}
	if !strings.Contains(m.Subject, "Cursos y talleres") || !strings.Contains(m.Text, "taller de huerta") || !strings.Contains(m.Text, "Ana Pérez") {
		t.Fatalf("contenido incompleto: %q / %q", m.Subject, m.Text)
	}
}

func TestContacto_ValidaCadaCampo(t *testing.T) {
	email := &emailFalso{}
	c := NewContactController(email, "centro@example.com")

	rec := postContacto(c, `{"nombre":"A","email":"no-es-email","asunto":"","mensaje":"corto","acepta_privacidad":false}`, "10.0.0.2")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperaba 400, fue %d", rec.Code)
	}
	var body struct {
		Errores map[string]string `json:"errores"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	for _, campo := range []string{"nombre", "email", "asunto", "mensaje", "acepta_privacidad"} {
		if body.Errores[campo] == "" {
			t.Errorf("falta el error de %s: %v", campo, body.Errores)
		}
	}
	// Inyección de headers en el asunto o el email.
	rec = postContacto(c, `{"nombre":"Ana","email":"ana@example.com\r\nBcc: x@y.com","asunto":"Hola\r\nBcc: x@y.com","mensaje":"Un mensaje de prueba largo","acepta_privacidad":true}`, "10.0.0.2")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("saltos de línea en encabezados: esperaba 400, fue %d", rec.Code)
	}
	if len(email.enviados) != 0 {
		t.Fatal("no debía enviar emails con datos inválidos")
	}
}

func TestContacto_LimitePorIP(t *testing.T) {
	c := NewContactController(&emailFalso{}, "")
	for i := 0; i < contactMaxPorVentana; i++ {
		if rec := postContacto(c, consultaValida, "10.0.0.3"); rec.Code != http.StatusAccepted {
			t.Fatalf("envío %d: %d", i, rec.Code)
		}
	}
	if rec := postContacto(c, consultaValida, "10.0.0.3"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("esperaba 429, fue %d", rec.Code)
	}
	// Otra IP no queda afectada.
	if rec := postContacto(c, consultaValida, "10.0.0.4"); rec.Code != http.StatusAccepted {
		t.Fatalf("otra IP: %d", rec.Code)
	}
}

func TestContacto_SinEmailOFallaDevuelve503(t *testing.T) {
	if rec := postContacto(NewContactController(nil, ""), consultaValida, "10.0.0.5"); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("sin servicio de email: esperaba 503, fue %d", rec.Code)
	}
	if rec := postContacto(NewContactController(&emailFalso{err: errors.New("resend caído")}, ""), consultaValida, "10.0.0.6"); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("Resend con error: esperaba 503, fue %d", rec.Code)
	}
}
