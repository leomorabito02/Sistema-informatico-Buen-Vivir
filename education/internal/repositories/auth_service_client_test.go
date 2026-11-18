package repositories

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"buenvivir/education/api/proto/authpb"
	"buenvivir/education/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	_, err = c.ListParticipants(t.Context(), "tok")
	assert.ErrorIs(t, err, ErrAuthServiceNotConfigured)
	_, err = c.ListUsers(t.Context(), "tok")
	assert.ErrorIs(t, err, ErrAuthServiceNotConfigured)

	var nilClient *AuthServiceClient
	_, err = nilClient.GetMyParticipantID(t.Context(), "tok")
	assert.ErrorIs(t, err, ErrAuthServiceNotConfigured)
	_, err = nilClient.GetParticipant(t.Context(), "p1", "tok")
	assert.ErrorIs(t, err, ErrAuthServiceNotConfigured)
	_, err = nilClient.ListParticipants(t.Context(), "tok")
	assert.ErrorIs(t, err, ErrAuthServiceNotConfigured)
	_, err = nilClient.ListUsers(t.Context(), "tok")
	assert.ErrorIs(t, err, ErrAuthServiceNotConfigured)
}

func TestAuthServiceClient_GetParticipant_InvalidJSON(t *testing.T) {
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{invalid-json`))
	}))
	defer srv.Close()

	c := NewAuthServiceClient(srv.URL)

	// Act
	_, err := c.GetParticipant(t.Context(), "p1", "tok")

	// Assert
	assert.Error(t, err)
}

func TestAuthServiceClient_DoGet_InvalidJSON(t *testing.T) {
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{invalid-json`))
	}))
	defer srv.Close()

	c := NewAuthServiceClient(srv.URL)

	// Act & Assert
	_, err := c.ListParticipants(t.Context(), "tok")
	assert.Error(t, err)

	_, err = c.ListUsers(t.Context(), "tok")
	assert.Error(t, err)
}

func TestAuthServiceClient_NetworkError(t *testing.T) {
	// Arrange
	c := NewAuthServiceClient("http://127.0.0.1:0")

	// Act & Assert
	_, err := c.GetParticipant(t.Context(), "p1", "tok")
	assert.Error(t, err)

	_, err = c.ListParticipants(t.Context(), "tok")
	assert.Error(t, err)
}

func TestAuthServiceClient_ServerlessAuthHeader(t *testing.T) {
	// Arrange
	var receivedServerlessAuth, receivedAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedServerlessAuth = r.Header.Get("X-Serverless-Authorization")
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"participant_id":"p1","user_id":"u1","approval_status":"accepted"}}`))
	}))
	defer srv.Close()

	idTokenFn := func(ctx context.Context) (string, error) {
		return "google-id-token-123", nil
	}
	c := NewAuthServiceClient(srv.URL, idTokenFn)

	// Act
	_, err := c.GetParticipant(t.Context(), "p1", "user-jwt-456")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Bearer google-id-token-123", receivedServerlessAuth)
	assert.Equal(t, "Bearer user-jwt-456", receivedAuth)
}

type fakeAuthGRPCServer struct {
	authpb.UnimplementedAuthServiceServer
}

func (f *fakeAuthGRPCServer) GetParticipant(_ context.Context, req *authpb.GetParticipantRequest) (*authpb.GetParticipantResponse, error) {
	if req.GetParticipantId() == "p1" {
		return &authpb.GetParticipantResponse{Id: "p1", UserId: "u1", ApprovalStatus: "accepted", Dni: "12345678"}, nil
	}
	return nil, status.Error(codes.NotFound, "not found")
}

func (f *fakeAuthGRPCServer) ListParticipants(context.Context, *authpb.ListParticipantsRequest) (*authpb.ListParticipantsResponse, error) {
	return &authpb.ListParticipantsResponse{
		Participants: []*authpb.GetParticipantResponse{
			{Id: "p1", UserId: "u1", ApprovalStatus: "accepted", Dni: "12345678"},
		},
	}, nil
}

func (f *fakeAuthGRPCServer) ListUsers(context.Context, *authpb.ListUsersRequest) (*authpb.GetUsersByIDsResponse, error) {
	return &authpb.GetUsersByIDsResponse{
		Users: []*authpb.UserSummary{
			{UserId: "u1", FirstName: "Ana", LastName: "Pérez"},
		},
	}, nil
}

func (f *fakeAuthGRPCServer) GetParticipantProfile(_ context.Context, req *authpb.GetParticipantProfileRequest) (*authpb.ParticipantProfile, error) {
	switch l := req.GetLookup().(type) {
	case *authpb.GetParticipantProfileRequest_ParticipantId:
		if l.ParticipantId == "p1" {
			return &authpb.ParticipantProfile{ParticipantId: "p1", UserId: "u1", FirstName: "Ana", LastName: "Pérez"}, nil
		}
	case *authpb.GetParticipantProfileRequest_UserId:
		if l.UserId == "u1" {
			return &authpb.ParticipantProfile{ParticipantId: "p1", UserId: "u1", FirstName: "Ana", LastName: "Pérez"}, nil
		}
	}
	return nil, status.Error(codes.NotFound, "not found")
}

func TestAuthServiceClient_GRPC(t *testing.T) {
	// Arrange
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	srv := grpc.NewServer()
	authpb.RegisterAuthServiceServer(srv, &fakeAuthGRPCServer{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	c := NewAuthServiceClient(lis.Addr().String())
	t.Cleanup(func() { _ = c.Close() })

	// Act & Assert
	p, err := c.GetParticipant(t.Context(), "p1", "u1")
	require.NoError(t, err)
	require.NotNil(t, p)
	assert.Equal(t, "u1", p.UserID)
	assert.Equal(t, "12345678", p.DNI)

	missing, err := c.GetParticipant(t.Context(), "no-existe", "u1")
	require.NoError(t, err)
	assert.Nil(t, missing)

	participants, err := c.ListParticipants(t.Context(), "u1")
	require.NoError(t, err)
	require.Len(t, participants, 1)
	assert.Equal(t, "p1", participants[0].ParticipantID)

	users, err := c.ListUsers(t.Context(), "u1")
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.Equal(t, "Ana", users[0].FirstName)

	u, err := c.GetParticipantName(t.Context(), "p1", "u1")
	require.NoError(t, err)
	assert.Equal(t, "Ana", u.FirstName)

	pid, err := c.GetMyParticipantID(t.Context(), "u1")
	require.NoError(t, err)
	assert.Equal(t, "p1", pid)
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
