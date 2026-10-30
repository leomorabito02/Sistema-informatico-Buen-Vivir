package services

import (
	"context"
	"errors"
	"testing"

	"firebase.google.com/go/v4/messaging"
)

type mockFCMClient struct {
	batchResp *messaging.BatchResponse
	err       error
}

func (m *mockFCMClient) SendEachForMulticast(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.batchResp, nil
}

func TestSendMulticast_EmptyTokens(t *testing.T) {
	// Arrange
	mockClient := &mockFCMClient{}
	svc, err := NewPushServiceWithClient(mockClient)
	if err != nil {
		t.Fatalf("unexpected init error: %v", err)
	}

	// Act
	res, err := svc.SendMulticast(context.Background(), []string{}, "Title", "Body", nil)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SuccessCount != 0 || len(res.InvalidTokens) != 0 {
		t.Errorf("expected empty result, got %+v", res)
	}
}

func TestSendMulticast_SuccessAndInvalidTokens(t *testing.T) {
	// Arrange
	unregisteredErr := errors.New("registration-token-not-registered: token is no longer valid")
	otherErr := errors.New("internal server error")

	mockClient := &mockFCMClient{
		batchResp: &messaging.BatchResponse{
			SuccessCount: 1,
			FailureCount: 2,
			Responses: []*messaging.SendResponse{
				{
					Success:   true,
					MessageID: "projects/test/messages/msg-1",
				},
				{
					Success: false,
					Error:   unregisteredErr,
				},
				{
					Success: false,
					Error:   otherErr,
				},
			},
		},
	}

	svc, err := NewPushServiceWithClient(mockClient)
	if err != nil {
		t.Fatalf("unexpected init error: %v", err)
	}

	tokens := []string{"token-valid", "token-unregistered", "token-other-err"}

	// Act
	res, err := svc.SendMulticast(context.Background(), tokens, "Title", "Body", map[string]string{"key": "value"})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.SuccessCount != 1 {
		t.Errorf("expected 1 success, got %d", res.SuccessCount)
	}
	if res.FailureCount != 2 {
		t.Errorf("expected 2 failures, got %d", res.FailureCount)
	}
	if len(res.InvalidTokens) != 1 || res.InvalidTokens[0] != "token-unregistered" {
		t.Errorf("expected ['token-unregistered'], got %v", res.InvalidTokens)
	}
}

func TestSendMulticast_ClientError(t *testing.T) {
	// Arrange
	expectedErr := errors.New("network failure")
	mockClient := &mockFCMClient{err: expectedErr}
	svc, err := NewPushServiceWithClient(mockClient)
	if err != nil {
		t.Fatalf("unexpected init error: %v", err)
	}

	// Act
	res, err := svc.SendMulticast(context.Background(), []string{"token-1"}, "Title", "Body", nil)

	// Assert
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if res != nil {
		t.Errorf("expected nil result on error, got %+v", res)
	}
}
