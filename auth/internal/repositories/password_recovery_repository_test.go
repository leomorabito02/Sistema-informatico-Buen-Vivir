package repositories_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"auth/internal/models"
	"auth/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPasswordRecoveryRepository_CRUD(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewPasswordRecoveryRepository(db)

	rec := &models.PasswordRecovery{
		PasswordRecoveryID: "pr-1",
		UserID:             "u-1",
		TokenHash:          "hash123",
		ExpiresAt:          time.Now().Add(time.Hour),
	}

	// Create
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "password_recoveries"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), rec)
	assert.NoError(t, err)

	// FindByTokenHash - Success
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "password_recoveries" WHERE token_hash = $1`)).
		WithArgs("hash123", 1).
		WillReturnRows(sqlmock.NewRows([]string{"password_recovery_id", "token_hash"}).AddRow("pr-1", "hash123"))

	got, err := repo.FindByTokenHash(context.Background(), "hash123")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "pr-1", got.PasswordRecoveryID)

	// FindByTokenHash - NotFound
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "password_recoveries" WHERE token_hash = $1`)).
		WithArgs("hash404", 1).
		WillReturnRows(sqlmock.NewRows([]string{"password_recovery_id"}))

	got, err = repo.FindByTokenHash(context.Background(), "hash404")
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "token no encontrado")

	// MarkAsUsed
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "password_recoveries" SET`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.MarkAsUsed(context.Background(), "pr-1")
	assert.NoError(t, err)

	// DeleteByUserID
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "password_recoveries" WHERE user_id = $1`)).
		WithArgs("u-1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.DeleteByUserID(context.Background(), "u-1")
	assert.NoError(t, err)

	// DeleteExpired
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "password_recoveries" WHERE expires_at <`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.DeleteExpired(context.Background())
	assert.NoError(t, err)
}
