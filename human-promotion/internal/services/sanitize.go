package services

import (
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"buenvivir/human-promotion/internal/models"
)

// Largos máximos de los textos libres, medidos en caracteres.
const (
	maxTextoFicha      = 5000
	maxDescripcionHito = 2000
	minAnioFechaHito   = 1900
	maxAnioFechaHito   = 2100
)

var (
	htmlTagPattern = regexp.MustCompile(`</?[a-zA-Z][^>]*>`)
	idPattern      = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)
)

// sanitizeText normaliza un texto libre antes de cifrarlo: descarta
// UTF-8 inválido, quita etiquetas HTML y caracteres de control (salvo
// salto de línea y tabulación) y recorta espacios en los extremos. Los
// textos se guardan y se devuelven como texto plano; el frontend los
// muestra con interpolación, nunca como HTML.
func sanitizeText(s string) string {
	s = strings.ToValidUTF8(s, "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = htmlTagPattern.ReplaceAllString(s, "")
	s = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return r
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}

func tooLong(s string, max int) bool {
	return utf8.RuneCountInString(s) > max
}

// validID acepta los ids que generan auth-service y Firestore
// (UUIDs, ids alfanuméricos), y rechaza cualquier otra cosa.
func validID(id string) bool {
	return idPattern.MatchString(id)
}

// parseFechaHito valida que la fecha venga como AAAA-MM-DD y caiga en
// un rango razonable.
func parseFechaHito(fecha string) (string, bool) {
	t, err := time.Parse(models.FechaLayout, strings.TrimSpace(fecha))
	if err != nil || t.Year() < minAnioFechaHito || t.Year() > maxAnioFechaHito {
		return "", false
	}
	return t.Format(models.FechaLayout), true
}
