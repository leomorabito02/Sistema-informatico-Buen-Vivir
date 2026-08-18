package repositories

import (
	"encoding/json"

	"buenvivir/education/internal/models"
)

// Helpers compartidos para leer campos crudos (json.RawMessage) de una
// entry de Contentful tal como vienen de la Content Delivery API
// (valores ya "planos", sin wrapper de locale).

func rawString(fields map[string]json.RawMessage, key string) string {
	raw, ok := fields[key]
	if !ok || raw == nil {
		return ""
	}
	var s string
	_ = json.Unmarshal(raw, &s)
	return s
}

// rawLink extrae el id de un campo Reference/Media de tipo Link.
func rawLink(fields map[string]json.RawMessage, key string) (string, bool) {
	raw, ok := fields[key]
	if !ok || raw == nil {
		return "", false
	}
	var link models.ContentfulLink
	if err := json.Unmarshal(raw, &link); err != nil || link.Sys.ID == "" {
		return "", false
	}
	return link.Sys.ID, true
}

// rawLinks extrae los ids de un campo References (many).
func rawLinks(fields map[string]json.RawMessage, key string) []string {
	raw, ok := fields[key]
	if !ok || raw == nil {
		return nil
	}
	var links []models.ContentfulLink
	if err := json.Unmarshal(raw, &links); err != nil {
		return nil
	}
	ids := make([]string, 0, len(links))
	for _, l := range links {
		if l.Sys.ID != "" {
			ids = append(ids, l.Sys.ID)
		}
	}
	return ids
}
