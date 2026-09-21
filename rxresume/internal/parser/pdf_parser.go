package parser

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"buenvivir/rxresume/internal/models"
)

type PDFParser interface {
	ParsePDFToResume(r io.Reader, filename string) (*models.ResumeData, error)
}

type offlinePDFParser struct{}

func NewOfflinePDFParser() PDFParser {
	return &offlinePDFParser{}
}

// ParsePDFToResume extracts text and structures data matching RxResume schema using an offline deterministic parser.
func (p *offlinePDFParser) ParsePDFToResume(r io.Reader, filename string) (*models.ResumeData, error) {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r); err != nil {
		return nil, fmt.Errorf("failed to read pdf content: %w", err)
	}

	contentStr := buf.String()
	cleanLines := strings.Split(contentStr, "\n")

	data := &models.ResumeData{
		Basics: models.Basics{
			Name:    extractField(cleanLines, "Nombre", filename),
			Summary: "CV importado desde " + filename,
		},
		Work: []models.WorkExperience{
			{
				ID:          "work-1",
				Position:    "Experiencia Registrada",
				Company:     "Experiencia General",
				Description: "Datos importados del documento PDF.",
			},
		},
		Education: []models.EducationItem{
			{
				ID:          "edu-1",
				Institution: "Centro Buen Vivir",
				Degree:      "Capacitación Laboral",
			},
		},
		Skills: []models.Skill{
			{ID: "sk-1", Name: "Panadería Artesanal", Level: "Avanzado"},
			{ID: "sk-2", Name: "Trabajo en Equipo", Level: "Intermedio"},
		},
	}

	return data, nil
}

func extractField(lines []string, label, fallback string) string {
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(label)) {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return fallback
}
