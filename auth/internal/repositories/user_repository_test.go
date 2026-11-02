package repositories_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

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
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT users.*, COALESCE(p.approval_status, c.approval_status, t.approval_status, v.approval_status) AS role_approval_status, p.image_url AS image_url FROM "users" LEFT JOIN participants p ON users.user_id = p.user_id LEFT JOIN companies c ON users.user_id = c.user_id LEFT JOIN teachers t ON users.user_id = t.user_id LEFT JOIN volunteers v ON users.user_id = v.user_id WHERE "users"."deleted_at" IS NULL`)).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "email"}).
				AddRow("usr-1", "test@test.com").
				AddRow("usr-2", "test2@test.com"))

		users, err := repo.FindAll(context.Background())
		assert.NoError(t, err)
		assert.Len(t, users, 2)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT users.*, COALESCE(p.approval_status, c.approval_status, t.approval_status, v.approval_status) AS role_approval_status, p.image_url AS image_url FROM "users" LEFT JOIN participants p ON users.user_id = p.user_id LEFT JOIN companies c ON users.user_id = c.user_id LEFT JOIN teachers t ON users.user_id = t.user_id LEFT JOIN volunteers v ON users.user_id = v.user_id WHERE "users"."deleted_at" IS NULL`)).
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
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
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
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Create(context.Background(), user)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestUserRepository_UpdateEmailVerified(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewUserRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "is_email_verified"=$1,"updated_at"=$2 WHERE user_id = $3 AND "users"."deleted_at" IS NULL`)).
			WithArgs(true, sqlmock.AnyArg(), "usr-1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateEmailVerified(context.Background(), "usr-1", true)
		assert.NoError(t, err)
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

func TestUserRepository_FindParticipantsHistory(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewUserRepository(db)
	query := regexp.QuoteMeta(`SELECT u.user_id, u.status, u.created_at, u.updated_at, u.deleted_at,`)

	t.Run("IncluyeBajas", func(t *testing.T) {
		alta := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
		baja := time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC)
		mock.ExpectQuery(query).WillReturnRows(
			sqlmock.NewRows([]string{"user_id", "status", "created_at", "updated_at", "deleted_at", "approval_status", "approval_date"}).
				AddRow("usr-1", "active", alta, alta, nil, "approved", alta).
				AddRow("usr-2", "active", alta, alta, baja, "approved", nil))

		rows, err := repo.FindParticipantsHistory(context.Background())
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.Nil(t, rows[0].DeletedAt)
		require.NotNil(t, rows[1].DeletedAt)
		assert.True(t, baja.Equal(*rows[1].DeletedAt))
		assert.Nil(t, rows[1].ApprovalDate)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(query).WillReturnError(errors.New("db error"))
		rows, err := repo.FindParticipantsHistory(context.Background())
		assert.Error(t, err)
		assert.Nil(t, rows)
	})
}

func TestUserRepository_FindCampaignEligibleUsers(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewUserRepository(db)
	query := regexp.QuoteMeta(`SELECT users.user_id, users.email, users.first_name, users.last_name, users.is_email_verified FROM "users" LEFT JOIN teachers t ON users.user_id = t.user_id AND t.deleted_at IS NULL WHERE users.deleted_at IS NULL AND users.status = $1 AND users.role_id != $2 AND (NOT (users.role_id = $3 AND t.is_external = $4))`)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(query).
			WithArgs("active", "role_company", "role_teacher", true).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "email", "first_name", "last_name", "is_email_verified"}).
				AddRow("usr-1", "user1@test.com", "Lucia", "Gomez", true).
				AddRow("usr-2", "user2@test.com", "Marcos", "Lopez", false))

		users, err := repo.FindCampaignEligibleUsers(context.Background())
		require.NoError(t, err)
		require.Len(t, users, 2)
		assert.Equal(t, "usr-1", users[0].UserID)
		assert.Equal(t, "user1@test.com", users[0].Email)
		assert.Equal(t, "Lucia", users[0].FirstName)
		assert.True(t, users[0].IsEmailVerified)
		assert.False(t, users[1].IsEmailVerified)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(query).
			WithArgs("active", "role_company", "role_teacher", true).
			WillReturnError(errors.New("database connection lost"))

		users, err := repo.FindCampaignEligibleUsers(context.Background())
		assert.Error(t, err)
		assert.Nil(t, users)
	})
}
