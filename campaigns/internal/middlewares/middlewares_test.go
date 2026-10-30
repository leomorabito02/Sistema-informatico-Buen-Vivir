package middlewares

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

// jwtFalso acepta solo el token "valido", con los claims indicados.
type jwtFalso struct{ claims jwt.MapClaims }

func (j jwtFalso) ValidateToken(tok string) (jwt.MapClaims, error) {
	if tok != "valido" {
		return nil, errors.New("token inválido")
	}
	return j.claims, nil
}

// eco devuelve el usuario y el rol que quedaron en el contexto.
var eco = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	rol, _ := r.Context().Value(RoleIDKey).(string)
	w.Header().Set("X-Usuario", UserIDFromContext(r.Context()))
	w.Header().Set("X-Rol", rol)
	w.WriteHeader(http.StatusOK)
})

func pedido(h http.Handler, modificar func(*http.Request)) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if modificar != nil {
		modificar(req)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestAuthMiddleware(t *testing.T) {
	h := AuthMiddleware(jwtFalso{jwt.MapClaims{"sub": "u-1", "role": "role_volunteer"}})(eco)

	casos := []struct {
		nombre    string
		modificar func(*http.Request)
		status    int
	}{
		{"sin token", nil, http.StatusUnauthorized},
		{"header sin Bearer", func(r *http.Request) { r.Header.Set("Authorization", "valido") }, http.StatusUnauthorized},
		{"Basic en vez de Bearer", func(r *http.Request) { r.Header.Set("Authorization", "Basic valido") }, http.StatusUnauthorized},
		{"token inválido", func(r *http.Request) { r.Header.Set("Authorization", "Bearer otro") }, http.StatusUnauthorized},
		{"Bearer válido", func(r *http.Request) { r.Header.Set("Authorization", "Bearer valido") }, http.StatusOK},
		{"cookie válida", func(r *http.Request) { r.AddCookie(&http.Cookie{Name: "__Host-access_token", Value: "valido"}) }, http.StatusOK},
		{"cookie gana sobre header", func(r *http.Request) {
			r.AddCookie(&http.Cookie{Name: "__Host-access_token", Value: "valido"})
			r.Header.Set("Authorization", "Bearer otro")
		}, http.StatusOK},
		{"OPTIONS pasa sin token (preflight)", func(r *http.Request) { r.Method = http.MethodOptions }, http.StatusOK},
	}
	for _, tc := range casos {
		t.Run(tc.nombre, func(t *testing.T) {
			if rec := pedido(h, tc.modificar); rec.Code != tc.status {
				t.Fatalf("esperaba %d, obtuve %d", tc.status, rec.Code)
			}
		})
	}

	rec := pedido(h, func(r *http.Request) { r.Header.Set("Authorization", "Bearer valido") })
	if rec.Header().Get("X-Usuario") != "u-1" || rec.Header().Get("X-Rol") != "role_volunteer" {
		t.Fatalf("no dejó usuario y rol en el contexto: %v", rec.Header())
	}
}

func TestRequireRole(t *testing.T) {
	conRol := func(rol string) http.Handler {
		return AuthMiddleware(jwtFalso{jwt.MapClaims{"sub": "u-1", "role": rol}})(RequireRole("role_volunteer")(eco))
	}
	conToken := func(r *http.Request) { r.Header.Set("Authorization", "Bearer valido") }

	for rol, esperado := range map[string]int{
		"role_volunteer":   http.StatusOK,
		"role_admin":       http.StatusOK, // bypass global
		"role_dev":         http.StatusOK, // bypass global
		"role_participant": http.StatusForbidden,
		"role_teacher":     http.StatusForbidden,
		"":                 http.StatusForbidden,
	} {
		if rec := pedido(conRol(rol), conToken); rec.Code != esperado {
			t.Errorf("rol %q: esperaba %d, obtuve %d", rol, esperado, rec.Code)
		}
	}

	// Sin AuthMiddleware antes no hay rol en el contexto: 401, no 403.
	if rec := pedido(RequireRole("role_volunteer")(eco), nil); rec.Code != http.StatusUnauthorized {
		t.Errorf("sin rol en el contexto: esperaba 401, obtuve %d", rec.Code)
	}
}

func TestSetupMiddlewares_CORSYLimiteDePedidos(t *testing.T) {
	r := chi.NewRouter()
	SetupMiddlewares(r, " http://localhost:4200 , ")
	r.Get("/x", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	// Origen permitido: CORS responde con ese origen.
	rec := pedido(r, func(req *http.Request) {
		req.URL.Path = "/x"
		req.Header.Set("Origin", "http://localhost:4200")
	})
	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:4200" {
		t.Errorf("origen permitido sin CORS: %v", rec.Header())
	}
	// Origen no permitido: sin cabecera CORS.
	rec = pedido(r, func(req *http.Request) {
		req.URL.Path = "/x"
		req.Header.Set("Origin", "https://evil.example")
	})
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("origen no permitido con CORS: %v", rec.Header())
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("faltan cabeceras de seguridad: %v", rec.Header())
	}

	// Más de 100 pedidos por minuto desde la misma IP: 429.
	codigo := 0
	for i := 0; i < 101; i++ {
		codigo = pedido(r, func(req *http.Request) { req.URL.Path = "/x" }).Code
	}
	if codigo != http.StatusTooManyRequests {
		t.Errorf("el pedido 101 debía dar 429, dio %d", codigo)
	}
}
