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

// rawInt extrae un campo Number opcional. Si el campo no existe todavía
// en el content type de Contentful (caso típico: "capacity" en cursos
// creados antes de agregar el campo al Space) o no vino en la respuesta,
// devuelve nil sin error — se interpreta como "sin límite".
func rawInt(fields map[string]json.RawMessage, key string) *int {
	raw, ok := fields[key]
	if !ok || raw == nil {
		return nil
	}
	var n int
	if err := json.Unmarshal(raw, &n); err != nil {
		return nil
	}
	return &n
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

// unwrapLocale desenvuelve el valor de un campo tal como lo devuelve la
// Content Management API — envuelto por locale ({"<locale>": valor}) —
// y devuelve el RawMessage interno. Si el locale configurado no está,
// cae al primero disponible.
func unwrapLocale(raw json.RawMessage, locale string) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	var byLocale map[string]json.RawMessage
	if err := json.Unmarshal(raw, &byLocale); err != nil {
		return nil
	}
	if v, ok := byLocale[locale]; ok {
		return v
	}
	for _, v := range byLocale {
		return v
	}
	return nil
}

// flattenCMAFields aplana los campos locale-wrapped de una entry de la
// CMA al mismo formato "plano" que sirve la CDA, para poder reusar
// rawString/rawLink/rawLinks y el resto del mapeo sin duplicarlo.
func flattenCMAFields(fields map[string]json.RawMessage, locale string) map[string]json.RawMessage {
	flat := make(map[string]json.RawMessage, len(fields))
	for k, v := range fields {
		if u := unwrapLocale(v, locale); u != nil {
			flat[k] = u
		}
	}
	return flat
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
