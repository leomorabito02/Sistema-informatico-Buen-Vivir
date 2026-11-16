package repositories

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"buenvivir/labor-integration/api/proto/educationpb"

	"google.golang.org/grpc"
)

type fakeEducationGRPCServer struct {
	educationpb.UnimplementedEducationServiceServer
}

func (f *fakeEducationGRPCServer) GetGrantedSkills(context.Context, *educationpb.GetGrantedSkillsRequest) (*educationpb.GetGrantedSkillsResponse, error) {
	return &educationpb.GetGrantedSkillsResponse{
		Skills: []*educationpb.GrantedSkillItem{
			{ParticipanteId: "part-1", Nombre: "Soldadura"},
			{ParticipanteId: "part-2", Nombre: "Electricidad"},
		},
	}, nil
}

func TestSkillsClient_GRPC(t *testing.T) {
	// Arrange
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	educationpb.RegisterEducationServiceServer(srv, &fakeEducationGRPCServer{})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	client := NewSkillsClient(lis.Addr().String())
	if client == nil {
		t.Fatal("expected non-nil SkillsClient")
	}

	// Act
	skills, err := client.GrantedSkills(context.Background(), "admin-token")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(skills) != 2 || skills[0].ParticipanteID != "part-1" || skills[0].Nombre != "Soldadura" {
		t.Errorf("unexpected skills: %+v", skills)
	}
}

func TestSkillsClient_HTTPWithIDToken(t *testing.T) {
	// Arrange
	var gotServerless, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotServerless = r.Header.Get("X-Serverless-Authorization")
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"participanteId":"p-1","nombre":"Carpintería"}]`))
	}))
	defer srv.Close()

	client := NewSkillsClient(srv.URL, func(context.Context) (string, error) {
		return "gcp-id-tok", nil
	})

	// Act
	skills, err := client.GrantedSkills(context.Background(), "user-jwt")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotServerless != "Bearer gcp-id-tok" {
		t.Errorf("expected X-Serverless-Authorization Bearer gcp-id-tok, got %s", gotServerless)
	}
	if gotAuth != "Bearer user-jwt" {
		t.Errorf("expected Authorization Bearer user-jwt, got %s", gotAuth)
	}
	if len(skills) != 1 || skills[0].Nombre != "Carpintería" {
		t.Errorf("unexpected skills: %+v", skills)
	}
}
