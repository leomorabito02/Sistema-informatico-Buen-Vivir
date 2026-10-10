package repositories

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewEducationClient_EmptyURLIsDisabled(t *testing.T) {
	if NewEducationClient("  ") != nil {
		t.Fatal("sin URL, el cliente tiene que ser nil (ficha sin habilidades)")
	}
}

func TestEducationClient_HabilidadesDe_ForwardsTokenAndParsesNames(t *testing.T) {
	var gotAuth, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotPath = r.Header.Get("Authorization"), r.URL.Path
		_, _ = w.Write([]byte(`[{"id":"compostaje","nombre":"Compostaje"},{"id":"x","nombre":""},{"id":"uso-de-excel","nombre":"Uso de Excel"}]`))
	}))
	defer srv.Close()

	client := NewEducationClient(srv.URL + "/")
	ctx := ContextWithToken(context.Background(), "jwt-del-voluntario")
	got, err := client.HabilidadesDe(ctx, "p1")
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "Bearer jwt-del-voluntario" {
		t.Errorf("no reenvió el token: %q", gotAuth)
	}
	if gotPath != "/api/participants/p1/habilidades" {
		t.Errorf("path inesperado: %q", gotPath)
	}
	if len(got) != 2 || got[0] != "Compostaje" || got[1] != "Uso de Excel" {
		t.Errorf("habilidades inesperadas: %v", got)
	}
}

func TestEducationClient_HabilidadesDe_Errors(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"HTTP 403":      func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusForbidden) },
		"JSON inválido": func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("<html>")) },
	}
	for name, h := range cases {
		srv := httptest.NewServer(h)
		_, err := NewEducationClient(srv.URL).HabilidadesDe(context.Background(), "p1")
		srv.Close()
		if err == nil {
			t.Errorf("%s: se esperaba error", name)
		}
	}

	_, err := NewEducationClient("http://127.0.0.1:1").HabilidadesDe(context.Background(), "p1")
	if err == nil {
		t.Error("servidor caído: se esperaba error")
	}
}
