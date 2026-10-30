package clients

import (
	"context"
	"errors"
	"testing"

	"buenvivir/notifications/api/proto/authpb"
	"google.golang.org/grpc"
)

type mockAuthServiceClient struct {
	tokens       []string
	getTokensErr error
	revokeErr    error
}

func (m *mockAuthServiceClient) GetUserDeviceTokens(ctx context.Context, in *authpb.GetUserDeviceTokensRequest, opts ...grpc.CallOption) (*authpb.GetUserDeviceTokensResponse, error) {
	if m.getTokensErr != nil {
		return nil, m.getTokensErr
	}
	return &authpb.GetUserDeviceTokensResponse{Tokens: m.tokens}, nil
}

func (m *mockAuthServiceClient) RevokeDeviceTokens(ctx context.Context, in *authpb.RevokeDeviceTokensRequest, opts ...grpc.CallOption) (*authpb.RevokeDeviceTokensResponse, error) {
	if m.revokeErr != nil {
		return nil, m.revokeErr
	}
	return &authpb.RevokeDeviceTokensResponse{RevokedCount: int32(len(in.Tokens))}, nil
}

func (m *mockAuthServiceClient) GetParticipant(ctx context.Context, in *authpb.GetParticipantRequest, opts ...grpc.CallOption) (*authpb.GetParticipantResponse, error) {
	return nil, nil
}

func (m *mockAuthServiceClient) GetParticipantProfile(ctx context.Context, in *authpb.GetParticipantProfileRequest, opts ...grpc.CallOption) (*authpb.ParticipantProfile, error) {
	return nil, nil
}

func (m *mockAuthServiceClient) SearchParticipants(ctx context.Context, in *authpb.SearchParticipantsRequest, opts ...grpc.CallOption) (*authpb.SearchParticipantsResponse, error) {
	return nil, nil
}

func (m *mockAuthServiceClient) GetUsersByIDs(ctx context.Context, in *authpb.GetUsersByIDsRequest, opts ...grpc.CallOption) (*authpb.GetUsersByIDsResponse, error) {
	return nil, nil
}

func (m *mockAuthServiceClient) GetCompany(ctx context.Context, in *authpb.GetCompanyRequest, opts ...grpc.CallOption) (*authpb.GetCompanyResponse, error) {
	return nil, nil
}

func TestAuthClient_GetUserDeviceTokens(t *testing.T) {
	// Arrange
	mockService := &mockAuthServiceClient{
		tokens: []string{"token-1", "token-2"},
	}
	client := NewAuthClientWithService(mockService)

	// Act
	tokens, err := client.GetUserDeviceTokens(context.Background(), "user-123")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tokens) != 2 || tokens[0] != "token-1" {
		t.Errorf("expected ['token-1', 'token-2'], got %v", tokens)
	}
}

func TestAuthClient_GetUserDeviceTokens_EmptyUserID(t *testing.T) {
	// Arrange
	mockService := &mockAuthServiceClient{}
	client := NewAuthClientWithService(mockService)

	// Act
	tokens, err := client.GetUserDeviceTokens(context.Background(), "")

	// Assert
	if err == nil {
		t.Fatal("expected error when userID is empty, got nil")
	}
	if tokens != nil {
		t.Errorf("expected nil tokens on error, got %v", tokens)
	}
}

func TestAuthClient_RevokeDeviceTokens(t *testing.T) {
	// Arrange
	mockService := &mockAuthServiceClient{}
	client := NewAuthClientWithService(mockService)

	// Act
	err := client.RevokeDeviceTokens(context.Background(), "user-123", []string{"token-1"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error revoking tokens: %v", err)
	}
}

func TestAuthClient_RevokeDeviceTokens_Error(t *testing.T) {
	// Arrange
	mockService := &mockAuthServiceClient{revokeErr: errors.New("rpc failed")}
	client := NewAuthClientWithService(mockService)

	// Act
	err := client.RevokeDeviceTokens(context.Background(), "user-123", []string{"token-1"})

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
