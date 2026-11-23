// Security probe for local Docker stack and CD Cloud Run deployments.
// Performs only read-only requests and requests rejected by auth/RBAC/validation
// before touching database state (safe for shared Neon/Firestore environments).
//
// Local usage:
//   JWT_SECRET=... go run ./tests/seguridad/sonda/main.go > tests/seguridad/sonda/resultado.json
//
// CD usage:
//   Set GATEWAY_URL and per-service *_URL / *_ID_TOKEN environment variables.
//   Cloud Run IAM tokens are sent via X-Serverless-Authorization so they never
//   collide with application-level Authorization JWT headers.
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
	pathHealth           = "/health"
	pathAuthUsers        = "/api/v1/users"
	pathCampaigns        = "/api/v1/campaigns"
	pathNotifications    = "/api/v1/notifications"
	pathRxResumeMe       = "/api/v1/rxresume/me"
	pathEduCursos        = "/api/cursos"
	pathEduBorradores    = "/api/cursos/borradores"
	pathHPFichas         = "/api/v1/human-promotion/fichas?participante_id=x"
	pathDashboardResumen = "/api/v1/dashboard/resumen?desde=2026-01-01&hasta=2026-01-31"
	groupHeaderSpoofing  = "header-spoofing"
	bearerPrefix         = "Bearer "
)

type target struct {
	name       string
	baseURL    string
	idToken    string
	healthPath string
}

type probeConfig struct {
	gw      target
	auth    target
	notif   target
	edu     target
	ph      target
	lab     target
	camp    target
	dash    target
	rx      target
	rend    target
	secret  string
	failOn  bool
	isLocal bool
}

type caso struct {
	Grupo     string `json:"grupo"`
	Nombre    string `json:"nombre"`
	Metodo    string `json:"metodo"`
	URL       string `json:"url"`
	Espera    []int  `json:"espera"`
	Obtuvo    int    `json:"obtuvo"`
	Resultado string `json:"resultado"` // OK | FALLA | ERROR
	Nota      string `json:"nota,omitempty"`

	token     string
	idToken   string
	headers   map[string]string
	body      string
	checkCORS bool
}

var enc = base64.RawURLEncoding

func envOrDefault(key, fallback string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return strings.TrimRight(val, "/")
	}
	return fallback
}

func loadConfig() probeConfig {
	isRemote := strings.TrimSpace(os.Getenv("GATEWAY_URL")) != ""
	defaultOrEmpty := func(localURL string) string {
		if isRemote {
			return ""
		}
		return localURL
	}
	return probeConfig{
		gw:      target{name: "gateway", baseURL: envOrDefault("GATEWAY_URL", "http://localhost:8080"), healthPath: pathHealth},
		auth:    target{name: "auth", baseURL: envOrDefault("AUTH_URL", defaultOrEmpty("")), idToken: strings.TrimSpace(os.Getenv("AUTH_ID_TOKEN")), healthPath: pathHealth},
		notif:   target{name: "notifications", baseURL: envOrDefault("NOTIFICATIONS_URL", defaultOrEmpty("")), idToken: strings.TrimSpace(os.Getenv("NOTIFICATIONS_ID_TOKEN")), healthPath: pathHealth},
		edu:     target{name: "education", baseURL: envOrDefault("EDUCATION_URL", defaultOrEmpty("http://localhost:8210")), idToken: strings.TrimSpace(os.Getenv("EDUCATION_ID_TOKEN")), healthPath: "/api/salud"},
		ph:      target{name: "human-promotion", baseURL: envOrDefault("HUMAN_PROMOTION_URL", defaultOrEmpty("http://localhost:8085")), idToken: strings.TrimSpace(os.Getenv("HUMAN_PROMOTION_ID_TOKEN")), healthPath: pathHealth},
		lab:     target{name: "labor-integration", baseURL: envOrDefault("LABOR_URL", defaultOrEmpty("http://localhost:8310")), idToken: strings.TrimSpace(os.Getenv("LABOR_ID_TOKEN")), healthPath: pathHealth},
		camp:    target{name: "campaigns", baseURL: envOrDefault("CAMPAIGNS_URL", defaultOrEmpty("")), idToken: strings.TrimSpace(os.Getenv("CAMPAIGNS_ID_TOKEN")), healthPath: pathHealth},
		dash:    target{name: "dashboard", baseURL: envOrDefault("DASHBOARD_URL", defaultOrEmpty("")), idToken: strings.TrimSpace(os.Getenv("DASHBOARD_ID_TOKEN")), healthPath: pathHealth},
		rx:      target{name: "rxresume", baseURL: envOrDefault("RXRESUME_URL", defaultOrEmpty("")), idToken: strings.TrimSpace(os.Getenv("RXRESUME_ID_TOKEN")), healthPath: pathHealth},
		rend:    target{name: "cv-renderer", baseURL: envOrDefault("CV_RENDERER_URL", defaultOrEmpty("")), idToken: strings.TrimSpace(os.Getenv("CV_RENDERER_ID_TOKEN")), healthPath: pathHealth},
		secret:  strings.TrimSpace(os.Getenv("JWT_SECRET")),
		failOn:  strings.EqualFold(strings.TrimSpace(os.Getenv("FAIL_ON_FINDINGS")), "true"),
		isLocal: !isRemote,
	}
}

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

