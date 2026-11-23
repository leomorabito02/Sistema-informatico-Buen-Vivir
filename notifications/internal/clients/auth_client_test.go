package clients

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
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
	activeUserIDs            []string
	getActiveUserIDsErr      error
	participantUserID        string
	getParticipantErr        error
	actionURL                string
	issueActionLinkErr       error
}

func (m *mockAuthServiceClient) IssueActionLink(ctx context.Context, in *authpb.IssueActionLinkRequest, opts ...grpc.CallOption) (*authpb.IssueActionLinkResponse, error) {
	if m.issueActionLinkErr != nil {
		return nil, m.issueActionLinkErr
	}
	if m.actionURL != "" {
		return &authpb.IssueActionLinkResponse{ActionUrl: m.actionURL}, nil
	}
	return &authpb.IssueActionLinkResponse{ActionUrl: "https://ceibuenvivir.org/action"}, nil
}

func (m *mockAuthServiceClient) GetParticipant(ctx context.Context, in *authpb.GetParticipantRequest, opts ...grpc.CallOption) (*authpb.GetParticipantResponse, error) {
	if m.getParticipantErr != nil {
		return nil, m.getParticipantErr
	}
	return &authpb.GetParticipantResponse{Id: in.ParticipantId, UserId: m.participantUserID}, nil
}

