package services

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"buenvivir/human-promotion/internal/models"
)

type fakeFichasPDF struct{ FichaService }

func (fakeFichasPDF) GetDetalle(_ context.Context, id string) (*models.FichaDetalle, error) {
	return &models.FichaDetalle{
		FichaProcesoPersonal: models.FichaProcesoPersonal{ID: id, SituacionActual: "Situación", Avances: "Avances"},
		Participante:         &models.DatosParticipante{Nombre: "Ana", Apellido: "Gómez"},
	}, nil
}

type fakeHitosPDF struct{ HitoService }

func (fakeHitosPDF) ListDetalle(context.Context, string) ([]models.HitoDetalle, error) {
	return []models.HitoDetalle{{Hito: models.Hito{Fecha: "2026-10-01", Descripcion: "Primer hito"}}}, nil
}

type fakeActividadPDF struct {
	fichaPedida string
	entradas    []models.EntradaActividad
	err         error
}

func (f *fakeActividadPDF) DeParticipante(context.Context, string) ([]models.EntradaActividad, error) {
	return nil, nil
}

func (f *fakeActividadPDF) DeFicha(_ context.Context, fichaID string) ([]models.EntradaActividad, error) {
	f.fichaPedida = fichaID
	return f.entradas, f.err
}

func nuevoPDF(act ...ActividadService) *pdfService {
	s := NewPDFService(fakeFichasPDF{}, fakeHitosPDF{}, act...).(*pdfService)
	s.now = func() time.Time { return time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC) }
	return s
}

func TestFichaPDF_IncluyeLaActividadEnElCentro(t *testing.T) {
	act := &fakeActividadPDF{entradas: []models.EntradaActividad{
		{ID: "e1", Tipo: "inscripcion", Fecha: "2026-09-10", Descripcion: "Se inscribió en un curso"},
		{ID: "asistencia-2026-10", Tipo: "asistencia_general", Fecha: "2026-10-03", Descripcion: "Asistió al centro en octubre de 2026"},
	}}
	conActividad, err := nuevoPDF(act).FichaPDF(context.Background(), "ficha-1")
	if err != nil {
		t.Fatal(err)
	}
	if act.fichaPedida != "ficha-1" {
		t.Errorf("pidió la actividad de %q, quiero ficha-1", act.fichaPedida)
	}
	sinActividad, err := nuevoPDF().FichaPDF(context.Background(), "ficha-1")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(conActividad, []byte("%PDF")) {
		t.Fatal("no es un PDF")
	}
	if len(conActividad) <= len(sinActividad) {
		t.Errorf("el PDF con actividad (%d bytes) tendría que ser más largo que sin ella (%d)", len(conActividad), len(sinActividad))
	}
}

func TestFichaPDF_SaleIgualSiLaActividadFalla(t *testing.T) {
	act := &fakeActividadPDF{err: errors.New("firestore caído")}
	out, err := nuevoPDF(act).FichaPDF(context.Background(), "ficha-1")
	if err != nil {
		t.Fatalf("el PDF no tiene que fallar por la actividad: %v", err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF")) {
		t.Fatal("no es un PDF")
	}
}
