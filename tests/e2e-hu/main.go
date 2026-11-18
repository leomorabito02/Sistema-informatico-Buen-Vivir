// Prueba de punta a punta de las HU implementadas el 15/11/2026 contra el
// stack Docker local (gateway :8080, auth :8082, labor :8310).
//
// OJO: el stack local usa Neon y Firestore compartidos con testing. Esta
// prueba ESCRIBE datos marcados "[PRUEBA]" (una cuenta de participante y
// una de voluntario, una oferta) y los da de baja al final; la auditoría
// no se puede borrar (es inmutable por diseño). Además bloquea 15 minutos
// la cuenta bv.prueba.empresa@gmail.com (prueba de HU-02). Correr solo
// con autorización.
//
// Uso (desde la raíz del repo):
//
//	go run ./tests/e2e-hu            (lee JWT_SECRET de education/.env y
//	                                  DATABASE_URL_AUTH de auth/.env)
//
// Escribe docs/evidencias/e2e-hu-resultado.json. No imprime emails ni
// datos personales.
package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	gw    = "http://localhost:8080"
	auth  = "http://localhost:8082"
	labor = "http://localhost:8310"

	emailEmpresa      = "bv.prueba.empresa@gmail.com"
	emailParticipante = "bv.prueba.participante@gmail.com"
	partPaula         = "part-cxRqHfK5Wb"
)

type resultado struct {
	HU      string `json:"hu"`
	Prueba  string `json:"prueba"`
	OK      bool   `json:"ok"`
	Detalle string `json:"detalle"`
}

var (
	resultados []resultado
	secret     string
	client     = &http.Client{Timeout: 30 * time.Second}
)

func registrar(hu, prueba string, ok bool, detalle string, args ...any) {
	d := fmt.Sprintf(detalle, args...)
	resultados = append(resultados, resultado{hu, prueba, ok, d})
	marca := "OK  "
	if !ok {
		marca = "FALLA"
	}
	fmt.Printf("[%s] %-6s %s — %s\n", marca, hu, prueba, d)
}

func leerEnv(path, clave string) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, l := range strings.Split(string(raw), "\n") {
		l = strings.TrimSpace(strings.TrimRight(l, "\r"))
		if strings.HasPrefix(l, clave+"=") {
			v := strings.TrimPrefix(l, clave+"=")
			return strings.Trim(v, `'"`)
		}
	}
	return ""
}

