package clients

import (
	"context"
	"errors"
	"testing"

	"buenvivir/notifications/api/proto/authpb"
	"google.golang.org/grpc"
)

type mockAuthServiceClient struct {
	tokens                   []string
	getTokensErr             error
	revokeErr                error
	campaignRecipients       []*authpb.CampaignRecipient
	getCampaignRecipientsErr error
	isVerified               bool
	checkEmailVerifiedErr    error
	userContact              *authpb.GetUserContactResponse
	getUserContactErr        error
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

func (m *mockAuthServiceClient) GetCampaignRecipients(ctx context.Context, in *authpb.GetCampaignRecipientsRequest, opts ...grpc.CallOption) (*authpb.GetCampaignRecipientsResponse, error) {
	if m.getCampaignRecipientsErr != nil {
		return nil, m.getCampaignRecipientsErr
	}
	return &authpb.GetCampaignRecipientsResponse{Recipients: m.campaignRecipients}, nil
}

func (m *mockAuthServiceClient) CheckUserEmailVerified(ctx context.Context, in *authpb.CheckUserEmailVerifiedRequest, opts ...grpc.CallOption) (*authpb.CheckUserEmailVerifiedResponse, error) {
	if m.checkEmailVerifiedErr != nil {
		return nil, m.checkEmailVerifiedErr
	}
	return &authpb.CheckUserEmailVerifiedResponse{IsVerified: m.isVerified}, nil
}

func (m *mockAuthServiceClient) GetUserContact(ctx context.Context, in *authpb.GetUserContactRequest, opts ...grpc.CallOption) (*authpb.GetUserContactResponse, error) {
	if m.getUserContactErr != nil {
		return nil, m.getUserContactErr
	}
	if m.userContact != nil {
		return m.userContact, nil
	}
	return nil, errors.New("user not found")
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

func TestAuthClient_GetCampaignRecipients_Success(t *testing.T) {
	mockService := &mockAuthServiceClient{
		campaignRecipients: []*authpb.CampaignRecipient{
			{UserId: "usr-1", Email: "u1@test.com", FirstName: "Ana", LastName: "Gomez"},
			{UserId: "usr-2", Email: "u2@test.com", FirstName: "Carlos", LastName: "Perez"},
		},
	}
	client := NewAuthClientWithService(mockService)

	recipients, err := client.GetCampaignRecipients(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(recipients) != 2 {
		t.Fatalf("expected 2 recipients, got %d", len(recipients))
	}
	if recipients[0].UserID != "usr-1" || recipients[0].Email != "u1@test.com" {
		t.Errorf("unexpected recipient 0: %+v", recipients[0])
	}
}

func TestAuthClient_GetCampaignRecipients_Error(t *testing.T) {
	mockService := &mockAuthServiceClient{
		getCampaignRecipientsErr: errors.New("rpc network failure"),
	}
	client := NewAuthClientWithService(mockService)

	_, err := client.GetCampaignRecipients(context.Background())
	if err == nil {
		t.Fatal("expected error from GetCampaignRecipients, got nil")
	}
}

func TestAuthClient_IsUserEmailVerified(t *testing.T) {
	t.Run("EmptyUserID", func(t *testing.T) {
		client := NewAuthClientWithService(&mockAuthServiceClient{})
		_, err := client.IsUserEmailVerified(context.Background(), "")
		if err == nil {
			t.Fatal("expected error on empty user_id")
		}
	})

	t.Run("Verified", func(t *testing.T) {
		client := NewAuthClientWithService(&mockAuthServiceClient{isVerified: true})
		verified, err := client.IsUserEmailVerified(context.Background(), "usr-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !verified {
			t.Errorf("expected isVerified true, got false")
		}
	})

	t.Run("Unverified", func(t *testing.T) {
		client := NewAuthClientWithService(&mockAuthServiceClient{isVerified: false})
		verified, err := client.IsUserEmailVerified(context.Background(), "usr-2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if verified {
			t.Errorf("expected isVerified false, got true")
		}
	})

	t.Run("RPCError", func(t *testing.T) {
		client := NewAuthClientWithService(&mockAuthServiceClient{checkEmailVerifiedErr: errors.New("auth down")})
		_, err := client.IsUserEmailVerified(context.Background(), "usr-1")
		if err == nil {
			t.Fatal("expected error from IsUserEmailVerified, got nil")
		}
	})
}

func TestAuthClient_GetUserContact(t *testing.T) {
	t.Run("EmptyUserID", func(t *testing.T) {
		client := NewAuthClientWithService(&mockAuthServiceClient{})
		_, err := client.GetUserContact(context.Background(), "")
		if err == nil {
			t.Fatal("expected error on empty user_id")
		}
	})

	t.Run("Success", func(t *testing.T) {
		mockService := &mockAuthServiceClient{
			userContact: &authpb.GetUserContactResponse{
				UserId:          "usr-10",
				Email:           "diego@example.com",
				FirstName:       "Diego",
				LastName:        "Armando",
				IsEmailVerified: true,
			},
		}
		client := NewAuthClientWithService(mockService)
		contact, err := client.GetUserContact(context.Background(), "usr-10")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if contact.Email != "diego@example.com" {
			t.Errorf("expected diego@example.com, got %s", contact.Email)
		}
		if contact.FirstName != "Diego" || contact.LastName != "Armando" {
			t.Errorf("expected Diego Armando, got %s %s", contact.FirstName, contact.LastName)
		}
		if !contact.IsEmailVerified {
			t.Errorf("expected IsEmailVerified true, got false")
		}
	})

	t.Run("RPCError", func(t *testing.T) {
		mockService := &mockAuthServiceClient{
			getUserContactErr: errors.New("grpc unavailable"),
		}
		client := NewAuthClientWithService(mockService)
		_, err := client.GetUserContact(context.Background(), "usr-10")
		if err == nil {
			t.Fatal("expected error from GetUserContact, got nil")
		}
	})
}


