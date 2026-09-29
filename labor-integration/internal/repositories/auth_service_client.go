package repositories

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"buenvivir/labor-integration/api/proto/authpb"
)

type AuthServiceClient interface {
	GetParticipant(ctx context.Context, participantID string) (*authpb.GetParticipantResponse, error)
	Close() error
}

type authServiceClient struct {
	conn   *grpc.ClientConn
	client authpb.AuthServiceClient
}

func NewAuthServiceClient(target string) (AuthServiceClient, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth-service: %w", err)
	}

	return &authServiceClient{
		conn:   conn,
		client: authpb.NewAuthServiceClient(conn),
	}, nil
}

func (c *authServiceClient) GetParticipant(ctx context.Context, participantID string) (*authpb.GetParticipantResponse, error) {
	req := &authpb.GetParticipantRequest{
		ParticipantId: participantID,
	}
	return c.client.GetParticipant(ctx, req)
}

func (c *authServiceClient) Close() error {
	return c.conn.Close()
}