func (m *mockAuthServiceClient) GetActiveUserIDsByRoles(ctx context.Context, in *authpb.GetActiveUserIDsByRolesRequest, opts ...grpc.CallOption) (*authpb.GetActiveUserIDsByRolesResponse, error) {
	if m.getActiveUserIDsErr != nil {
		return nil, m.getActiveUserIDsErr
	}
	return &authpb.GetActiveUserIDsByRolesResponse{UserIds: m.activeUserIDs}, nil
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

func (m *mockAuthServiceClient) GetParticipantProfile(ctx context.Context, in *authpb.GetParticipantProfileRequest, opts ...grpc.CallOption) (*authpb.ParticipantProfile, error) {
	return nil, nil
}

func (m *mockAuthServiceClient) SearchParticipants(ctx context.Context, in *authpb.SearchParticipantsRequest, opts ...grpc.CallOption) (*authpb.SearchParticipantsResponse, error) {
	return nil, nil
}

func (m *mockAuthServiceClient) GetUsersByIDs(ctx context.Context, in *authpb.GetUsersByIDsRequest, opts ...grpc.CallOption) (*authpb.GetUsersByIDsResponse, error) {
	return nil, nil
}

func (m *mockAuthServiceClient) ListParticipants(ctx context.Context, in *authpb.ListParticipantsRequest, opts ...grpc.CallOption) (*authpb.ListParticipantsResponse, error) {
	return nil, nil
}

func (m *mockAuthServiceClient) ListUsers(ctx context.Context, in *authpb.ListUsersRequest, opts ...grpc.CallOption) (*authpb.GetUsersByIDsResponse, error) {
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

func TestAuthClient_GetActiveUserIDsByRoles(t *testing.T) {
	t.Run("EmptyRoles", func(t *testing.T) {
		client := NewAuthClientWithService(&mockAuthServiceClient{})
		res, err := client.GetActiveUserIDsByRoles(context.Background(), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != 0 {
			t.Fatalf("expected 0 user IDs, got %d", len(res))
		}
	})

	t.Run("Success", func(t *testing.T) {
		mockService := &mockAuthServiceClient{
			activeUserIDs: []string{"usr-admin-1", "usr-admin-2"},
		}
		client := NewAuthClientWithService(mockService)
		res, err := client.GetActiveUserIDsByRoles(context.Background(), []string{"role_admin"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(res) != 2 || res[0] != "usr-admin-1" {
			t.Fatalf("unexpected response: %v", res)
		}
	})

	t.Run("RPCError", func(t *testing.T) {
		mockService := &mockAuthServiceClient{
			getActiveUserIDsErr: errors.New("connection refused"),
		}
		client := NewAuthClientWithService(mockService)
		_, err := client.GetActiveUserIDsByRoles(context.Background(), []string{"role_admin"})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestAuthClient_GetParticipantUserID(t *testing.T) {
	t.Run("EmptyParticipantID", func(t *testing.T) {
		client := NewAuthClientWithService(&mockAuthServiceClient{})
		_, err := client.GetParticipantUserID(context.Background(), "")
		if err == nil {
			t.Fatal("expected error on empty participantID")
		}
	})

	t.Run("Success", func(t *testing.T) {
		client := NewAuthClientWithService(&mockAuthServiceClient{participantUserID: "usr-participant-10"})
		uid, err := client.GetParticipantUserID(context.Background(), "part-99")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if uid != "usr-participant-10" {
			t.Errorf("expected usr-participant-10, got %s", uid)
		}
	})

	t.Run("RPCError", func(t *testing.T) {
		client := NewAuthClientWithService(&mockAuthServiceClient{getParticipantErr: errors.New("participant not found")})
		_, err := client.GetParticipantUserID(context.Background(), "part-99")
		if err == nil {
			t.Fatal("expected error from GetParticipantUserID, got nil")
		}
	})
}

func TestAuthClient_IssueActionLink(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		client := NewAuthClientWithService(&mockAuthServiceClient{actionURL: "https://ceibuenvivir.org/verify?token=xyz"})
		link, err := client.IssueActionLink(context.Background(), "req-1", "usr-1", "verificacion")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if link != "https://ceibuenvivir.org/verify?token=xyz" {
			t.Errorf("expected link https://ceibuenvivir.org/verify?token=xyz, got %s", link)
		}
	})

	t.Run("Error", func(t *testing.T) {
		client := NewAuthClientWithService(&mockAuthServiceClient{issueActionLinkErr: errors.New("rpc failed")})
		_, err := client.IssueActionLink(context.Background(), "req-1", "usr-1", "verificacion")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestHTTPAuthClient_S2SHeaders(t *testing.T) {
	t.Run("CloudRunOIDCInAuthorizationAndServiceJWTInXService", func(t *testing.T) {
		// Arrange
		var gotAuth, gotServerless, gotService, gotForwarded string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			gotServerless = r.Header.Get("X-Serverless-Authorization")
			gotService = r.Header.Get("X-Service-Authorization")
			gotForwarded = r.Header.Get("X-Forwarded-Authorization")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"user_id":"usr-1","email":"u@test.com","is_email_verified":false}`))
		}))
		defer srv.Close()

		client, err := NewAuthClient(
			srv.URL,
			WithJWTSecret("shared-secret"),
			WithIDToken(func(context.Context) (string, error) { return "oidc-id-token", nil }),
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = client.Close() }()

		// Act
		contact, err := client.GetUserContact(context.Background(), "usr-1")

		// Assert
		if err != nil {
			t.Fatalf("unexpected GetUserContact error: %v", err)
		}
		if contact.Email != "u@test.com" {
			t.Fatalf("unexpected email: %s", contact.Email)
		}
		if gotAuth != "Bearer oidc-id-token" {
			t.Errorf("expected OIDC token in Authorization, got %q", gotAuth)
		}
		if gotServerless != "Bearer oidc-id-token" {
			t.Errorf("expected OIDC token in X-Serverless-Authorization, got %q", gotServerless)
		}
		if gotService == "" || gotService == "Bearer oidc-id-token" {
			t.Errorf("expected internal HS256 JWT in X-Service-Authorization, got %q", gotService)
		}
		if gotForwarded != gotService {
			t.Errorf("expected X-Forwarded-Authorization to match X-Service-Authorization")
		}
	})

	t.Run("LocalWithoutIDTokenUsesServiceJWTInAuthorization", func(t *testing.T) {
		// Arrange
		var gotAuth, gotService string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			gotService = r.Header.Get("X-Service-Authorization")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"action_url":"https://testing.ceibuenvivir.pages.dev/verify-email?token=abc"}`))
		}))
		defer srv.Close()

		client, err := NewAuthClient(srv.URL, WithJWTSecret("shared-secret"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defer func() { _ = client.Close() }()

		// Act
		url, err := client.IssueActionLink(context.Background(), "req-1", "usr-1", "verificacion")

		// Assert
		if err != nil {
			t.Fatalf("unexpected IssueActionLink error: %v", err)
		}
		if url == "" {
			t.Fatal("expected action URL")
		}
		if gotAuth == "" || gotAuth != gotService {
			t.Errorf("expected service JWT in both Authorization (%q) and X-Service-Authorization (%q)", gotAuth, gotService)
		}
	})
}
