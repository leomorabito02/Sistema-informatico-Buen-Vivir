package repositories

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

// ErrNotFound es el sentinel que usan los repositories que no hablan
// con Contentful (Firestore) para señalar "no existe", equivalente a
// un ContentfulAPIError con StatusCode 404. IsNotFound reconoce ambos
// casos para que la capa de controllers no tenga que distinguir de
// qué fuente de datos vino el error.
var ErrNotFound = errors.New("recurso no encontrado")

// fieldLabels traduce los ids de campo de Contentful a como los ve el
// usuario en el formulario, para armar el mensaje de ValidationMessage.
var fieldLabels = map[string]string{
	"title":              "título",
	"description":        "descripción",
	"responsibleTeacher": "docente responsable",
	"coverImage":         "imagen de portada",
	"mode":               "modalidad",
	"duration":           "duración",
	"trainingArea":       "área formativa",
	"material":           "materiales",
	"capacity":           "cupo",
	"name":               "nombre",
	"type":               "tipo",
	"file":               "archivo",
	"url":                "URL",
}

// ValidationMessage reconoce un 422 ValidationFailed de la CMA (por ej.
// al publicar un borrador al que le faltan campos obligatorios) y lo
// traduce a un mensaje legible. Devuelve ok=false si err no es eso.
func ValidationMessage(err error) (string, bool) {
	apiErr, ok := err.(*ContentfulAPIError)
	if !ok || apiErr.StatusCode != http.StatusUnprocessableEntity {
		return "", false
	}

	var body struct {
		Details struct {
			Errors []struct {
				Name string `json:"name"`
				Path []any  `json:"path"`
			} `json:"errors"`
		} `json:"details"`
	}
	_ = json.Unmarshal([]byte(apiErr.Body), &body)

	var missing, unresolved, duplicated, other []string
	for _, e := range body.Details.Errors {
		field := "?"
		if len(e.Path) >= 2 {
			if s, ok := e.Path[1].(string); ok {
				field = s
			}
		}
		if label, ok := fieldLabels[field]; ok {
			field = label
		}
		switch e.Name {
		case "required":
			missing = appendUnique(missing, field)
		case "notResolvable":
			unresolved = appendUnique(unresolved, field)
		case "unique":
			duplicated = appendUnique(duplicated, field)
		default:
			other = appendUnique(other, field)
		}
	}

	var parts []string
	if len(missing) > 0 {
		parts = append(parts, "faltan campos obligatorios: "+strings.Join(missing, ", "))
	}
	if len(unresolved) > 0 {
		parts = append(parts, "hay referencias borradas o sin publicar en: "+strings.Join(unresolved, ", "))
	}
	if len(duplicated) > 0 {
		parts = append(parts, "ya existe otro contenido con el mismo valor en: "+strings.Join(duplicated, ", "))
	}
	if len(other) > 0 {
		parts = append(parts, "hay valores inválidos en: "+strings.Join(other, ", "))
	}
	if len(parts) == 0 {
		return "Contentful rechazó el contenido por validación", true
	}
	return "No se puede publicar: " + strings.Join(parts, "; ") + ". Editalo, completá esos datos y volvé a publicar.", true
}

func appendUnique(list []string, v string) []string {
	for _, x := range list {
		if x == v {
			return list
		}
	}
	return append(list, v)
}

func IsNotFound(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}
	apiErr, ok := err.(*ContentfulAPIError)
	return ok && apiErr.StatusCode == http.StatusNotFound
}
