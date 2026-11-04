package clients

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"buenvivir/notifications/api/proto/authpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultAuthTimeout = 5 * time.Second

// RecipientContact represents contact details for a campaign recipient.
type RecipientContact struct {
	UserID          string
	Email           string
	FirstName       string
	LastName        string
	IsEmailVerified bool
}

// AuthClient abstracts gRPC communication with the Auth service for device token management and recipient queries.
type AuthClient interface {
	GetUserDeviceTokens(ctx context.Context, userID string) ([]string, error)
	RevokeDeviceTokens(ctx context.Context, userID string, tokens []string) error
	GetCampaignRecipients(ctx context.Context) ([]RecipientContact, error)
	IsUserEmailVerified(ctx context.Context, userID string) (bool, error)
	GetUserContact(ctx context.Context, userID string) (*RecipientContact, error)
	GetActiveUserIDsByRoles(ctx context.Context, roleIDs []string) ([]string, error)
	GetParticipantUserID(ctx context.Context, participantID string) (string, error)
	IssueActionLink(ctx context.Context, requestID, userID, actionType string) (string, error)
	Close() error
}

type authClient struct {
	conn   *grpc.ClientConn
	client authpb.AuthServiceClient
}

// NewAuthClient establishes a gRPC client connection to the Auth service.
func NewAuthClient(target string) (AuthClient, error) {
	if target == "" {
		target = "localhost:50051"
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}
	return &authClient{
		conn:   conn,
		client: authpb.NewAuthServiceClient(conn),
	}, nil
}

// NewAuthClientWithService creates an AuthClient using an existing AuthServiceClient (useful for testing).
func NewAuthClientWithService(client authpb.AuthServiceClient) AuthClient {
	return &authClient{client: client}
}

func (c *authClient) GetUserDeviceTokens(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, errors.New("userID cannot be empty")
	}

	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.GetUserDeviceTokens(callCtx, &authpb.GetUserDeviceTokensRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user device tokens: %w", err)
	}
	return resp.Tokens, nil
}

func (c *authClient) RevokeDeviceTokens(ctx context.Context, userID string, tokens []string) error {
	if userID == "" || len(tokens) == 0 {
		return nil
	}

	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	_, err := c.client.RevokeDeviceTokens(callCtx, &authpb.RevokeDeviceTokensRequest{
		UserId: userID,
		Tokens: tokens,
	})
	if err != nil {
		return fmt.Errorf("failed to revoke device tokens: %w", err)
	}
	return nil
}

func (c *authClient) GetCampaignRecipients(ctx context.Context) ([]RecipientContact, error) {
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.GetCampaignRecipients(callCtx, &authpb.GetCampaignRecipientsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve campaign recipients: %w", err)
	}

	out := make([]RecipientContact, 0, len(resp.Recipients))
	for _, r := range resp.Recipients {
		out = append(out, RecipientContact{
			UserID:          r.UserId,
			Email:           r.Email,
			FirstName:       r.FirstName,
			LastName:        r.LastName,
			IsEmailVerified: r.IsEmailVerified,
		})
	}
	return out, nil
}

func (c *authClient) IsUserEmailVerified(ctx context.Context, userID string) (bool, error) {
	if strings.TrimSpace(userID) == "" {
		return false, errors.New("user_id cannot be empty")
	}
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.CheckUserEmailVerified(callCtx, &authpb.CheckUserEmailVerifiedRequest{UserId: userID})
	if err != nil {
		return false, fmt.Errorf("failed to check email verification: %w", err)
	}
	return resp.IsVerified, nil
}

func (c *authClient) GetUserContact(ctx context.Context, userID string) (*RecipientContact, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, errors.New("user_id cannot be empty")
	}
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.GetUserContact(callCtx, &authpb.GetUserContactRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user contact: %w", err)
	}
	return &RecipientContact{
		UserID:          resp.UserId,
		Email:           resp.Email,
		FirstName:       resp.FirstName,
		LastName:        resp.LastName,
		IsEmailVerified: resp.IsEmailVerified,
	}, nil
}

func (c *authClient) GetActiveUserIDsByRoles(ctx context.Context, roleIDs []string) ([]string, error) {
	if len(roleIDs) == 0 {
		return []string{}, nil
	}
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.GetActiveUserIDsByRoles(callCtx, &authpb.GetActiveUserIDsByRolesRequest{RoleIds: roleIDs})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user IDs by roles: %w", err)
	}
	return resp.UserIds, nil
}

func (c *authClient) GetParticipantUserID(ctx context.Context, participantID string) (string, error) {
	if strings.TrimSpace(participantID) == "" {
		return "", errors.New("participant_id cannot be empty")
	}
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.GetParticipant(callCtx, &authpb.GetParticipantRequest{ParticipantId: participantID})
	if err != nil {
		return "", fmt.Errorf("failed to retrieve participant: %w", err)
	}
	return resp.UserId, nil
}

func (c *authClient) IssueActionLink(ctx context.Context, requestID, userID, actionType string) (string, error) {
	callCtx, cancel := context.WithTimeout(ctx, defaultAuthTimeout)
	defer cancel()

	resp, err := c.client.IssueActionLink(callCtx, &authpb.IssueActionLinkRequest{
		RequestId: requestID,
		UserId:    userID,
		Type:      actionType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to issue action link: %w", err)
	}
	return resp.ActionUrl, nil
}

func (c *authClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
