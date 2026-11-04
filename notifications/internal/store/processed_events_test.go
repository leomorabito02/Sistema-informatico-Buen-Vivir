package store_test

import (
	"context"
	"testing"

	"buenvivir/notifications/internal/store"
)

func TestMemoryProcessedStoreFallback(t *testing.T) {
	// Arrange
	s := store.NewProcessedStore(nil)
	ctx := context.Background()

	// Act 1: Initial check
	processed, err := s.IsProcessed(ctx, "c1", "ev-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if processed {
		t.Errorf("expected not processed, got true")
	}

	// Act 2: Mark processed
	err = s.MarkProcessed(ctx, "c1", "ev-1", "test.event")
	if err != nil {
		t.Fatalf("unexpected error on MarkProcessed: %v", err)
	}

	// Act 3: Check again
	processed, err = s.IsProcessed(ctx, "c1", "ev-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !processed {
		t.Errorf("expected processed, got false")
	}

	// Act 4: Different consumer
	processed, err = s.IsProcessed(ctx, "c2", "ev-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if processed {
		t.Errorf("expected c2 not processed, got true")
	}
}

