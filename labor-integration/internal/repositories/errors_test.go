package repositories

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNotFoundAs(t *testing.T) {
	if got := notFoundAs(status.Error(codes.NotFound, "no doc"), ErrApplicationNotFound); !errors.Is(got, ErrApplicationNotFound) {
		t.Errorf("NotFound de Firestore: got %v, quiero ErrApplicationNotFound", got)
	}
	otro := status.Error(codes.Unavailable, "caído")
	if got := notFoundAs(otro, ErrApplicationNotFound); got != otro {
		t.Errorf("otro error: got %v, quiero el mismo error", got)
	}
}
