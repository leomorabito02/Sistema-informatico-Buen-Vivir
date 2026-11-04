package routes

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

func TestParseOrigins(t *testing.T) {
	got := ParseOrigins(" http://localhost:4200 , ,https://testing.ceibuenvivir.pages.dev,")
	want := []string{"http://localhost:4200", origenTesting}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseOrigins = %v, want %v", got, want)
	}
	if ParseOrigins("") != nil {
		t.Fatal("una lista vacía tiene que dar nil")
	}
}

func TestWithCORSOrigenPermitido(t *testing.T) {
	h := WithCORS(handlerOK(), []string{origenTesting})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/areas", nil)
	req.Header.Set("Origin", origenTesting)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != origenTesting {
		t.Fatalf("Allow-Origin = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("Allow-Credentials = %q", got)
	}
}

func TestWithCORSOrigenNoPermitido(t *testing.T) {
	h := WithCORS(handlerOK(), []string{origenTesting})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/areas", nil)
	req.Header.Set("Origin", "https://otro-sitio.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (el handler responde igual)", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("un origen no permitido no tiene que recibir Allow-Origin, vino %q", got)
	}
}

func TestWithCORSPreflight(t *testing.T) {
	llamado := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { llamado = true })
	h := WithCORS(next, []string{origenTesting})
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/dashboard/reportes", nil)
	req.Header.Set("Origin", origenTesting)
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "x-skip-global-error")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if llamado {
		t.Fatal("el preflight no tiene que llegar a las rutas (darían 405)")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got != corsHeaders {
		t.Fatalf("Allow-Headers = %q", got)
	}
}

func TestWithCORSSinOrigenesNoHaceNada(t *testing.T) {
	next := handlerOK()
	if h := WithCORS(next, nil); reflect.ValueOf(h).Pointer() != reflect.ValueOf(next).Pointer() {
		t.Fatal("sin orígenes tiene que devolver el mismo handler")
	}
}
