package controllers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"buenvivir/campaigns/internal/models"
)

// La foto subida viaja embebida en el JSON (~700 KB como máximo): el
// límite del cuerpo tiene que dejarla pasar y cortar lo exagerado.
func TestDecodeJSONBody_AceptaFotoYRechazaExagerado(t *testing.T) {
	foto := "data:image/jpeg;base64," + strings.Repeat("A", 690000)
	body := `{"titulo":"t","texto":"x","imagen_url":"` + foto + `","fecha":"2026-11-23"}`
	var in models.CampaignInput
	if err := decodeJSONBody(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(body)), &in); err != nil {
		t.Fatalf("una foto de ~700 KB debería entrar: %v", err)
	}
	if in.ImagenURL != foto {
		t.Fatal("la imagen no llegó completa")
	}

	grande := `{"titulo":"` + strings.Repeat("x", 2<<20) + `"}`
	if err := decodeJSONBody(httptest.NewRecorder(), httptest.NewRequest("POST", "/", strings.NewReader(grande)), &in); err == nil {
		t.Fatal("un cuerpo de 2 MB debería rechazarse")
	}
}
