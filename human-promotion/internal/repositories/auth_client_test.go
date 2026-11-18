package repositories

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPAuthDirectory_AllMethodsAndHeaders(t *testing.T) {
	var receivedServerlessAuth, receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedServerlessAuth = r.Header.Get("X-Serverless-Authorization")
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/directory/participants/p1":
			_, _ = w.Write([]byte(`{"success":true,"data":{"participant_id":"p1","user_id":"u1","nombre":"Juan","apellido":"Perez","fecha_nacimiento":"1995-05-10"}}`))
		case "/api/v1/directory/participants/by-user/u1":
			_, _ = w.Write([]byte(`{"success":true,"data":{"participant_id":"p1","user_id":"u1","nombre":"Juan","apellido":"Perez","fecha_nacimiento":"1995-05-10"}}`))
		case "/api/v1/directory/participants/search":
			_, _ = w.Write([]byte(`{"success":true,"data":[{"participant_id":"p1","user_id":"u1","nombre":"Juan","apellido":"Perez"}]}`))
		case "/api/v1/directory/users/summaries":
			_, _ = w.Write([]byte(`{"success":true,"data":[{"user_id":"u1","full_name":"Juan Perez"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	idTokenFn := func(ctx context.Context) (string, error) {
		return "mock-id-token", nil
	}
	dir, err := NewAuthDirectory(srv.URL, idTokenFn)
	if err != nil {
		t.Fatalf("NewAuthDirectory falló: %v", err)
	}
	defer func() { _ = dir.Close() }()

	ctx := ContextWithToken(context.Background(), "user-token-123")

	// 1. GetParticipante
	part, err := dir.GetParticipante(ctx, "p1")
	if err != nil {
		t.Fatalf("GetParticipante falló: %v", err)
	}
	if part.Nombre != "Juan" || part.Apellido != "Perez" {
		t.Errorf("datos inesperados: %+v", part)
	}
	if receivedServerlessAuth != "Bearer mock-id-token" {
		t.Errorf("X-Serverless-Authorization esperado 'Bearer mock-id-token', recibido '%s'", receivedServerlessAuth)
	}
	if receivedAuth != "Bearer user-token-123" {
		t.Errorf("Authorization esperado 'Bearer user-token-123', recibido '%s'", receivedAuth)
	}

	// 2. GetParticipanteByUserID
	partByUser, err := dir.GetParticipanteByUserID(ctx, "u1")
	if err != nil {
		t.Fatalf("GetParticipanteByUserID falló: %v", err)
	}
	if partByUser.ParticipanteID != "p1" {
		t.Errorf("ParticipanteID esperado 'p1', recibido '%s'", partByUser.ParticipanteID)
	}

	// 3. SearchParticipantes
	results, err := dir.SearchParticipantes(ctx, "Juan", 10)
	if err != nil {
		t.Fatalf("SearchParticipantes falló: %v", err)
	}
	if len(results) != 1 || results[0].Nombre != "Juan" {
		t.Errorf("resultados de búsqueda inesperados: %+v", results)
	}

	// 4. GetNombresUsuarios
	names, err := dir.GetNombresUsuarios(ctx, []string{"u1"})
	if err != nil {
		t.Fatalf("GetNombresUsuarios falló: %v", err)
	}
	if names["u1"] != "Juan Perez" {
		t.Errorf("nombre esperado 'Juan Perez', recibido '%s'", names["u1"])
	}
}
