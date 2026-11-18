package services

import (
	"testing"

	"buenvivir/labor-integration/internal/models"
)

func TestFiltrarOfertas(t *testing.T) {
	offers := []*models.JobOffer{
		{ID: "1", Title: "Ayudante de cocina", CreatorName: "Panadería Núñez", CompanyID: "comp-1", Location: "Alta Gracia"},
		{ID: "2", Title: "Jardinero", CreatorName: "Viveros del Sur", CompanyID: "comp-2", Location: "Córdoba"},
		{ID: "3", Title: "Cocinero", CreatorName: "Viveros del Sur", CompanyID: "comp-2", Location: "Alta Gracia"},
	}
	ids := func(os []*models.JobOffer) string {
		s := ""
		for _, o := range os {
			s += o.ID
		}
		return s
	}

	casos := []struct {
		nombre, q, empresa, quiero string
	}{
		{"sin filtros", "", "", "123"},
		{"por título, sin tildes ni mayúsculas", "COCIN", "", "13"},
		{"por empresa en el buscador (HU-46)", "nunez", "", "1"},
		{"por ubicación", "cordoba", "", "2"},
		{"empresa por nombre", "", "viveros", "23"},
		{"empresa por id", "", "comp-1", "1"},
		{"texto y empresa combinados", "cocin", "viveros", "3"},
		{"sin resultados", "plomero", "", ""},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := ids(FiltrarOfertas(offers, c.q, c.empresa)); got != c.quiero {
				t.Errorf("FiltrarOfertas(%q, %q) = %q, quiero %q", c.q, c.empresa, got, c.quiero)
			}
		})
	}
	if FiltrarOfertas(nil, "", "") == nil {
		t.Error("una lista vacía tiene que ser [] y no null en el JSON")
	}
}
