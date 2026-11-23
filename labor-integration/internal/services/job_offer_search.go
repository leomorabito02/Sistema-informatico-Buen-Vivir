package services

import (
	"strings"

	"buenvivir/labor-integration/internal/models"
)

// MaxBusquedaOfertas acota el texto de búsqueda de ofertas.
const MaxBusquedaOfertas = 100

var sinTildes = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u", "ñ", "n")

func normalizar(s string) string {
	return sinTildes.Replace(strings.ToLower(strings.TrimSpace(s)))
}

// FiltrarOfertas deja las ofertas cuyo título, empresa o ubicación contienen q (HU-45,
// HU-46) y, si viene empresa, solo las de esa empresa. No distingue mayúsculas ni tildes.
func FiltrarOfertas(offers []*models.JobOffer, q, empresa string) []*models.JobOffer {
	q = normalizar(q)
	empresa = normalizar(empresa)
	out := []*models.JobOffer{}
	for _, o := range offers {
		nombreEmpresa := normalizar(o.CreatorName)
		if empresa != "" && !strings.Contains(nombreEmpresa, empresa) && normalizar(o.CompanyID) != empresa {
			continue
		}
		if q != "" && !strings.Contains(normalizar(o.Title), q) &&
			!strings.Contains(nombreEmpresa, q) && !strings.Contains(normalizar(o.Location), q) {
			continue
		}
		out = append(out, o)
	}
	return out
}
