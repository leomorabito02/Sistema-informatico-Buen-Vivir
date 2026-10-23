package services

import (
	"encoding/json"
	"strings"

	"buenvivir/education/internal/models"
)

// MaxBusquedaCursos acota el texto de búsqueda del catálogo.
const MaxBusquedaCursos = 100

var sinTildes = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u")

func normalizar(s string) string {
	return sinTildes.Replace(strings.ToLower(strings.TrimSpace(s)))
}

// FiltrarCursos deja los cursos cuyo título, área formativa o descripción
// contienen q, sin distinguir mayúsculas ni tildes ("huerta" encuentra
// "Huerta agroecológica"). q vacío devuelve todos. Es el buscador del
// catálogo, que antes filtraba el frontend con la lista entera.
func FiltrarCursos(cursos []models.Course, q string) []models.Course {
	q = normalizar(q)
	if q == "" {
		return cursos
	}
	out := []models.Course{}
	for _, c := range cursos {
		area := ""
		if c.TrainingArea != nil {
			area = c.TrainingArea.Name
		}
		if strings.Contains(normalizar(c.Title), q) ||
			strings.Contains(normalizar(area), q) ||
			strings.Contains(normalizar(TextoPlano(c.Description)), q) {
			out = append(out, c)
		}
	}
	return out
}

// FiltrarCursosPorArea deja los cursos del área formativa areaID
// (HU-28). areaID vacío devuelve todos. Se combina con FiltrarCursos
// para buscar por nombre dentro de un área.
func FiltrarCursosPorArea(cursos []models.Course, areaID string) []models.Course {
	areaID = strings.TrimSpace(areaID)
	if areaID == "" {
		return cursos
	}
	out := []models.Course{}
	for _, c := range cursos {
		if c.TrainingArea != nil && c.TrainingArea.ID == areaID {
			out = append(out, c)
		}
	}
	return out
}

// TextoPlano junta el texto de un documento Rich Text de Contentful (los
// campos "value" de los nodos de texto), separado por espacios. Un JSON
// inválido o vacío devuelve "".
func TextoPlano(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return ""
	}
	var partes []string
	var recorrer func(n any)
	recorrer = func(n any) {
		switch v := n.(type) {
		case map[string]any:
			if v["nodeType"] == "text" {
				if s, ok := v["value"].(string); ok && s != "" {
					partes = append(partes, s)
				}
			}
			if hijos, ok := v["content"].([]any); ok {
				for _, h := range hijos {
					recorrer(h)
				}
			}
		case []any:
			for _, h := range v {
				recorrer(h)
			}
		}
	}
	recorrer(doc)
	return strings.Join(partes, " ")
}