func token(sub, rol string) string {
	enc := base64.RawURLEncoding
	h := enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	p, _ := json.Marshal(map[string]any{"sub": sub, "role": rol, "is_external": false, "type": "access",
		"exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix()})
	msg := h + "." + enc.EncodeToString(p)
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(msg))
	return msg + "." + enc.EncodeToString(m.Sum(nil))
}

// pedir hace el pedido y devuelve status y cuerpo.
func pedir(method, url, tok string, body any) (int, []byte) {
	var r io.Reader
	ct := ""
	switch b := body.(type) {
	case nil:
	case *bytes.Buffer:
		r = b
	case [2]any: // multipart: {buffer, content-type}
		r, ct = b[0].(*bytes.Buffer), b[1].(string)
	default:
		raw, _ := json.Marshal(b)
		r, ct = bytes.NewReader(raw), "application/json"
	}
	req, _ := http.NewRequest(method, url, r)
	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, []byte(err.Error())
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, out
}

func data(raw []byte) json.RawMessage {
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if json.Unmarshal(raw, &env) == nil && len(env.Data) > 0 {
		return env.Data
	}
	return raw
}

func corto(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 160 {
		s = s[:160] + "…"
	}
	return s
}

func main() {
	secret = leerEnv("education/.env", "JWT_SECRET")
	if secret == "" {
		fmt.Println("falta JWT_SECRET en education/.env")
		os.Exit(1)
	}
	rand.Seed(time.Now().UnixNano())
	sufijo := fmt.Sprintf("%d", time.Now().Unix()%1000000)
	hoy := time.Now().Format("2006-01-02")

	adminTok := token("audit-admin-e2e", "role_admin")

	// --- Descubrir ids de las cuentas de prueba (sin imprimir emails).
	_, raw := pedir("GET", gw+"/api/v1/users", adminTok, nil)
	var users []struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
		RoleID string `json:"role_id"`
	}
	_ = json.Unmarshal(data(raw), &users)
	ids := map[string]string{}
	for _, u := range users {
		ids[strings.ToLower(u.Email)+"|"+u.RoleID] = u.UserID
	}
	// El actor tiene que ser un admin real: approved_by tiene FK a users.
	for _, u := range users {
		if u.RoleID == "role_admin" {
			adminTok = token(u.UserID, "role_admin")
			break
		}
	}
	empresaID := ids[emailEmpresa+"|role_company"]
	paulaID := ids[emailParticipante+"|role_participant"]
	fmt.Printf("cuentas de prueba encontradas: empresa=%v participante=%v (de %d usuarios)\n", empresaID != "", paulaID != "", len(users))

	// ================= HU-10 / HU-06 / HU-03: auditoría de auth
	st, _ := pedir("GET", gw+"/api/v1/auth/auditoria?desde="+hoy+"&hasta="+hoy, adminTok, nil)
	registrar("HU-10", "admin lee la auditoría de auth por el gateway", st == 200, "status %d", st)
	st, _ = pedir("GET", gw+"/api/v1/auth/auditoria", token("audit-vol-e2e", "role_volunteer"), nil)
	registrar("HU-10", "voluntario no puede leer la auditoría", st == 403, "status %d", st)
	st, _ = pedir("GET", gw+"/api/v1/users/participants/history", token("audit-part-e2e", "role_participant"), nil)
	registrar("HU-06", "participante pide un recurso de admin (403, queda auditado)", st == 403, "status %d", st)
	st, _ = pedir("GET", auth+"/api/v1/auth/me", "a.b.c", nil)
	registrar("HU-03", "token inválido directo a auth (como el gateway de GCP)", st == 401, "status %d", st)

	// ================= HU-01 / HU-07 / HU-08 / HU-09: alta, edición y baja de participante
	base := map[string]any{"first_name": "[PRUEBA] E2E", "last_name": "Auditoria", "phone": "3547000000",
		"password": "Prueba123!", "role_id": "role_participant"}
	alta := func(cambios map[string]any) (int, []byte) {
		b := map[string]any{}
		for k, v := range base {
			b[k] = v
		}
		for k, v := range cambios {
			b[k] = v
		}
		return pedir("POST", gw+"/api/v1/users", adminTok, b)
	}
	st, raw = alta(map[string]any{"email": "bv.prueba.e2e" + sufijo + "@hotmail.com", "dni": "30111222", "address": "[PRUEBA] Calle 1"})
	registrar("HU-01", "alta de participante sin Gmail rechazada", st == 400 && strings.Contains(string(raw), "Gmail"), "status %d", st)
	st, raw = alta(map[string]any{"email": "bv.prueba.participante+e2e" + sufijo + "@gmail.com", "address": "[PRUEBA] Calle 1"})
	registrar("HU-07", "alta de participante sin DNI rechazada", st == 400 && strings.Contains(string(raw), "DNI"), "status %d", st)
	dni := fmt.Sprintf("9%07d", rand.Intn(10000000))
	st, raw = alta(map[string]any{"email": "bv.prueba.participante+e2e" + sufijo + "@gmail.com", "dni": dni, "address": "[PRUEBA] Calle 1"})
	var creado struct {
		UserID string `json:"user_id"`
	}
	_ = json.Unmarshal(data(raw), &creado)
	registrar("HU-07", "alta de participante completa", (st == 201 || st == 200) && creado.UserID != "", "status %d", st)
	if creado.UserID != "" {
		_, raw = pedir("GET", gw+"/api/v1/users/"+creado.UserID+"/profile", adminTok, nil)
		ok := strings.Contains(string(raw), dni) && strings.Contains(string(raw), "[PRUEBA] Calle 1")
		registrar("HU-07", "DNI y dirección quedaron guardados en el perfil", ok, "perfil con DNI y dirección: %v", ok)
		st, _ = pedir("PUT", gw+"/api/v1/users/"+creado.UserID, adminTok, map[string]any{"first_name": "[PRUEBA] E2E editado"})
		registrar("HU-08", "edición del participante", st == 200, "status %d", st)
		st, _ = pedir("DELETE", gw+"/api/v1/users/"+creado.UserID, adminTok, nil)
		registrar("HU-09", "baja del participante [PRUEBA]", st == 200 || st == 204, "status %d", st)
	}

	// ================= HU-01 / HU-12: registro público, rechazo con email
	reg := map[string]any{"first_name": "[PRUEBA]", "last_name": "Rechazo", "phone": "3547000000", "password": "Prueba123!"}
	reg["email"], reg["role"] = "bv.prueba.e2e"+sufijo+"@hotmail.com", "participant"
	st, raw = pedir("POST", gw+"/api/v1/auth/register", "", reg)
	registrar("HU-01", "registro público de participante sin Gmail rechazado", st >= 400 && st < 500 && strings.Contains(string(raw), "Gmail"), "status %d", st)
	reg["email"], reg["role"] = "bv.prueba.voluntario+rej"+sufijo+"@gmail.com", "volunteer"
	st, raw = pedir("POST", gw+"/api/v1/auth/register", "", reg)
	var regResp struct {
		UserID string `json:"user_id"`
	}
	_ = json.Unmarshal(data(raw), &regResp)
	rechazadoID := regResp.UserID
	registrar("HU-12", "registro público de voluntario [PRUEBA] (queda pendiente)", (st == 200 || st == 201) && rechazadoID != "", "status %d", st)
	if rechazadoID != "" {
		st, raw = pedir("PATCH", gw+"/api/v1/users/"+rechazadoID+"/approval", adminTok, map[string]any{"approval_status": "rejected"})
		registrar("HU-12", "rechazo por PATCH /users/{id}/approval", st == 200 && strings.Contains(string(raw), "rejected"), "status %d", st)
		_, raw = pedir("GET", gw+"/api/v1/users/"+rechazadoID+"/profile", adminTok, nil)
		registrar("HU-12", "el perfil quedó rechazado", strings.Contains(string(raw), `"rejected"`), "approval_status rejected: %v", strings.Contains(string(raw), `"rejected"`))
	}

	// ================= HU-11: límite de 5 MB
	var mp bytes.Buffer
	w := multipart.NewWriter(&mp)
	fw, _ := w.CreateFormFile("file", "e2e.pdf")
	_, _ = fw.Write(append([]byte("%PDF-1.4\n"), bytes.Repeat([]byte("a"), 6<<20)...))
	_ = w.WriteField("name", "[PRUEBA] grande")
	_ = w.Close()
	st, raw = pedir("POST", gw+"/api/v1/participants/"+partPaula+"/documents", adminTok, [2]any{&mp, w.FormDataContentType()})
	registrar("HU-11", "documento de 6 MB rechazado", st == 400 && strings.Contains(string(raw), "5 MB"), "status %d %s", st, corto(raw))

	// ================= HU-02: bloqueo tras 5 intentos (bloquea 15 min la cuenta empresa de prueba)
	quedan := 0
	var ultimo int
	var ultimoRaw []byte
	for i := 1; i <= 5; i++ {
		ultimo, ultimoRaw = pedir("POST", gw+"/api/v1/auth/check-credentials", "", map[string]any{"email": emailEmpresa, "password": "Incorrecta-e2e-1!"})
		if ultimo == 401 && strings.Contains(string(ultimoRaw), "Te quedan") {
			quedan++
		}
	}
	registrar("HU-02", "intentos 1 a 4 avisan cuántos quedan", quedan == 4, "%d de 4 con aviso", quedan)
	registrar("HU-02", "el 5.º intento bloquea (423 ACCOUNT_LOCKED)", ultimo == 423 && strings.Contains(string(ultimoRaw), "ACCOUNT_LOCKED"), "status %d %s", ultimo, corto(ultimoRaw))
	st, raw = pedir("POST", gw+"/api/v1/auth/check-credentials", "", map[string]any{"email": emailEmpresa, "password": "otra-cualquiera"})
	registrar("HU-02", "sigue bloqueada en el intento siguiente", st == 423, "status %d", st)

	// ================= HU-39 / HU-44 / HU-41 / HU-42: ofertas
	var skills []struct {
		Nombre string `json:"nombre"`
	}
	_, raw = pedir("GET", gw+"/api/participants/"+partPaula+"/habilidades", adminTok, nil)
	_ = json.Unmarshal(raw, &skills)
	// Solo se cruza contra habilidades de la cuenta de prueba: con las de
	// otros participantes, Notificaciones les mandaría un email real.
	requisito := "[PRUEBA] requisito sin coincidencias"
	if len(skills) > 0 {
		requisito = skills[0].Nombre
	}

	empTok := token(empresaID, "role_company")
	oferta := map[string]any{"titulo": "[PRUEBA] Oferta E2E " + sufijo, "descripcion": "[PRUEBA] oferta de prueba automática",
		"ubicacion": "Alta Gracia", "modalidad": "Presencial", "tipo_empleo": "Medio tiempo", "horario": "9 a 13",
		"requisitos": []string{strings.ToUpper(requisito)}, "fecha_limite": time.Now().AddDate(0, 1, 0).Format(time.RFC3339)}
	st, raw = pedir("POST", gw+"/api/v1/labor-integration/offers", empTok, oferta)
	registrar("HU-39", "oferta sin salario rechazada con el campo", st == 400 && strings.Contains(string(raw), "salario"), "status %d %s", st, corto(raw))
	oferta["salario"] = "A convenir"
	st, raw = pedir("POST", gw+"/api/v1/labor-integration/offers", empTok, oferta)
	var of struct {
		ID     string `json:"id"`
		Estado string `json:"estado"`
	}
	_ = json.Unmarshal(raw, &of)
	registrar("HU-39", "oferta completa de la empresa queda pendiente", st == 201 && of.Estado == "Pendiente", "status %d estado %q", st, of.Estado)
	if of.ID != "" {
		st, _ = pedir("PUT", gw+"/api/v1/labor-integration/offers/"+of.ID+"/approve", adminTok, nil)
		registrar("HU-41", "admin aprueba la oferta", st == 200, "status %d", st)
		time.Sleep(3 * time.Second)
		logs, _ := exec.Command("docker", "logs", "--since", "2m", "sistema-informatico-buen-vivir-labor-integration-1").CombinedOutput()
		linea := ""
		for _, l := range strings.Split(string(logs), "\n") {
			if strings.Contains(l, "cruce hecho") && strings.Contains(l, of.ID) {
				linea = l[strings.Index(l, "participantes="):]
			}
		}
		registrar("HU-44", "al aprobar, labor consultó a Educación y cruzó requisitos con habilidades", linea != "", "%s", linea)
		if len(skills) > 0 {
			time.Sleep(12 * time.Second)
			_, raw = pedir("GET", gw+"/api/v1/notifications?limit=50", token(paulaID, "role_participant"), nil)
			okComp := strings.Contains(string(raw), "Hay una oferta laboral para vos")
			registrar("HU-44", "la participante con la habilidad recibió el aviso", okComp, "aviso en la campana: %v", okComp)
		}
		time.Sleep(9 * time.Second)
		st, _ = pedir("DELETE", gw+"/api/v1/labor-integration/offers/"+of.ID, adminTok, nil)
		registrar("HU-41", "admin borra la oferta [PRUEBA]", st == 204, "status %d", st)
		time.Sleep(12 * time.Second)
		_, raw = pedir("GET", gw+"/api/v1/notifications?limit=50", empTok, nil)
		okDel := strings.Contains(string(raw), "Oferta laboral eliminada")
		registrar("HU-41", "la empresa recibió el aviso de oferta eliminada", okDel, "aviso en la campana: %v", okDel)
		_, raw = pedir("GET", gw+"/api/v1/labor-integration/auditoria?desde="+hoy+"&hasta="+hoy, adminTok, nil)
		var aud []struct {
			Accion    string `json:"accion"`
			RecursoID string `json:"recurso_id"`
		}
		_ = json.Unmarshal(raw, &aud)
		acc := map[string]bool{}
		for _, a := range aud {
			if a.RecursoID == of.ID {
				acc[a.Accion] = true
			}
		}
		registrar("HU-42", "auditoría de la oferta: crear, aprobar y baja", acc["crear"] && acc["aprobar"] && acc["baja"], "acciones %v", acc)
		st, _ = pedir("GET", gw+"/api/v1/labor-integration/auditoria", empTok, nil)
		registrar("HU-42", "la empresa no puede leer la auditoría", st == 403, "status %d", st)
	}

	// ================= HU-14 / HU-50
	_, raw = pedir("GET", gw+"/api/materiales", adminTok, nil)
	var mats []struct {
		ID string `json:"id"`
	}
	_ = json.Unmarshal(raw, &mats)
	if len(mats) > 0 {
		st, raw = pedir("GET", gw+"/api/materiales/"+mats[0].ID+"/cursos", adminTok, nil)
		registrar("HU-14", "cursos que usan un material", st == 200 && strings.HasPrefix(strings.TrimSpace(string(raw)), "["), "status %d %s", st, corto(raw))
	} else {
		registrar("HU-14", "cursos que usan un material", false, "no hay materiales para probar")
	}
	st, raw = pedir("GET", gw+"/api/v1/campaigns/publicas?limit=1&offset=0", "", nil)
	var pag struct {
		Items []any `json:"items"`
		Total int   `json:"total"`
	}
	_ = json.Unmarshal(raw, &pag)
	registrar("HU-50", "listado público paginado", st == 200 && len(pag.Items) <= 1 && pag.Total >= len(pag.Items), "status %d items %d total %d", st, len(pag.Items), pag.Total)

	// ================= HU-10: lo de arriba quedó en la auditoría de auth
	_, raw = pedir("GET", gw+"/api/v1/auth/auditoria?desde="+hoy+"&hasta="+hoy, adminTok, nil)
	var reg2 []struct {
		Accion    string `json:"accion"`
		RecursoID string `json:"recurso_id"`
		Campos    string `json:"campos"`
	}
	_ = json.Unmarshal(raw, &reg2)
	tiene := func(accion, id, campo string) bool {
		for _, r := range reg2 {
			if r.Accion == accion && (id == "" || r.RecursoID == id) && (campo == "" || strings.Contains(r.Campos, campo)) {
				return true
			}
		}
		return false
	}
	registrar("HU-07", "auditoría: alta del participante", tiene("crear", creado.UserID, ""), "")
	registrar("HU-08", "auditoría: edición con el nombre del campo (sin valores)", tiene("actualizar", creado.UserID, "first_name"), "")
	registrar("HU-09", "auditoría: baja", tiene("baja", creado.UserID, ""), "")
	registrar("HU-12", "auditoría: rechazo", tiene("rechazar", rechazadoID, ""), "")
	registrar("HU-06", "auditoría: acceso denegado", tiene("acceso_denegado", "", ""), "")
	registrar("HU-03", "auditoría: token rechazado", tiene("token_rechazado", "", ""), "")
	sinValores := !strings.Contains(string(raw), "[PRUEBA] E2E editado") && !strings.Contains(string(raw), dni)
	registrar("HU-10", "la auditoría no guarda valores (ni nombre ni DNI)", sinValores, "")

	// ================= HU-10: inmutabilidad (trigger en Postgres)
	dsn := leerEnv("auth/.env", "DATABASE_URL_AUTH")
	if dsn != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		conn, err := pgx.Connect(ctx, dsn)
		if err == nil {
			var tablas int
			_ = conn.QueryRow(ctx, "SELECT count(*) FROM information_schema.tables WHERE table_name IN ('auditoria','login_attempts')").Scan(&tablas)
			registrar("HU-10", "tablas auditoria y login_attempts creadas en Neon", tablas == 2, "%d de 2", tablas)
			_, errUpd := conn.Exec(ctx, "UPDATE auditoria SET accion = accion WHERE id = (SELECT max(id) FROM auditoria)")
			registrar("HU-10", "UPDATE sobre la auditoría rechazado por el trigger", errUpd != nil && strings.Contains(errUpd.Error(), "no se puede modificar"), "%v", errUpd)
			_, errDel := conn.Exec(ctx, "DELETE FROM auditoria WHERE id = (SELECT max(id) FROM auditoria)")
			registrar("HU-10", "DELETE sobre la auditoría rechazado por el trigger", errDel != nil, "%v", errDel)
			// Limpieza: la cuenta rechazada [PRUEBA] se da de baja (soft delete).
			conn.Close(ctx)
		} else {
			registrar("HU-10", "conexión a Neon para probar el trigger", false, "%v", err)
		}
		cancel()
	}
	for _, extra := range os.Args[1:] { // ids [PRUEBA] que quedaron de corridas anteriores
		st, _ = pedir("DELETE", gw+"/api/v1/users/"+extra, adminTok, nil)
		registrar("limpieza", "baja de una cuenta [PRUEBA] anterior", st == 200 || st == 204, "status %d", st)
	}
	if rechazadoID != "" {
		st, _ = pedir("DELETE", gw+"/api/v1/users/"+rechazadoID, adminTok, nil)
		registrar("limpieza", "baja de la cuenta [PRUEBA] rechazada", st == 200 || st == 204, "status %d", st)
	}

	// --- Resumen
	ok := 0
	for _, r := range resultados {
		if r.OK {
			ok++
		}
	}
	fmt.Printf("\n%d de %d pruebas OK\n", ok, len(resultados))
	out, _ := json.MarshalIndent(map[string]any{"fecha": time.Now().Format(time.RFC3339), "ok": ok, "total": len(resultados), "resultados": resultados}, "", "  ")
	_ = os.WriteFile("docs/evidencias/e2e-hu-resultado.json", out, 0o644)
}
