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

func TestRoleRepository_FindByID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewRoleRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "roles" WHERE role_id = $1 AND "roles"."deleted_at" IS NULL ORDER BY "roles"."role_id" LIMIT $2`)).
			WithArgs("role-1", 1).
			WillReturnRows(sqlmock.NewRows([]string{"role_id", "name"}).
				AddRow("role-1", "Admin"))

		role, err := repo.FindByID(context.Background(), "role-1")
		assert.NoError(t, err)
		assert.NotNil(t, role)
		assert.Equal(t, "role-1", role.RoleID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "roles" WHERE role_id = $1 AND "roles"."deleted_at" IS NULL ORDER BY "roles"."role_id" LIMIT $2`)).
			WithArgs("notfound", 1).
			WillReturnRows(sqlmock.NewRows([]string{"role_id"}))

		role, err := repo.FindByID(context.Background(), "notfound")
		assert.Error(t, err)
		assert.Nil(t, role)
		assert.Contains(t, err.Error(), "rol no encontrado")
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "roles" WHERE role_id = $1 AND "roles"."deleted_at" IS NULL ORDER BY "roles"."role_id" LIMIT $2`)).
			WithArgs("error", 1).
			WillReturnError(errors.New("db error"))

		role, err := repo.FindByID(context.Background(), "error")
		assert.Error(t, err)
		assert.Nil(t, role)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestRoleRepository_FindAll(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewRoleRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "roles" WHERE "roles"."deleted_at" IS NULL`)).
			WillReturnRows(sqlmock.NewRows([]string{"role_id", "name"}).
				AddRow("role-1", "Admin").
				AddRow("role-2", "User"))

		roles, err := repo.FindAll(context.Background())
		assert.NoError(t, err)
		assert.Len(t, roles, 2)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "roles" WHERE "roles"."deleted_at" IS NULL`)).
			WillReturnError(errors.New("db error"))

		roles, err := repo.FindAll(context.Background())
		assert.Error(t, err)
		assert.Nil(t, roles)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestRoleRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewRoleRepository(db)

	t.Run("Success", func(t *testing.T) {
		role := &models.Role{
			RoleID: "role-1",
			Name:   "Admin",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "roles"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Create(context.Background(), role)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		role := &models.Role{
			RoleID: "role-1",
			Name:   "Admin",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "roles"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Create(context.Background(), role)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestRoleRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewRoleRepository(db)

	t.Run("Success", func(t *testing.T) {
		role := &models.Role{
			RoleID: "role-1",
			Name:   "Admin",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "roles" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(context.Background(), role)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		role := &models.Role{
			RoleID: "role-1",
			Name:   "Admin",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "roles" SET`)).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Update(context.Background(), role)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestRoleRepository_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewRoleRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "roles" SET "deleted_at"=$1 WHERE role_id = $2 AND "roles"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "role-1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), "role-1")
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "roles" SET "deleted_at"=$1 WHERE role_id = $2 AND "roles"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "error").
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Delete(context.Background(), "error")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}
