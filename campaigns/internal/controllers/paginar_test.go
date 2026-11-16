package controllers

import (
	"testing"

	"buenvivir/campaigns/internal/models"
)

// HU-50: página del listado público y validación de limit/offset.
func TestPaginar(t *testing.T) {
	all := make([]models.Campaign, 25)
	for i := range all {
		all[i].ID = string(rune('a' + i))
	}
	p, ok := paginar(all, "10", "20")
	if !ok || p.Total != 25 || len(p.Items) != 5 || p.Items[0].ID != "u" {
		t.Fatalf("página inesperada: ok=%v total=%d items=%d", ok, p.Total, len(p.Items))
	}
	if p, ok := paginar(all, "10", "40"); !ok || len(p.Items) != 0 || p.Items == nil {
		t.Fatalf("fuera de rango debe dar lista vacía: %+v", p)
	}
	for _, c := range [][2]string{{"0", ""}, {"101", ""}, {"x", ""}, {"10", "-1"}, {"10", "y"}} {
		if _, ok := paginar(all, c[0], c[1]); ok {
			t.Errorf("limit=%q offset=%q debería ser inválido", c[0], c[1])
		}
	}
}
