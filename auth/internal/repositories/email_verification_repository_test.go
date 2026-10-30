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

func TestEmailVerificationRepository_Create(t *testing.T) {
	// Arrange
	db, mock := setupTestDB(t)
	repo := repositories.NewEmailVerificationRepository(db)
	ev := &models.EmailVerification{
		EmailVerificationID: "ev-1",
		UserID:              "usr-1",
		TokenHash:           "hash-xyz",
		ExpiresAt:           time.Now().Add(24 * time.Hour),
	}

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "email_verifications"`)).
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Act
	err := repo.Create(context.Background(), ev)

	// Assert
	assert.NoError(t, err)
}

func TestEmailVerificationRepository_FindByTokenHash(t *testing.T) {
	// Arrange
	db, mock := setupTestDB(t)
	repo := repositories.NewEmailVerificationRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "email_verifications" WHERE token_hash = $1 ORDER BY "email_verifications"."email_verification_id" LIMIT $2`)).
			WithArgs("hash-xyz", 1).
			WillReturnRows(sqlmock.NewRows([]string{"email_verification_id", "user_id", "token_hash"}).
				AddRow("ev-1", "usr-1", "hash-xyz"))

		// Act
		ev, err := repo.FindByTokenHash(context.Background(), "hash-xyz")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, ev)
		assert.Equal(t, "ev-1", ev.EmailVerificationID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "email_verifications" WHERE token_hash = $1 ORDER BY "email_verifications"."email_verification_id" LIMIT $2`)).
			WithArgs("not-found", 1).
			WillReturnRows(sqlmock.NewRows([]string{"email_verification_id"}))

		// Act
		ev, err := repo.FindByTokenHash(context.Background(), "not-found")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, ev)
		assert.Contains(t, err.Error(), "token no encontrado")
	})
}

func TestEmailVerificationRepository_MarkAsUsed(t *testing.T) {
	// Arrange
	db, mock := setupTestDB(t)
	repo := repositories.NewEmailVerificationRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "email_verifications" SET "used_at"=$1,"updated_at"=$2 WHERE email_verification_id = $3`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "ev-1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Act
	err := repo.MarkAsUsed(context.Background(), "ev-1")

	// Assert
	assert.NoError(t, err)
}

func TestEmailVerificationRepository_Delete(t *testing.T) {
	// Arrange
	db, mock := setupTestDB(t)
	repo := repositories.NewEmailVerificationRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "email_verifications" WHERE email_verification_id = $1`)).
		WithArgs("ev-1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Act
	err := repo.Delete(context.Background(), "ev-1")

	// Assert
	assert.NoError(t, err)
}

func TestEmailVerificationRepository_DeleteExpired(t *testing.T) {
	// Arrange
	db, mock := setupTestDB(t)
	repo := repositories.NewEmailVerificationRepository(db)

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "email_verifications" WHERE expires_at < $1 OR used_at IS NOT NULL`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Act
	err := repo.DeleteExpired(context.Background())

	// Assert
	assert.NoError(t, err)
}
