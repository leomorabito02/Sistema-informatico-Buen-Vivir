// Sonda de seguridad contra el stack Docker LOCAL (auditoría 2026-11-11,
// Problemas 2 y 4). Solo hace pedidos que NO modifican datos: lecturas,
// pedidos que deben ser rechazados por autenticación o por rol antes de
// tocar la base, y operaciones sobre ids inexistentes.
//
// Uso (desde la raíz del repo, con el stack levantado):
//
//	JWT_SECRET=... go run ./tests/seguridad/sonda > tests/seguridad/sonda/resultado.json
//
// JWT_SECRET es el mismo de los servicios (por ejemplo, el de
// education/.env). Sin JWT_SECRET se omiten los casos con tokens de rol.
// El secreto no se imprime en ningún lado.
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	gw  = "http://localhost:8080"
	edu = "http://localhost:8210"
	ph  = "http://localhost:8085"
	lab = "http://localhost:8310"
)

type caso struct {
	Grupo     string `json:"grupo"`
	Nombre    string `json:"nombre"`
	Metodo    string `json:"metodo"`
	URL       string `json:"url"`
	Espera    []int  `json:"espera"`
	Obtuvo    int    `json:"obtuvo"`
	Resultado string `json:"resultado"` // OK | FALLA | ERROR
	Nota      string `json:"nota,omitempty"`

	token   string
	headers map[string]string
	body    string
}

var enc = base64.RawURLEncoding

func firmar(secret string, claims map[string]any) string {
	h := enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	p, _ := json.Marshal(claims)
	body := h + "." + enc.EncodeToString(p)
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(body))
	return body + "." + enc.EncodeToString(m.Sum(nil))
}

func sinFirma(claims map[string]any) string {
	h := enc.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	p, _ := json.Marshal(claims)
	return h + "." + enc.EncodeToString(p) + "."
}

func claims(rol string, externo bool, tipo string, exp time.Duration) map[string]any {
	return map[string]any{
		"sub": "audit-" + strings.TrimPrefix(rol, "role_"), "role": rol, "is_external": externo,
		"type": tipo, "exp": time.Now().Add(exp).Unix(), "iat": time.Now().Unix(),
	}
}

