package repositories_test

import (
	"context"
	"testing"

	"auth/internal/repositories"

	"github.com/stretchr/testify/assert"
)

func TestFirestoreDeviceRepository_NilClient(t *testing.T) {
	repo := repositories.NewFirestoreDeviceRepository(nil)
	assert.NotNil(t, repo)

	// RegisterDevice with nil client
	err := repo.RegisterDevice(context.Background(), "u1", "token1", "UA")
	assert.NoError(t, err)

	// GetUserTokens with nil client or empty userID
	tokens, err := repo.GetUserTokens(context.Background(), "u1")
	assert.NoError(t, err)
	assert.Nil(t, tokens)

	tokens, err = repo.GetUserTokens(context.Background(), "")
	assert.NoError(t, err)
	assert.Nil(t, tokens)

	// DeleteDeviceTokens with nil client, empty userID or empty tokens
	err = repo.DeleteDeviceTokens(context.Background(), "u1", []string{"tok"})
	assert.NoError(t, err)

	err = repo.DeleteDeviceTokens(context.Background(), "", []string{"tok"})
	assert.NoError(t, err)

	err = repo.DeleteDeviceTokens(context.Background(), "u1", nil)
	assert.NoError(t, err)
}
