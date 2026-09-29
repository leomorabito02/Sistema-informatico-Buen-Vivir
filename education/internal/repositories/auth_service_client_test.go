package repositories

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"buenvivir/education/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeAuthService responde como auth-service y exige que education
// reenvíe el mismo token del usuario.
func fakeAuthService(t *testing.T) *AuthServiceClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer tok" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/profile/me":
			_, _ = w.Write([]byte(`{"user":{"user_id":"u1"},"role_data":{"participant_id":"p1","user_id":"u1"}}`))
		case "/api/v1/participants":
			_, _ = w.Write([]byte(`{"success":true,"data":[{"participant_id":"p1","user_id":"u1","approval_status":"accepted"}]}`))
		case "/api/v1/participants/p1":
			_, _ = w.Write([]byte(`{"success":true,"data":{"participant_id":"p1","user_id":"u1","approval_status":"accepted"}}`))
		case "/api/v1/participants/roto":
			w.WriteHeader(http.StatusInternalServerError)
		case "/api/v1/users":
			_, _ = w.Write([]byte(`{"success":true,"data":[{"user_id":"u1","first_name":"Ana","last_name":"Pérez"}]}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return NewAuthServiceClient(srv.URL)
}

func TestAuthServiceClient_GetMyParticipantID(t *testing.T) {
	c := fakeAuthService(t)
	id, err := c.GetMyParticipantID(t.Context(), "tok")
	require.NoError(t, err)
	assert.Equal(t, "p1", id)

	_, err = c.GetMyParticipantID(t.Context(), "otro")
	assert.Error(t, err, "401 de auth-service se propaga")
}

func TestAuthServiceClient_ListParticipantsAndUsers(t *testing.T) {
	c := fakeAuthService(t)

	participants, err := c.ListParticipants(t.Context(), "tok")
	require.NoError(t, err)
	assert.Equal(t, []models.AuthParticipant{{ParticipantID: "p1", UserID: "u1", ApprovalStatus: "accepted"}}, participants)

	users, err := c.ListUsers(t.Context(), "tok")
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Ana", users[0].FirstName)
}

func TestAuthServiceClient_GetParticipant(t *testing.T) {
	c := fakeAuthService(t)

	p, err := c.GetParticipant(t.Context(), "p1", "tok")
	require.NoError(t, err)
	assert.Equal(t, "u1", p.UserID)

	p, err = c.GetParticipant(t.Context(), "no-existe", "tok")
	require.NoError(t, err)
	assert.Nil(t, p, "404 = el participante no existe, no es error de infraestructura")

	_, err = c.GetParticipant(t.Context(), "roto", "tok")
	assert.Error(t, err)
}

func TestAuthServiceClient_NotConfigured(t *testing.T) {
	c := NewAuthServiceClient("")

	_, err := c.GetMyParticipantID(t.Context(), "tok")
	assert.ErrorIs(t, err, ErrAuthServiceNotConfigured)
	_, err = c.GetParticipant(t.Context(), "p1", "tok")
	assert.ErrorIs(t, err, ErrAuthServiceNotConfigured)
}

func TestGeneralAttendanceRepository_RegisterAndList(t *testing.T) {
	client := newTestFirestore(t)
	wipeCollection(t, client, generalAttendanceCollection)
	repo := NewGeneralAttendanceRepository(client)

	require.NoError(t, repo.RegisterBatch(t.Context(), "2026-10-05", []models.GeneralAttendanceInput{
		{ParticipanteID: "p1", Presente: true, Observacion: "llegó tarde"},
		{ParticipanteID: "p2", Presente: false},
	}))
	require.NoError(t, repo.RegisterBatch(t.Context(), "2026-10-06", []models.GeneralAttendanceInput{{ParticipanteID: "p1", Presente: true}}))

	got, err := repo.ListByDate(t.Context(), "2026-10-05")
	require.NoError(t, err)
	assert.Len(t, got, 2)
}