func main() {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	tok := func(rol string) string {
		if secret == "" {
			return ""
		}
		return firmar(secret, claims(rol, false, "access", time.Hour))
	}
	adminFalso := sinFirma(claims("role_admin", false, "access", time.Hour))
	otraClave := firmar("una-clave-cualquiera-que-no-es-la-real-123456", claims("role_admin", false, "access", time.Hour))

	casos := []caso{
		// --- Autenticación: tokens inválidos ---
		{Grupo: "autenticacion", Nombre: "education sin token", Metodo: "GET", URL: edu + "/api/cursos", Espera: []int{401}},
		{Grupo: "autenticacion", Nombre: "education token alg=none (admin)", Metodo: "GET", URL: edu + "/api/cursos/borradores", Espera: []int{401}, token: adminFalso},
		{Grupo: "autenticacion", Nombre: "education token firmado con otra clave", Metodo: "GET", URL: edu + "/api/cursos/borradores", Espera: []int{401}, token: otraClave},
		{Grupo: "autenticacion", Nombre: "education token basura", Metodo: "GET", URL: edu + "/api/cursos", Espera: []int{401}, token: "a.b.c"},
		{Grupo: "autenticacion", Nombre: "human-promotion sin token", Metodo: "GET", URL: ph + "/api/v1/human-promotion/fichas?participante_id=x", Espera: []int{401}},
		{Grupo: "autenticacion", Nombre: "human-promotion token alg=none", Metodo: "GET", URL: ph + "/api/v1/human-promotion/indicadores", Espera: []int{401}, token: adminFalso},
		{Grupo: "autenticacion", Nombre: "gateway auth /users sin token", Metodo: "GET", URL: gw + "/api/v1/users", Espera: []int{401}},
		{Grupo: "autenticacion", Nombre: "gateway auth /users token alg=none", Metodo: "GET", URL: gw + "/api/v1/users", Espera: []int{401}, token: adminFalso},
		{Grupo: "autenticacion", Nombre: "gateway campañas gestión sin token", Metodo: "GET", URL: gw + "/api/v1/campaigns", Espera: []int{401}},
		{Grupo: "autenticacion", Nombre: "gateway notificaciones sin token", Metodo: "GET", URL: gw + "/api/v1/notifications", Espera: []int{401}},
		{Grupo: "autenticacion", Nombre: "gateway dashboard sin token", Metodo: "GET", URL: gw + "/api/v1/dashboard/resumen", Espera: []int{401}},

		// --- Inserción laboral: identidad por headers (código que corre en Docker) ---
		{Grupo: "labor-headers", Nombre: "labor directo: aprobar oferta inexistente sin token y X-User-Role admin falsificado",
			Metodo: "PUT", URL: lab + "/api/v1/labor-integration/offers/audit-oferta-inexistente/approve", Espera: []int{401, 403},
			headers: map[string]string{"X-User-ID": "audit-atacante", "X-User-Role": "role_admin"},
			Nota:    "Debe rechazar. Si responde 404/500, el rol falsificado pasó el control (vulnerabilidad)."},
		{Grupo: "labor-headers", Nombre: "gateway: aprobar oferta inexistente sin token y X-User-Role admin falsificado",
			Metodo: "PUT", URL: gw + "/api/v1/labor-integration/offers/audit-oferta-inexistente/approve", Espera: []int{401, 403},
			headers: map[string]string{"X-User-ID": "audit-atacante", "X-User-Role": "role_admin"}},

		// --- CORS ---
		{Grupo: "cors", Nombre: "education preflight desde origen no permitido", Metodo: "OPTIONS", URL: edu + "/api/cursos", Espera: []int{200, 204, 403},
			headers: map[string]string{"Origin": "https://sitio-malicioso.example", "Access-Control-Request-Method": "GET"}},

		// --- Validación de entradas (lecturas públicas) ---
		{Grupo: "entradas", Nombre: "campañas públicas con búsqueda de inyección", Metodo: "GET", URL: gw + "/api/v1/campaigns/publicas?q=%27%20OR%201%3D1--%20%7B%22%24gt%22%3A%22%22%7D", Espera: []int{200, 400}},
		{Grupo: "entradas", Nombre: "campañas públicas con búsqueda de 5000 caracteres", Metodo: "GET", URL: gw + "/api/v1/campaigns/publicas?q=" + strings.Repeat("a", 5000), Espera: []int{400, 414}},
		{Grupo: "entradas", Nombre: "comentarios: listar con entidad inválida", Metodo: "GET", URL: edu + "/api/comentarios?entidadTipo=%3Cscript%3E&entidadId=x", Espera: []int{400}},
	}

	if secret != "" {
		casos = append(casos,
			// --- Autorización por rol (rechazos antes de tocar datos) ---
			caso{Grupo: "roles", Nombre: "education token vencido", Metodo: "GET", URL: edu + "/api/cursos", Espera: []int{401}, token: firmar(secret, claims("role_admin", false, "access", -time.Hour))},
			caso{Grupo: "roles", Nombre: "education token de refresh usado como acceso", Metodo: "GET", URL: edu + "/api/cursos", Espera: []int{401}, token: firmar(secret, claims("role_admin", false, "refresh", time.Hour))},
			caso{Grupo: "roles", Nombre: "participante -> borradores de cursos (admin)", Metodo: "GET", URL: edu + "/api/cursos/borradores", Espera: []int{403}, token: tok("role_participant")},
			caso{Grupo: "roles", Nombre: "participante -> crear curso", Metodo: "POST", URL: edu + "/api/cursos", Espera: []int{403}, token: tok("role_participant"), body: "{}"},
			caso{Grupo: "roles", Nombre: "voluntario -> crear curso (solo admin)", Metodo: "POST", URL: edu + "/api/cursos", Espera: []int{403}, token: tok("role_volunteer"), body: "{}"},
			caso{Grupo: "roles", Nombre: "participante -> auditoría de educación", Metodo: "GET", URL: edu + "/api/auditoria", Espera: []int{403}, token: tok("role_participant")},
			caso{Grupo: "roles", Nombre: "voluntario -> auditoría de educación (solo admin)", Metodo: "GET", URL: edu + "/api/auditoria", Espera: []int{403}, token: tok("role_volunteer")},
			caso{Grupo: "roles", Nombre: "participante -> asistencia general", Metodo: "GET", URL: edu + "/api/asistencia-general?fecha=2026-01-01", Espera: []int{403}, token: tok("role_participant")},
			caso{Grupo: "roles", Nombre: "empresa -> buscador de participantes", Metodo: "GET", URL: edu + "/api/participantes?q=ab", Espera: []int{403}, token: tok("role_company")},
			caso{Grupo: "roles", Nombre: "docente externo -> /api/docente", Metodo: "GET", URL: edu + "/api/docente/cursos", Espera: []int{403}, token: firmar(secret, claims("role_teacher", true, "access", time.Hour))},
			caso{Grupo: "roles", Nombre: "docente -> rutas de participante", Metodo: "GET", URL: edu + "/api/participante/cursos", Espera: []int{403}, token: tok("role_teacher")},
			caso{Grupo: "roles", Nombre: "participante -> borrar comentario (moderación admin)", Metodo: "DELETE", URL: edu + "/api/comentarios/audit-inexistente", Espera: []int{403}, token: tok("role_participant")},
			caso{Grupo: "roles", Nombre: "participante -> fichas de promoción humana", Metodo: "GET", URL: ph + "/api/v1/human-promotion/fichas?participante_id=x", Espera: []int{403}, token: tok("role_participant")},
			caso{Grupo: "roles", Nombre: "docente -> fichas de promoción humana", Metodo: "GET", URL: ph + "/api/v1/human-promotion/fichas?participante_id=x", Espera: []int{403}, token: tok("role_teacher")},
			caso{Grupo: "roles", Nombre: "empresa -> fichas de promoción humana", Metodo: "GET", URL: ph + "/api/v1/human-promotion/fichas?participante_id=x", Espera: []int{403}, token: tok("role_company")},
			caso{Grupo: "roles", Nombre: "voluntario -> indicadores (solo admin)", Metodo: "GET", URL: ph + "/api/v1/human-promotion/indicadores", Espera: []int{403}, token: tok("role_volunteer")},
			caso{Grupo: "roles", Nombre: "voluntario -> auditoría de PH (solo admin)", Metodo: "GET", URL: ph + "/api/v1/human-promotion/auditoria", Espera: []int{403}, token: tok("role_volunteer")},
			caso{Grupo: "roles", Nombre: "empresa -> mi línea de tiempo (solo participante)", Metodo: "GET", URL: ph + "/api/v1/human-promotion/mi-linea-de-tiempo", Espera: []int{403}, token: tok("role_company")},
			caso{Grupo: "roles", Nombre: "participante -> crear campaña (gateway)", Metodo: "POST", URL: gw + "/api/v1/campaigns", Espera: []int{403}, token: tok("role_participant"), body: "{}"},
			caso{Grupo: "roles", Nombre: "empresa -> gestión de campañas (gateway)", Metodo: "GET", URL: gw + "/api/v1/campaigns", Espera: []int{403}, token: tok("role_company")},
			caso{Grupo: "roles", Nombre: "participante -> eliminar campaña (gateway)", Metodo: "DELETE", URL: gw + "/api/v1/campaigns/audit-inexistente", Espera: []int{403, 405},
				token: tok("role_participant"), Nota: "405 = el contenedor corre el código anterior, sin DELETE (Problema 5)."},
			caso{Grupo: "roles", Nombre: "voluntario -> dashboard (solo admin)", Metodo: "GET", URL: gw + "/api/v1/dashboard/resumen", Espera: []int{401, 403}, token: tok("role_volunteer")},
		)
	}

	cli := &http.Client{Timeout: 15 * time.Second}
	for i := range casos {
		c := &casos[i]
		var body io.Reader
		if c.body != "" {
			body = strings.NewReader(c.body)
		}
		req, err := http.NewRequest(c.Metodo, c.URL, body)
		if err != nil {
			c.Resultado, c.Nota = "ERROR", err.Error()
			continue
		}
		if c.body != "" {
			req.Header.Set("Content-Type", "application/json")
		}
		if c.token != "" {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}
		res, err := cli.Do(req)
		if err != nil {
			c.Resultado, c.Nota = "ERROR", err.Error()
			continue
		}
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
		c.Obtuvo = res.StatusCode
		c.Resultado = "FALLA"
		for _, e := range c.Espera {
			if e == res.StatusCode {
				c.Resultado = "OK"
			}
		}
		if c.Grupo == "cors" {
			acao := res.Header.Get("Access-Control-Allow-Origin")
			if acao == "*" || acao == "https://sitio-malicioso.example" {
				c.Resultado, c.Nota = "FALLA", "devuelve Access-Control-Allow-Origin="+acao
			} else {
				c.Nota = "sin Access-Control-Allow-Origin para el origen no permitido"
			}
		}
		// el URL va recortado en el informe (búsqueda de 5000 caracteres)
		if len(c.URL) > 140 {
			c.URL = c.URL[:140] + "…"
		}
		time.Sleep(150 * time.Millisecond) // respeta el límite de 10 pedidos/s del gateway
	}

	// Headers de seguridad en una respuesta del gateway y de education.
	cabeceras := map[string]map[string]string{}
	for nombre, url := range map[string]string{"gateway /health": gw + "/health", "education /api/salud": edu + "/api/salud", "human-promotion /health": ph + "/health"} {
		res, err := cli.Get(url)
		if err != nil {
			continue
		}
		_ = res.Body.Close()
		h := map[string]string{}
		for _, k := range []string{"X-Content-Type-Options", "X-Frame-Options", "Content-Security-Policy", "Strict-Transport-Security", "Referrer-Policy", "Server"} {
			h[k] = res.Header.Get(k)
		}
		cabeceras[nombre] = h
	}

	ok, falla, errores := 0, 0, 0
	for _, c := range casos {
		switch c.Resultado {
		case "OK":
			ok++
		case "FALLA":
			falla++
		default:
			errores++
		}
	}
	out := map[string]any{
		"fecha":          time.Now().Format(time.RFC3339),
		"con_tokens_rol": secret != "",
		"resumen":        map[string]int{"casos": len(casos), "ok": ok, "falla": falla, "error": errores},
		"casos":          casos,
		"cabeceras":      cabeceras,
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	_ = e.Encode(out)
	fmt.Fprintf(os.Stderr, "casos=%d ok=%d falla=%d error=%d\n", len(casos), ok, falla, errores)
}
