package controllers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"buenvivir/rxresume/internal/models"
	"buenvivir/rxresume/internal/services"
)

// servicioPublico implementa solo lo que usan las rutas públicas.
type servicioPublico struct {
	services.ResumeService
	err error
}

func (s servicioPublico) GetPublicResume(context.Context, string) (*models.Resume, error) {
	return nil, s.err
}
func (s servicioPublico) ExportPDF(context.Context, string) ([]byte, error) { return nil, s.err }

func pedirPublico(t *testing.T, err error, ruta string, h func(*ResumeController) http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	c := NewResumeController(servicioPublico{err: err})
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/rxresume/public/{id}/pdf", c.DownloadPublicPDF)
	mux.HandleFunc("GET /api/v1/rxresume/public/{id}", c.GetPublicResume)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ruta, nil))
	return rec
}

func TestRutasPublicas_NoExistenteDa404(t *testing.T) {
	for _, ruta := range []string{"/api/v1/rxresume/public/id", "/api/v1/rxresume/public/id/pdf"} {
		rec := pedirPublico(t, services.ErrResumeNotFound, ruta, nil)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s: esperaba 404, fue %d", ruta, rec.Code)
		}
	}
}

func TestRutasPublicas_ErrorInternoNoSeMuestra(t *testing.T) {
	interno := errors.New("failed to connect to `host=localhost user=postgres`")
	for _, ruta := range []string{"/api/v1/rxresume/public/id", "/api/v1/rxresume/public/id/pdf"} {
		rec := pedirPublico(t, interno, ruta, nil)
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("%s: esperaba 500, fue %d", ruta, rec.Code)
		}
		if strings.Contains(rec.Body.String(), "postgres") || strings.Contains(rec.Body.String(), "localhost") {
			t.Fatalf("%s: el cuerpo expone el error interno: %q", ruta, rec.Body.String())
		}
	}
}
