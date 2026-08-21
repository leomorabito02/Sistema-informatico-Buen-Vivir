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

func TestVolunteerRepository_FindByID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewVolunteerRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "volunteers" WHERE volunteer_id = $1 AND "volunteers"."deleted_at" IS NULL ORDER BY "volunteers"."volunteer_id" LIMIT $2`)).
			WithArgs("vol-1", 1).
			WillReturnRows(sqlmock.NewRows([]string{"volunteer_id", "approval_status"}).
				AddRow("vol-1", "pending"))

		vol, err := repo.FindByID(context.Background(), "vol-1")
		assert.NoError(t, err)
		assert.NotNil(t, vol)
		assert.Equal(t, "vol-1", vol.VolunteerID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "volunteers" WHERE volunteer_id = $1 AND "volunteers"."deleted_at" IS NULL ORDER BY "volunteers"."volunteer_id" LIMIT $2`)).
			WithArgs("notfound", 1).
			WillReturnRows(sqlmock.NewRows([]string{"volunteer_id"}))

		vol, err := repo.FindByID(context.Background(), "notfound")
		assert.Error(t, err)
		assert.Nil(t, vol)
		assert.Contains(t, err.Error(), "voluntario no encontrado")
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "volunteers" WHERE volunteer_id = $1 AND "volunteers"."deleted_at" IS NULL ORDER BY "volunteers"."volunteer_id" LIMIT $2`)).
			WithArgs("error", 1).
			WillReturnError(errors.New("db error"))

		vol, err := repo.FindByID(context.Background(), "error")
		assert.Error(t, err)
		assert.Nil(t, vol)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestVolunteerRepository_FindAll(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewVolunteerRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "volunteers" WHERE "volunteers"."deleted_at" IS NULL`)).
			WillReturnRows(sqlmock.NewRows([]string{"volunteer_id"}).
				AddRow("vol-1").
				AddRow("vol-2"))

		vols, err := repo.FindAll(context.Background())
		assert.NoError(t, err)
		assert.Len(t, vols, 2)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "volunteers" WHERE "volunteers"."deleted_at" IS NULL`)).
			WillReturnError(errors.New("db error"))

		vols, err := repo.FindAll(context.Background())
		assert.Error(t, err)
		assert.Nil(t, vols)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestVolunteerRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewVolunteerRepository(db)

	t.Run("Success", func(t *testing.T) {
		vol := &models.Volunteer{
			VolunteerID: "vol-1",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "volunteers"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Create(context.Background(), vol)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		vol := &models.Volunteer{
			VolunteerID: "vol-1",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "volunteers"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Create(context.Background(), vol)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestVolunteerRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewVolunteerRepository(db)

	t.Run("Success", func(t *testing.T) {
		vol := &models.Volunteer{
			VolunteerID: "vol-1",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "volunteers" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(context.Background(), vol)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		vol := &models.Volunteer{
			VolunteerID: "vol-1",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "volunteers" SET`)).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Update(context.Background(), vol)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestVolunteerRepository_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewVolunteerRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "volunteers" SET "deleted_at"=$1 WHERE volunteer_id = $2 AND "volunteers"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "vol-1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), "vol-1")
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "volunteers" SET "deleted_at"=$1 WHERE volunteer_id = $2 AND "volunteers"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "error").
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Delete(context.Background(), "error")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}
