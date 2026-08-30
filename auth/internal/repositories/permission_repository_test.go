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

func TestPermissionRepository_FindByID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewPermissionRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "permissions" WHERE permission_id = $1 AND "permissions"."deleted_at" IS NULL ORDER BY "permissions"."permission_id" LIMIT $2`)).
			WithArgs("perm-1", 1).
			WillReturnRows(sqlmock.NewRows([]string{"permission_id", "action"}).
				AddRow("perm-1", "create"))

		perm, err := repo.FindByID(context.Background(), "perm-1")
		assert.NoError(t, err)
		assert.NotNil(t, perm)
		assert.Equal(t, "perm-1", perm.PermissionID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "permissions" WHERE permission_id = $1 AND "permissions"."deleted_at" IS NULL ORDER BY "permissions"."permission_id" LIMIT $2`)).
			WithArgs("notfound", 1).
			WillReturnRows(sqlmock.NewRows([]string{"permission_id"}))

		perm, err := repo.FindByID(context.Background(), "notfound")
		assert.Error(t, err)
		assert.Nil(t, perm)
		assert.Contains(t, err.Error(), "permiso no encontrado")
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "permissions" WHERE permission_id = $1 AND "permissions"."deleted_at" IS NULL ORDER BY "permissions"."permission_id" LIMIT $2`)).
			WithArgs("error", 1).
			WillReturnError(errors.New("db error"))

		perm, err := repo.FindByID(context.Background(), "error")
		assert.Error(t, err)
		assert.Nil(t, perm)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestPermissionRepository_FindAll(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewPermissionRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "permissions" WHERE "permissions"."deleted_at" IS NULL`)).
			WillReturnRows(sqlmock.NewRows([]string{"permission_id", "action"}).
				AddRow("perm-1", "create").
				AddRow("perm-2", "read"))

		perms, err := repo.FindAll(context.Background())
		assert.NoError(t, err)
		assert.Len(t, perms, 2)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "permissions" WHERE "permissions"."deleted_at" IS NULL`)).
			WillReturnError(errors.New("db error"))

		perms, err := repo.FindAll(context.Background())
		assert.Error(t, err)
		assert.Nil(t, perms)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestPermissionRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewPermissionRepository(db)

	t.Run("Success", func(t *testing.T) {
		perm := &models.Permission{
			PermissionID: "perm-1",
			Action:       "create",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "permissions"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Create(context.Background(), perm)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		perm := &models.Permission{
			PermissionID: "perm-1",
			Action:       "create",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "permissions"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Create(context.Background(), perm)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestPermissionRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewPermissionRepository(db)

	t.Run("Success", func(t *testing.T) {
		perm := &models.Permission{
			PermissionID: "perm-1",
			Action:       "create",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "permissions" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(context.Background(), perm)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		perm := &models.Permission{
			PermissionID: "perm-1",
			Action:       "create",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "permissions" SET`)).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Update(context.Background(), perm)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestPermissionRepository_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewPermissionRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "permissions" SET "deleted_at"=$1 WHERE permission_id = $2 AND "permissions"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "perm-1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), "perm-1")
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "permissions" SET "deleted_at"=$1 WHERE permission_id = $2 AND "permissions"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "error").
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Delete(context.Background(), "error")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestPermissionRepository_CheckRolePermission(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewPermissionRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "role_permissions" JOIN permissions ON permissions.permission_id = role_permissions.permission_id WHERE role_permissions.role_id = $1 AND permissions.action = $2 AND permissions.resource = $3`)).
			WithArgs("role-1", "read", "user").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		var count int64
		err := repo.CheckRolePermission(context.Background(), "role-1", "read", "user", &count)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "role_permissions" JOIN permissions ON permissions.permission_id = role_permissions.permission_id WHERE role_permissions.role_id = $1 AND permissions.action = $2 AND permissions.resource = $3`)).
			WithArgs("role-1", "read", "user").
			WillReturnError(errors.New("db error"))

		var count int64
		err := repo.CheckRolePermission(context.Background(), "role-1", "read", "user", &count)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}
