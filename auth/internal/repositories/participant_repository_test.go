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

func TestParticipantRepository_FindByID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewParticipantRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "participants" WHERE participant_id = $1 AND "participants"."deleted_at" IS NULL ORDER BY "participants"."participant_id" LIMIT $2`)).
			WithArgs("part-1", 1).
			WillReturnRows(sqlmock.NewRows([]string{"participant_id", "dni"}).
				AddRow("part-1", "12345678"))

		part, err := repo.FindByID(context.Background(), "part-1")
		assert.NoError(t, err)
		assert.NotNil(t, part)
		assert.Equal(t, "part-1", part.ParticipantID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "participants" WHERE participant_id = $1 AND "participants"."deleted_at" IS NULL ORDER BY "participants"."participant_id" LIMIT $2`)).
			WithArgs("notfound", 1).
			WillReturnRows(sqlmock.NewRows([]string{"participant_id"}))

		part, err := repo.FindByID(context.Background(), "notfound")
		assert.Error(t, err)
		assert.Nil(t, part)
		assert.Contains(t, err.Error(), "participante no encontrado")
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "participants" WHERE participant_id = $1 AND "participants"."deleted_at" IS NULL ORDER BY "participants"."participant_id" LIMIT $2`)).
			WithArgs("error", 1).
			WillReturnError(errors.New("db error"))

		part, err := repo.FindByID(context.Background(), "error")
		assert.Error(t, err)
		assert.Nil(t, part)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestParticipantRepository_FindAll(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewParticipantRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "participants" WHERE "participants"."deleted_at" IS NULL`)).
			WillReturnRows(sqlmock.NewRows([]string{"participant_id", "dni"}).
				AddRow("part-1", "1111").
				AddRow("part-2", "2222"))

		parts, err := repo.FindAll(context.Background())
		assert.NoError(t, err)
		assert.Len(t, parts, 2)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "participants" WHERE "participants"."deleted_at" IS NULL`)).
			WillReturnError(errors.New("db error"))

		parts, err := repo.FindAll(context.Background())
		assert.Error(t, err)
		assert.Nil(t, parts)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestParticipantRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewParticipantRepository(db)

	t.Run("Success", func(t *testing.T) {
		part := &models.Participant{
			ParticipantID: "part-1",
			DNI:           "123",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "participants"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Create(context.Background(), part)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		part := &models.Participant{
			ParticipantID: "part-1",
			DNI:           "123",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "participants"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Create(context.Background(), part)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestParticipantRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewParticipantRepository(db)

	t.Run("Success", func(t *testing.T) {
		part := &models.Participant{
			ParticipantID: "part-1",
			DNI:           "123",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "participants" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(context.Background(), part)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		part := &models.Participant{
			ParticipantID: "part-1",
			DNI:           "123",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "participants" SET`)).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Update(context.Background(), part)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestParticipantRepository_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewParticipantRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "participants" SET "deleted_at"=$1 WHERE participant_id = $2 AND "participants"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "part-1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), "part-1")
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "participants" SET "deleted_at"=$1 WHERE participant_id = $2 AND "participants"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "error").
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Delete(context.Background(), "error")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}