func addIfConfigured(casos *[]caso, t target, c caso) {
	if t.baseURL == "" {
		return
	}
	c.URL = t.baseURL + c.URL
	c.idToken = t.idToken
	*casos = append(*casos, c)
}

func buildGatewayCases(cfg probeConfig, adminFalso, otraClave string) []caso {
	var casos []caso
	gw := cfg.gw
	addIfConfigured(&casos, gw, caso{Grupo: "autenticacion", Nombre: "gateway auth /users sin token", Metodo: "GET", URL: pathAuthUsers, Espera: []int{401}})
	addIfConfigured(&casos, gw, caso{Grupo: "autenticacion", Nombre: "gateway auth /users token alg=none", Metodo: "GET", URL: pathAuthUsers, Espera: []int{401}, token: adminFalso})
	addIfConfigured(&casos, gw, caso{Grupo: "autenticacion", Nombre: "gateway auth /users token otra clave", Metodo: "GET", URL: pathAuthUsers, Espera: []int{401}, token: otraClave})
	addIfConfigured(&casos, gw, caso{Grupo: "autenticacion", Nombre: "gateway campañas gestión sin token", Metodo: "GET", URL: pathCampaigns, Espera: []int{401}})
	addIfConfigured(&casos, gw, caso{Grupo: "autenticacion", Nombre: "gateway notificaciones sin token", Metodo: "GET", URL: pathNotifications, Espera: []int{401}})
	addIfConfigured(&casos, gw, caso{Grupo: "autenticacion", Nombre: "gateway dashboard sin token", Metodo: "GET", URL: "/api/v1/dashboard/resumen", Espera: []int{401}})
	addIfConfigured(&casos, gw, caso{Grupo: "autenticacion", Nombre: "gateway rxresume sin token", Metodo: "GET", URL: pathRxResumeMe, Espera: []int{401}})
	addIfConfigured(&casos, gw, caso{
		Grupo: "labor-headers", Nombre: "gateway: aprobar oferta inexistente sin token y X-User-Role admin falsificado",
		Metodo: "PUT", URL: "/api/v1/labor-integration/offers/audit-oferta-inexistente/approve", Espera: []int{401, 403},
		headers: map[string]string{"X-User-ID": "audit-atacante", "X-User-Role": "role_admin"},
	})
	addIfConfigured(&casos, gw, caso{Grupo: "entradas", Nombre: "campañas públicas con búsqueda de inyección", Metodo: "GET", URL: "/api/v1/campaigns/publicas?q=%27%20OR%201%3D1--%20%7B%22%24gt%22%3A%22%22%7D", Espera: []int{200, 400}})
	addIfConfigured(&casos, gw, caso{Grupo: "entradas", Nombre: "campañas públicas con búsqueda de 5000 caracteres", Metodo: "GET", URL: "/api/v1/campaigns/publicas?q=" + strings.Repeat("a", 5000), Espera: []int{400, 414}})
	return casos
}

