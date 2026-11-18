package services_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"buenvivir/education/internal/models"
	"buenvivir/education/internal/repositories"
)

// fakeAuthServer simula los endpoints de auth-service que consume
// AuthServiceClient (GET /participants, /participants/{id}, /users),
// con el mismo sobre {success, data} que auth/internal/utils/response.go.
// AuthServiceClient es un tipo concreto (no una interfaz), así que se
// prueba contra HTTP real en vez de mockearlo.
type fakeAuthServer struct {
	participants  []models.AuthParticipant
	users         []models.AuthUser
	failList      bool // GET /participants responde 500
	failUsers     bool // GET /users responde 500
	failGet       bool // GET /participants/{id} responde 500
	receivedToken string
}

func (f *fakeAuthServer) start(t *testing.T) *repositories.AuthServiceClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(f.handle))
	t.Cleanup(srv.Close)
	return repositories.NewAuthServiceClient(srv.URL)
}

func (f *fakeAuthServer) handle(w http.ResponseWriter, r *http.Request) {
	f.receivedToken = strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")

	switch {
	case r.URL.Path == "/api/v1/participants":
		if f.failList {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeAuthData(w, f.participants)
	case strings.HasPrefix(r.URL.Path, "/api/v1/participants/"):
		if f.failGet {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/participants/")
		for _, p := range f.participants {
			if p.ParticipantID == id || p.UserID == id {
				writeAuthData(w, p)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
	case r.URL.Path == "/api/v1/users":
		if f.failUsers {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		writeAuthData(w, f.users)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func writeAuthData(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "data": data})
}
