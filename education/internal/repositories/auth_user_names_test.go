package repositories

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestGetUserNames_PideConTokenDeServicio(t *testing.T) {
	t.Setenv("JWT_SECRET", "secreto-de-prueba")
	var pedidos int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pedidos++
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/directory/users/summaries" {
			t.Errorf("pedido inesperado %s %s", r.Method, r.URL.Path)
		}
		tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		claims := jwt.MapClaims{}
		if _, err := jwt.ParseWithClaims(tok, claims, func(*jwt.Token) (any, error) { return []byte("secreto-de-prueba"), nil }); err != nil {
			t.Errorf("token inválido: %v", err)
		}
		if claims["role"] != "role_dev" || claims["type"] != "access" {
			t.Errorf("claims inesperados: %v", claims)
		}
		var body struct {
			UserIDs []string `json:"user_ids"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		data := []map[string]string{}
		for _, id := range body.UserIDs {
			data = append(data, map[string]string{"user_id": id, "full_name": "Nombre " + id})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": data})
	}))
	defer srv.Close()

	c := NewAuthServiceClient(srv.URL)
	ids := make([]string, 0, 60)
	for i := 0; i < 60; i++ {
		ids = append(ids, "usr-"+strings.Repeat("x", i%3)+string(rune('a'+i%26)))
	}
	got, err := c.GetUserNames(context.Background(), ids)
	if err != nil {
		t.Fatal(err)
	}
	if pedidos != 2 {
		t.Errorf("60 ids deberían ir en 2 pedidos de hasta 50, fueron %d", pedidos)
	}
	if got[ids[0]] != "Nombre "+ids[0] {
		t.Errorf("nombre inesperado: %q", got[ids[0]])
	}
}

func TestGetUserNames_SinSecretoFalla(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	c := NewAuthServiceClient("http://127.0.0.1:1")
	if _, err := c.GetUserNames(context.Background(), []string{"u"}); err == nil {
		t.Error("sin JWT_SECRET debe fallar")
	}
}