func buildDirectServiceCases(cfg probeConfig, adminFalso, otraClave string) []caso {
	var casos []caso
	spoofHeaders := map[string]string{"X-User-ID": "audit-atacante", "X-User-Role": "role_admin"}

	// Auth
	addIfConfigured(&casos, cfg.auth, caso{Grupo: "autenticacion", Nombre: "auth directo /users sin token", Metodo: "GET", URL: pathAuthUsers, Espera: []int{401}})
	addIfConfigured(&casos, cfg.auth, caso{Grupo: "autenticacion", Nombre: "auth directo /users token alg=none", Metodo: "GET", URL: pathAuthUsers, Espera: []int{401}, token: adminFalso})
	addIfConfigured(&casos, cfg.auth, caso{Grupo: "autenticacion", Nombre: "auth directo /users token otra clave", Metodo: "GET", URL: pathAuthUsers, Espera: []int{401}, token: otraClave})

	// Notifications
	addIfConfigured(&casos, cfg.notif, caso{Grupo: "autenticacion", Nombre: "notifications directo sin token", Metodo: "GET", URL: pathNotifications, Espera: []int{401}})
	addIfConfigured(&casos, cfg.notif, caso{Grupo: "autenticacion", Nombre: "notifications directo token alg=none", Metodo: "GET", URL: pathNotifications, Espera: []int{401}, token: adminFalso})
	addIfConfigured(&casos, cfg.notif, caso{Grupo: groupHeaderSpoofing, Nombre: "notifications directo con X-User-ID/X-User-Role falsificados", Metodo: "GET", URL: pathNotifications, Espera: []int{401}, headers: spoofHeaders})

	// Education
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "autenticacion", Nombre: "education sin token", Metodo: "GET", URL: pathEduCursos, Espera: []int{401}})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "autenticacion", Nombre: "education token alg=none (admin)", Metodo: "GET", URL: pathEduBorradores, Espera: []int{401}, token: adminFalso})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "autenticacion", Nombre: "education token firmado con otra clave", Metodo: "GET", URL: pathEduBorradores, Espera: []int{401}, token: otraClave})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "autenticacion", Nombre: "education token basura", Metodo: "GET", URL: pathEduCursos, Espera: []int{401}, token: "a.b.c"})
	addIfConfigured(&casos, cfg.edu, caso{
		Grupo: "cors", Nombre: "education preflight desde origen no permitido", Metodo: "OPTIONS", URL: pathEduCursos, Espera: []int{200, 204, 403},
		headers: map[string]string{"Origin": "https://sitio-malicioso.example", "Access-Control-Request-Method": "GET"}, checkCORS: true,
	})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "entradas", Nombre: "comentarios: listar con entidad inválida", Metodo: "GET", URL: "/api/comentarios?entidadTipo=%3Cscript%3E&entidadId=x", Espera: []int{400}})

	// Human Promotion
	addIfConfigured(&casos, cfg.ph, caso{Grupo: "autenticacion", Nombre: "human-promotion sin token", Metodo: "GET", URL: pathHPFichas, Espera: []int{401}})
	addIfConfigured(&casos, cfg.ph, caso{Grupo: "autenticacion", Nombre: "human-promotion token alg=none", Metodo: "GET", URL: "/api/v1/human-promotion/indicadores", Espera: []int{401}, token: adminFalso})

	// Labor Integration
	addIfConfigured(&casos, cfg.lab, caso{
		Grupo: "labor-headers", Nombre: "labor directo: aprobar oferta inexistente sin token y X-User-Role admin falsificado",
		Metodo: "PUT", URL: "/api/v1/labor-integration/offers/audit-oferta-inexistente/approve", Espera: []int{401, 403},
		headers: spoofHeaders,
		Nota:    "Debe rechazar. Si responde 404/500, el rol falsificado pasó el control (vulnerabilidad).",
	})
	addIfConfigured(&casos, cfg.lab, caso{Grupo: "autenticacion", Nombre: "labor directo /applications/me sin token", Metodo: "GET", URL: "/api/v1/labor-integration/applications/me", Espera: []int{401}})

	// Campaigns
	addIfConfigured(&casos, cfg.camp, caso{Grupo: "autenticacion", Nombre: "campaigns directo gestión sin token", Metodo: "GET", URL: pathCampaigns, Espera: []int{401}})
	addIfConfigured(&casos, cfg.camp, caso{Grupo: "autenticacion", Nombre: "campaigns directo token alg=none", Metodo: "GET", URL: pathCampaigns, Espera: []int{401}, token: adminFalso})

	// Dashboard
	addIfConfigured(&casos, cfg.dash, caso{Grupo: "autenticacion", Nombre: "dashboard directo sin token", Metodo: "GET", URL: pathDashboardResumen, Espera: []int{401}})
	addIfConfigured(&casos, cfg.dash, caso{Grupo: "autenticacion", Nombre: "dashboard directo token alg=none", Metodo: "GET", URL: pathDashboardResumen, Espera: []int{401}, token: adminFalso})
	addIfConfigured(&casos, cfg.dash, caso{Grupo: groupHeaderSpoofing, Nombre: "dashboard directo con X-User-Role admin falsificado", Metodo: "GET", URL: pathDashboardResumen, Espera: []int{401}, headers: spoofHeaders})

	// RxResume
	addIfConfigured(&casos, cfg.rx, caso{Grupo: "autenticacion", Nombre: "rxresume directo sin token", Metodo: "GET", URL: pathRxResumeMe, Espera: []int{401}})
	addIfConfigured(&casos, cfg.rx, caso{Grupo: "autenticacion", Nombre: "rxresume directo token alg=none", Metodo: "GET", URL: pathRxResumeMe, Espera: []int{401}, token: adminFalso})
	addIfConfigured(&casos, cfg.rx, caso{Grupo: groupHeaderSpoofing, Nombre: "rxresume directo con X-User-ID/X-User-Role falsificados", Metodo: "GET", URL: pathRxResumeMe, Espera: []int{401}, headers: spoofHeaders})

	// CV Renderer (internal service)
	addIfConfigured(&casos, cfg.rend, caso{Grupo: "entradas", Nombre: "cv-renderer POST /render sin campo cv", Metodo: "POST", URL: "/render", Espera: []int{400}, body: "{}"})
	addIfConfigured(&casos, cfg.rend, caso{Grupo: "entradas", Nombre: "cv-renderer ruta inexistente", Metodo: "GET", URL: "/audit-nonexistent", Espera: []int{404}})

	return casos
}

