package middlewares

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const origenTesting = "https://testing.ceibuenvivir.pages.dev"

func handlerOK() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func pedirCORS(t *testing.T, env []string, method, origin string, next http.Handler) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, "/api/x", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	CORS(env)(next).ServeHTTP(rec, req)
	return rec
}

func TestParseOrigins(t *testing.T) {
	got := ParseOrigins(" http://localhost:4200 , ,https://otro.example,")
	want := []string{"http://localhost:4200", "https://otro.example"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseOrigins = %v, want %v", got, want)
	}
	if ParseOrigins("") != nil {
		t.Fatal("una lista vacía tiene que dar nil")
	}
}

func TestCORSOrigenDelEntorno(t *testing.T) {
	rec := pedirCORS(t, []string{"https://otro.example"}, http.MethodGet, "https://otro.example", handlerOK())
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://otro.example" {
		t.Fatalf("Allow-Origin = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("Allow-Credentials = %q", got)
	}
}

// El caso de testing: el secret no tiene pages.dev y el front igual tiene
// que poder leer las respuestas.
func TestCORSListaFijaSinSecret(t *testing.T) {
	for _, o := range []string{origenTesting, "https://ceibuenvivir.org", "https://testing.ceibuenvivir.org", "http://localhost:4200"} {
		rec := pedirCORS(t, nil, http.MethodGet, o, handlerOK())
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != o {
			t.Errorf("%s: Allow-Origin = %q", o, got)
		}
	}
}

func TestCORSOrigenNoPermitido(t *testing.T) {
	for _, o := range []string{"https://atacante.example", "https://ceibuenvivir.pages.dev.atacante.example", "https://atacanteceibuenvivir.org"} {
		rec := pedirCORS(t, []string{"https://otro.example"}, http.MethodGet, o, handlerOK())
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, el handler responde igual", o, rec.Code)
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("%s: no tiene que recibir Allow-Origin, vino %q", o, got)
		}
	}
}

// El API Gateway saca Access-Control-Request-Method: el OPTIONS llega sin
// él y tiene que contestarse igual, sin pasar por las rutas.
func TestCORSPreflightSinRequestMethod(t *testing.T) {
	llamado := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { llamado = true })
	rec := pedirCORS(t, nil, http.MethodOptions, origenTesting, next)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if llamado {
		t.Fatal("el preflight no tiene que llegar a las rutas (darían 405)")
	}
	h := rec.Header()
	if h.Get("Access-Control-Allow-Origin") != origenTesting || h.Get("Access-Control-Allow-Headers") != corsHeaders || h.Get("Access-Control-Allow-Methods") != corsMethods {
		t.Fatalf("faltan headers del preflight: %v", h)
	}
}

func TestCORSPreflightOrigenNoPermitido(t *testing.T) {
	rec := pedirCORS(t, nil, http.MethodOptions, "https://atacante.example", handlerOK())
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("un origen no permitido no tiene que recibir Allow-Origin, vino %q", got)
	}
}
