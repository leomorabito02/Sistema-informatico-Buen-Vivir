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
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	dialector := postgres.New(postgres.Config{
		Conn:       db,
		DriverName: "postgres",
	})
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewUserRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."user_id" LIMIT $2`)).
			WithArgs("test@test.com", 1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "email", "first_name"}).
				AddRow("usr-1", "test@test.com", "Test"))

		user, err := repo.FindByEmail(context.Background(), "test@test.com")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "usr-1", user.UserID)
		assert.Equal(t, "test@test.com", user.Email)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."user_id" LIMIT $2`)).
			WithArgs("notfound@test.com", 1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"})) // empty

		user, err := repo.FindByEmail(context.Background(), "notfound@test.com")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "usuario no encontrado")
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."user_id" LIMIT $2`)).
			WithArgs("error@test.com", 1).
			WillReturnError(errors.New("db error"))

		user, err := repo.FindByEmail(context.Background(), "error@test.com")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestUserRepository_FindAllByEmail(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewUserRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL`)).
			WithArgs("test@test.com").
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "email"}).
				AddRow("usr-1", "test@test.com").
				AddRow("usr-2", "test@test.com"))

		users, err := repo.FindAllByEmail(context.Background(), "test@test.com")
		assert.NoError(t, err)
		assert.Len(t, users, 2)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL`)).
			WithArgs("notfound@test.com").
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

		users, err := repo.FindAllByEmail(context.Background(), "notfound@test.com")
		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Contains(t, err.Error(), "usuario no encontrado")
	})
	
	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL`)).
			WithArgs("error@test.com").
			WillReturnError(errors.New("db error"))

		users, err := repo.FindAllByEmail(context.Background(), "error@test.com")
		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestUserRepository_FindAllByGoogleID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewUserRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE google_auth_id = $1 AND "users"."deleted_at" IS NULL`)).
			WithArgs("google-123").
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "google_auth_id"}).
				AddRow("usr-1", "google-123"))

		users, err := repo.FindAllByGoogleID(context.Background(), "google-123")
		assert.NoError(t, err)
		assert.Len(t, users, 1)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE google_auth_id = $1 AND "users"."deleted_at" IS NULL`)).
			WithArgs("notfound").
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

		users, err := repo.FindAllByGoogleID(context.Background(), "notfound")
		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Contains(t, err.Error(), "no se encontró ninguna cuenta")
	})
	
	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE google_auth_id = $1 AND "users"."deleted_at" IS NULL`)).
			WithArgs("error").
			WillReturnError(errors.New("db error"))

		users, err := repo.FindAllByGoogleID(context.Background(), "error")
		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewUserRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE user_id = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."user_id" LIMIT $2`)).
			WithArgs("usr-1", 1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "email"}).
				AddRow("usr-1", "test@test.com"))

		user, err := repo.FindByID(context.Background(), "usr-1")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "usr-1", user.UserID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE user_id = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."user_id" LIMIT $2`)).
			WithArgs("notfound", 1).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}))

		user, err := repo.FindByID(context.Background(), "notfound")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "usuario no encontrado")
	})
	
	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE user_id = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."user_id" LIMIT $2`)).
			WithArgs("error", 1).
			WillReturnError(errors.New("db error"))

		user, err := repo.FindByID(context.Background(), "error")
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestUserRepository_FindAll(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewUserRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL`)).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "email"}).
				AddRow("usr-1", "test@test.com").
				AddRow("usr-2", "test2@test.com"))

		users, err := repo.FindAll(context.Background())
		assert.NoError(t, err)
		assert.Len(t, users, 2)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."deleted_at" IS NULL`)).
			WillReturnError(errors.New("db error"))

		users, err := repo.FindAll(context.Background())
		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestUserRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewUserRepository(db)

	t.Run("Success", func(t *testing.T) {
		user := &models.User{
			UserID: "usr-1",
			Email:  "test@test.com",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "users"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Create(context.Background(), user)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		user := &models.User{
			UserID: "usr-1",
			Email:  "test@test.com",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "users"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Create(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestUserRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewUserRepository(db)

	t.Run("Success", func(t *testing.T) {
		user := &models.User{
			UserID: "usr-1",
			Email:  "test@test.com",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(context.Background(), user)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		user := &models.User{
			UserID: "usr-1",
			Email:  "test@test.com",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET`)).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Update(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestUserRepository_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewUserRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "deleted_at"=$1 WHERE user_id = $2 AND "users"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "usr-1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), "usr-1")
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "deleted_at"=$1 WHERE user_id = $2 AND "users"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "error").
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Delete(context.Background(), "error")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}