func buildRoleCases(cfg probeConfig) []caso {
	if cfg.secret == "" {
		return nil
	}
	tok := func(rol string) string {
		return firmar(cfg.secret, claims(rol, false, "access", time.Hour))
	}
	var casos []caso

	addIfConfigured(&casos, cfg.edu, caso{Grupo: "roles", Nombre: "education token vencido", Metodo: "GET", URL: pathEduCursos, Espera: []int{401}, token: firmar(cfg.secret, claims("role_admin", false, "access", -time.Hour))})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "roles", Nombre: "education token de refresh usado como acceso", Metodo: "GET", URL: pathEduCursos, Espera: []int{401}, token: firmar(cfg.secret, claims("role_admin", false, "refresh", time.Hour))})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "roles", Nombre: "participante -> borradores de cursos (admin)", Metodo: "GET", URL: pathEduBorradores, Espera: []int{403}, token: tok("role_participant")})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "roles", Nombre: "participante -> crear curso", Metodo: "POST", URL: pathEduCursos, Espera: []int{403}, token: tok("role_participant"), body: "{}"})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "roles", Nombre: "voluntario -> crear curso (solo admin)", Metodo: "POST", URL: pathEduCursos, Espera: []int{403}, token: tok("role_volunteer"), body: "{}"})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "roles", Nombre: "participante -> auditoría de educación", Metodo: "GET", URL: "/api/auditoria", Espera: []int{403}, token: tok("role_participant")})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "roles", Nombre: "voluntario -> auditoría de educación (solo admin)", Metodo: "GET", URL: "/api/auditoria", Espera: []int{403}, token: tok("role_volunteer")})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "roles", Nombre: "participante -> asistencia general", Metodo: "GET", URL: "/api/asistencia-general?fecha=2026-01-01", Espera: []int{403}, token: tok("role_participant")})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "roles", Nombre: "empresa -> buscador de participantes", Metodo: "GET", URL: "/api/participantes?q=ab", Espera: []int{403}, token: tok("role_company")})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "roles", Nombre: "docente externo -> /api/docente", Metodo: "GET", URL: "/api/docente/cursos", Espera: []int{403}, token: firmar(cfg.secret, claims("role_teacher", true, "access", time.Hour))})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "roles", Nombre: "docente -> rutas de participante", Metodo: "GET", URL: "/api/participante/cursos", Espera: []int{403}, token: tok("role_teacher")})
	addIfConfigured(&casos, cfg.edu, caso{Grupo: "roles", Nombre: "participante -> borrar comentario (moderación admin)", Metodo: "DELETE", URL: "/api/comentarios/audit-inexistente", Espera: []int{403}, token: tok("role_participant")})

	addIfConfigured(&casos, cfg.ph, caso{Grupo: "roles", Nombre: "participante -> fichas de promoción humana", Metodo: "GET", URL: pathHPFichas, Espera: []int{403}, token: tok("role_participant")})
	addIfConfigured(&casos, cfg.ph, caso{Grupo: "roles", Nombre: "docente -> fichas de promoción humana", Metodo: "GET", URL: pathHPFichas, Espera: []int{403}, token: tok("role_teacher")})
	addIfConfigured(&casos, cfg.ph, caso{Grupo: "roles", Nombre: "empresa -> fichas de promoción humana", Metodo: "GET", URL: pathHPFichas, Espera: []int{403}, token: tok("role_company")})
	addIfConfigured(&casos, cfg.ph, caso{Grupo: "roles", Nombre: "voluntario -> indicadores (solo admin)", Metodo: "GET", URL: "/api/v1/human-promotion/indicadores", Espera: []int{403}, token: tok("role_volunteer")})
	addIfConfigured(&casos, cfg.ph, caso{Grupo: "roles", Nombre: "voluntario -> auditoría de PH (solo admin)", Metodo: "GET", URL: "/api/v1/human-promotion/auditoria", Espera: []int{403}, token: tok("role_volunteer")})
	addIfConfigured(&casos, cfg.ph, caso{Grupo: "roles", Nombre: "empresa -> mi línea de tiempo (solo participante)", Metodo: "GET", URL: "/api/v1/human-promotion/mi-linea-de-tiempo", Espera: []int{403}, token: tok("role_company")})

	addIfConfigured(&casos, cfg.gw, caso{Grupo: "roles", Nombre: "participante -> crear campaña (gateway)", Metodo: "POST", URL: pathCampaigns, Espera: []int{403}, token: tok("role_participant"), body: "{}"})
	addIfConfigured(&casos, cfg.gw, caso{Grupo: "roles", Nombre: "empresa -> gestión de campañas (gateway)", Metodo: "GET", URL: pathCampaigns, Espera: []int{403}, token: tok("role_company")})
	addIfConfigured(&casos, cfg.gw, caso{Grupo: "roles", Nombre: "participante -> eliminar campaña (gateway)", Metodo: "DELETE", URL: "/api/v1/campaigns/audit-inexistente", Espera: []int{403, 405}, token: tok("role_participant")})
	addIfConfigured(&casos, cfg.gw, caso{Grupo: "roles", Nombre: "voluntario -> dashboard (solo admin)", Metodo: "GET", URL: "/api/v1/dashboard/resumen", Espera: []int{401, 403}, token: tok("role_volunteer")})
	addIfConfigured(&casos, cfg.rx, caso{Grupo: "roles", Nombre: "participante -> listado global rxresume (solo admin/voluntario)", Metodo: "GET", URL: "/api/v1/rxresume/participants", Espera: []int{403}, token: tok("role_participant")})
	addIfConfigured(&casos, cfg.auth, caso{Grupo: "roles", Nombre: "participante -> listado de usuarios auth (solo admin)", Metodo: "GET", URL: pathAuthUsers, Espera: []int{403}, token: tok("role_participant")})

	return casos
}

