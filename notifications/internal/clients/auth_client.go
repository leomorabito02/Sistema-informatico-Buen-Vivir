package clients

import (
	"context"
	"errors"
	"fmt"
	"time"

	"buenvivir/notifications/api/proto/authpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const defaultAuthTimeout = 5 * time.Second

// AuthClient abstracts gRPC communication with the Auth service for device token management.
type AuthClient interface {
	GetUserDeviceTokens(ctx context.Context, userID string) ([]string, error)
	RevokeDeviceTokens(ctx context.Context, userID string, tokens []string) error
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

func (c *authClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
