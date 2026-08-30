package repositories_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"auth/internal/models"
	"auth/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestRefreshTokenRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewRefreshTokenRepository(db)

	t.Run("Success", func(t *testing.T) {
		token := &models.RefreshToken{
			RefreshTokenID: "token-1",
			TokenHash:      "hash-123",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "refresh_tokens"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Create(context.Background(), token)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		token := &models.RefreshToken{
			RefreshTokenID: "token-1",
			TokenHash:      "hash-123",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "refresh_tokens"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Create(context.Background(), token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestRefreshTokenRepository_FindByTokenHash(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewRefreshTokenRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "refresh_tokens" WHERE token_hash = $1 ORDER BY "refresh_tokens"."refresh_token_id" LIMIT $2`)).
			WithArgs("hash-123", 1).
			WillReturnRows(sqlmock.NewRows([]string{"refresh_token_id", "token_hash"}).
				AddRow("token-1", "hash-123"))

		token, err := repo.FindByTokenHash(context.Background(), "hash-123")
		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, "token-1", token.RefreshTokenID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "refresh_tokens" WHERE token_hash = $1 ORDER BY "refresh_tokens"."refresh_token_id" LIMIT $2`)).
			WithArgs("notfound", 1).
			WillReturnRows(sqlmock.NewRows([]string{"refresh_token_id"}))

		token, err := repo.FindByTokenHash(context.Background(), "notfound")
		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "refresh token not found")
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "refresh_tokens" WHERE token_hash = $1 ORDER BY "refresh_tokens"."refresh_token_id" LIMIT $2`)).
			WithArgs("error", 1).
			WillReturnError(errors.New("db error"))

		token, err := repo.FindByTokenHash(context.Background(), "error")
		assert.Error(t, err)
		assert.Nil(t, token)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestRefreshTokenRepository_Revoke(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewRefreshTokenRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "refresh_tokens" SET "revoked_at"=$1,"updated_at"=$2 WHERE refresh_token_id = $3`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "token-1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Revoke(context.Background(), "token-1")
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "refresh_tokens" SET "revoked_at"=$1,"updated_at"=$2 WHERE refresh_token_id = $3`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "error").
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Revoke(context.Background(), "error")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestRefreshTokenRepository_RevokeAllForUser(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewRefreshTokenRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "refresh_tokens" SET "revoked_at"=$1,"updated_at"=$2 WHERE user_id = $3 AND revoked_at IS NULL`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "user-1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.RevokeAllForUser(context.Background(), "user-1")
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "refresh_tokens" SET "revoked_at"=$1,"updated_at"=$2 WHERE user_id = $3 AND revoked_at IS NULL`)).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), "error").
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.RevokeAllForUser(context.Background(), "error")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestRefreshTokenRepository_DeleteExpired(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewRefreshTokenRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "refresh_tokens" WHERE expires_at < $1 OR revoked_at IS NOT NULL`)).
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.DeleteExpired(context.Background())
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "refresh_tokens" WHERE expires_at < $1 OR revoked_at IS NOT NULL`)).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.DeleteExpired(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}