func runCase(cli *http.Client, c *caso) {
	var body io.Reader
	if c.body != "" {
		body = strings.NewReader(c.body)
	}
	req, err := http.NewRequest(c.Metodo, c.URL, body)
	if err != nil {
		c.Resultado, c.Nota = "ERROR", err.Error()
		return
	}
	if c.body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.idToken != "" {
		req.Header.Set("X-Serverless-Authorization", bearerPrefix+c.idToken)
	}
	if c.token != "" {
		req.Header.Set("Authorization", bearerPrefix+c.token)
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	res, err := cli.Do(req)
	if err != nil {
		c.Resultado, c.Nota = "ERROR", err.Error()
		return
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
	if c.checkCORS {
		acao := res.Header.Get("Access-Control-Allow-Origin")
		if acao == "*" || acao == "https://sitio-malicioso.example" {
			c.Resultado, c.Nota = "FALLA", "devuelve Access-Control-Allow-Origin="+acao
		} else {
			c.Nota = "sin Access-Control-Allow-Origin para el origen no permitido"
		}
	}
	if len(c.URL) > 140 {
		c.URL = c.URL[:140] + "…"
	}
}

func inspectHeaders(cli *http.Client, targets []target) map[string]map[string]string {
	cabeceras := map[string]map[string]string{}
	headerNames := []string{"X-Content-Type-Options", "X-Frame-Options", "Content-Security-Policy", "Strict-Transport-Security", "Referrer-Policy", "Server"}
	for _, t := range targets {
		if t.baseURL == "" {
			continue
		}
		req, err := http.NewRequest(http.MethodGet, t.baseURL+t.healthPath, nil)
		if err != nil {
			continue
		}
		if t.idToken != "" {
			req.Header.Set("X-Serverless-Authorization", bearerPrefix+t.idToken)
		}
		res, err := cli.Do(req)
		if err != nil {
			continue
		}
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
		h := map[string]string{}
		for _, k := range headerNames {
			h[k] = res.Header.Get(k)
		}
		cabeceras[t.name+" "+t.healthPath] = h
	}
	return cabeceras
}

func main() {
	cfg := loadConfig()
	adminFalso := sinFirma(claims("role_admin", false, "access", time.Hour))
	otraClave := firmar("una-clave-cualquiera-que-no-es-la-real-123456", claims("role_admin", false, "access", time.Hour))

	var casos []caso
	casos = append(casos, buildGatewayCases(cfg, adminFalso, otraClave)...)
	casos = append(casos, buildDirectServiceCases(cfg, adminFalso, otraClave)...)
	casos = append(casos, buildRoleCases(cfg)...)

	cli := &http.Client{Timeout: 15 * time.Second}
	for i := range casos {
		runCase(cli, &casos[i])
		time.Sleep(150 * time.Millisecond)
	}

	allTargets := []target{cfg.gw, cfg.auth, cfg.notif, cfg.edu, cfg.ph, cfg.lab, cfg.camp, cfg.dash, cfg.rx, cfg.rend}
	cabeceras := inspectHeaders(cli, allTargets)

	ok, falla, errores := 0, 0, 0
	for _, c := range casos {
		switch c.Resultado {
		case "OK":
			ok++
		case "FALLA":
			falla++
			fmt.Fprintf(os.Stderr, "[FALLA] (%s) %s — %s %s: esperaba %v, obtuvo %d %s\n", c.Grupo, c.Nombre, c.Metodo, c.URL, c.Espera, c.Obtuvo, c.Nota)
		default:
			errores++
			fmt.Fprintf(os.Stderr, "[ERROR] (%s) %s — %s %s: %s\n", c.Grupo, c.Nombre, c.Metodo, c.URL, c.Nota)
		}
	}
	out := map[string]any{
		"fecha":          time.Now().Format(time.RFC3339),
		"con_tokens_rol": cfg.secret != "",
		"resumen":        map[string]int{"casos": len(casos), "ok": ok, "falla": falla, "error": errores},
		"casos":          casos,
		"cabeceras":      cabeceras,
	}
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	_ = e.Encode(out)
	fmt.Fprintf(os.Stderr, "casos=%d ok=%d falla=%d error=%d\n", len(casos), ok, falla, errores)
	if cfg.failOn && (falla > 0 || errores > 0) {
		os.Exit(1)
	}
}
