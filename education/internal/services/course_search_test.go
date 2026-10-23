package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"buenvivir/education/internal/models"
)

func richText(textos ...string) json.RawMessage {
	parrafos := []map[string]any{}
	for _, t := range textos {
		parrafos = append(parrafos, map[string]any{
			"nodeType": "paragraph",
			"content":  []map[string]any{{"nodeType": "text", "value": t, "marks": []any{}}},
		})
	}
	b, _ := json.Marshal(map[string]any{"nodeType": "document", "content": parrafos})
	return b
}

var catalogo = []models.Course{
	{ID: "c1", Title: "Huerta agroecológica", TrainingArea: &models.TrainingArea{Name: "Ambiente"}, Description: richText("Cultivo de verduras de estación.")},
	{ID: "c2", Title: "Panadería", TrainingArea: &models.TrainingArea{Name: "Oficios"}, Description: richText("Pan casero y facturas.", "Incluye compostaje de restos.")},
	{ID: "c3", Title: "Electricidad básica", Description: nil},
}

func ids(cs []models.Course) []string {
	out := []string{}
	for _, c := range cs {
		out = append(out, c.ID)
	}
	return out
}

func TestFiltrarCursos(t *testing.T) {
	casos := map[string][]string{
		"":              {"c1", "c2", "c3"},
		"   ":           {"c1", "c2", "c3"},
		"huerta":        {"c1"}, // título, sin importar mayúsculas
		"AGROECOLOGICA": {"c1"}, // sin tildes encuentra con tildes
		"oficios":       {"c2"}, // área formativa
		"compostaje":    {"c2"}, // segundo párrafo de la descripción
		"estación":      {"c1"}, // con tilde en la búsqueda
		"electricidad":  {"c3"}, // curso sin área ni descripción
		"carpintería":   {},
	}
	for q, want := range casos {
		assert.Equal(t, want, ids(FiltrarCursos(catalogo, q)), "búsqueda %q", q)
	}
}

func TestTextoPlano(t *testing.T) {
	assert.Equal(t, "Pan casero y facturas. Incluye compostaje de restos.", TextoPlano(richText("Pan casero y facturas.", "Incluye compostaje de restos.")))
	assert.Equal(t, "", TextoPlano(nil))
	assert.Equal(t, "", TextoPlano(json.RawMessage(`{no es json`)))
	assert.Equal(t, "", TextoPlano(json.RawMessage(`null`)))
}
