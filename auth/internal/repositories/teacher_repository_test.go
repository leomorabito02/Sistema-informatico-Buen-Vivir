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

func TestTeacherRepository_FindByID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewTeacherRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "teachers" WHERE teacher_id = $1 AND "teachers"."deleted_at" IS NULL ORDER BY "teachers"."teacher_id" LIMIT $2`)).
			WithArgs("teacher-1", 1).
			WillReturnRows(sqlmock.NewRows([]string{"teacher_id", "specialization"}).
				AddRow("teacher-1", "Math"))

		teacher, err := repo.FindByID(context.Background(), "teacher-1")
		assert.NoError(t, err)
		assert.NotNil(t, teacher)
		assert.Equal(t, "teacher-1", teacher.TeacherID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "teachers" WHERE teacher_id = $1 AND "teachers"."deleted_at" IS NULL ORDER BY "teachers"."teacher_id" LIMIT $2`)).
			WithArgs("notfound", 1).
			WillReturnRows(sqlmock.NewRows([]string{"teacher_id"}))

		teacher, err := repo.FindByID(context.Background(), "notfound")
		assert.Error(t, err)
		assert.Nil(t, teacher)
		assert.Contains(t, err.Error(), "profesor no encontrado")
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "teachers" WHERE teacher_id = $1 AND "teachers"."deleted_at" IS NULL ORDER BY "teachers"."teacher_id" LIMIT $2`)).
			WithArgs("error", 1).
			WillReturnError(errors.New("db error"))

		teacher, err := repo.FindByID(context.Background(), "error")
		assert.Error(t, err)
		assert.Nil(t, teacher)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestTeacherRepository_FindAll(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewTeacherRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "teachers" WHERE "teachers"."deleted_at" IS NULL`)).
			WillReturnRows(sqlmock.NewRows([]string{"teacher_id", "specialization"}).
				AddRow("teacher-1", "Math").
				AddRow("teacher-2", "Science"))

		teachers, err := repo.FindAll(context.Background())
		assert.NoError(t, err)
		assert.Len(t, teachers, 2)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "teachers" WHERE "teachers"."deleted_at" IS NULL`)).
			WillReturnError(errors.New("db error"))

		teachers, err := repo.FindAll(context.Background())
		assert.Error(t, err)
		assert.Nil(t, teachers)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestTeacherRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewTeacherRepository(db)

	t.Run("Success", func(t *testing.T) {
		teacher := &models.Teacher{
			TeacherID: "teacher-1",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "teachers"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Create(context.Background(), teacher)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		teacher := &models.Teacher{
			TeacherID: "teacher-1",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "teachers"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Create(context.Background(), teacher)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestTeacherRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewTeacherRepository(db)

	t.Run("Success", func(t *testing.T) {
		teacher := &models.Teacher{
			TeacherID: "teacher-1",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "teachers" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(context.Background(), teacher)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		teacher := &models.Teacher{
			TeacherID: "teacher-1",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "teachers" SET`)).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Update(context.Background(), teacher)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestTeacherRepository_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewTeacherRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "teachers" SET "deleted_at"=$1 WHERE teacher_id = $2 AND "teachers"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "teacher-1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), "teacher-1")
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "teachers" SET "deleted_at"=$1 WHERE teacher_id = $2 AND "teachers"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "error").
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Delete(context.Background(), "error")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}
