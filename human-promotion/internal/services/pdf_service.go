package services

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"

	"buenvivir/human-promotion/internal/models"
)

// Paleta del manual de identidad (mocks): verde #485839 y texto #1D1D1B.
var (
	colorPrimario = [3]int{72, 88, 57}
	colorTexto    = [3]int{29, 29, 27}
	colorSuave    = [3]int{110, 110, 105}
)

// PDFService genera en el backend el PDF de la ficha y su línea de
// hitos (HU-37). Los permisos son los mismos que los de la ficha: los
// aplica la ruta, que es la misma que protege a la ficha.
type PDFService interface {
	FichaPDF(ctx context.Context, fichaID string) ([]byte, error)
}

type pdfService struct {
	fichas FichaService
	hitos  HitoService
	now    func() time.Time
}

func NewPDFService(fichas FichaService, hitos HitoService) PDFService {
	return &pdfService{fichas: fichas, hitos: hitos, now: time.Now}
}

func (s *pdfService) FichaPDF(ctx context.Context, fichaID string) ([]byte, error) {
	ficha, err := s.fichas.GetDetalle(ctx, fichaID)
	if err != nil {
		return nil, err
	}
	hitos, err := s.hitos.ListDetalle(ctx, fichaID)
	if err != nil {
		return nil, err
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	// Las fuentes base de PDF son cp1252: el traductor convierte tildes y eñes.
	tr := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)
	pdf.SetTitle(tr("Ficha de proceso personal"), false)
	pdf.SetCreator("Buen Vivir", false)

	generado := s.now().Format("02/01/2006")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "I", 8)
		setColor(pdf, colorSuave)
		pdf.CellFormat(0, 10, tr(fmt.Sprintf("Documento confidencial (Ley 25.326) - Generado el %s - Página %d", generado, pdf.PageNo())), "", 0, "C", false, 0, "")
	})
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 18)
	setColor(pdf, colorPrimario)
	pdf.CellFormat(0, 10, tr("Ficha de proceso personal"), "", 1, "L", false, 0, "")
	pdf.SetDrawColor(colorPrimario[0], colorPrimario[1], colorPrimario[2])
	pdf.Line(20, pdf.GetY()+1, 190, pdf.GetY()+1)
	pdf.Ln(6)

	seccion(pdf, tr, "Datos del participante")
	if p := ficha.Participante; p != nil {
		campo(pdf, tr, "Nombre", p.Nombre)
		campo(pdf, tr, "Apellido", p.Apellido)
		if p.FechaNacimiento != "" {
			campo(pdf, tr, "Fecha de nacimiento", formatoFecha(p.FechaNacimiento))
		}
		if p.Edad != nil {
			campo(pdf, tr, "Edad", fmt.Sprintf("%d años", *p.Edad))
		}
	} else {
		parrafo(pdf, tr, "Los datos personales no están disponibles en este momento.")
	}
	habilidades := "Sin habilidades registradas"
	if len(ficha.Habilidades) > 0 {
		habilidades = strings.Join(ficha.Habilidades, ", ")
	}
	campo(pdf, tr, "Habilidades", habilidades)
	campo(pdf, tr, "Última actualización", ficha.UpdatedAt.Local().Format("02/01/2006"))
	pdf.Ln(4)

	seccion(pdf, tr, "Situación actual")
	parrafo(pdf, tr, ficha.SituacionActual)
	pdf.Ln(2)

	seccion(pdf, tr, "Avances")
	parrafo(pdf, tr, ficha.Avances)
	pdf.Ln(2)

	seccion(pdf, tr, "Línea de tiempo")
	if len(hitos) == 0 {
		parrafo(pdf, tr, "Todavía no hay hitos registrados.")
	}
	for _, h := range hitos {
		pdf.SetFont("Helvetica", "B", 10)
		setColor(pdf, colorPrimario)
		encabezado := formatoFecha(h.Fecha)
		if h.ResponsableNombre != "" {
			encabezado += " - Responsable: " + h.ResponsableNombre
		}
		pdf.MultiCell(0, 5, tr(encabezado), "", "L", false)
		parrafo(pdf, tr, h.Descripcion)
		pdf.Ln(2)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("error generando el PDF: %w", err)
	}
	return buf.Bytes(), nil
}

func seccion(pdf *fpdf.Fpdf, tr func(string) string, titulo string) {
	pdf.SetFont("Helvetica", "B", 12)
	setColor(pdf, colorPrimario)
	pdf.CellFormat(0, 8, tr(titulo), "", 1, "L", false, 0, "")
}

func campo(pdf *fpdf.Fpdf, tr func(string) string, etiqueta, valor string) {
	pdf.SetFont("Helvetica", "B", 10)
	setColor(pdf, colorTexto)
	pdf.CellFormat(45, 6, tr(etiqueta+":"), "", 0, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 10)
	pdf.MultiCell(0, 6, tr(valor), "", "L", false)
}

func parrafo(pdf *fpdf.Fpdf, tr func(string) string, texto string) {
	pdf.SetFont("Helvetica", "", 10)
	setColor(pdf, colorTexto)
	pdf.MultiCell(0, 5, tr(texto), "", "L", false)
}

func setColor(pdf *fpdf.Fpdf, c [3]int) {
	pdf.SetTextColor(c[0], c[1], c[2])
}

// formatoFecha pasa AAAA-MM-DD a DD/MM/AAAA (HU-31).
func formatoFecha(fecha string) string {
	t, err := time.Parse(models.FechaLayout, fecha)
	if err != nil {
		return fecha
	}
	return t.Format("02/01/2006")
}
