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

func TestCompanyRepository_FindByID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewCompanyRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "companies" WHERE (company_id = $1 OR user_id = $2) AND "companies"."deleted_at" IS NULL ORDER BY "companies"."company_id" LIMIT $3`)).
			WithArgs("comp-1", "comp-1", 1).
			WillReturnRows(sqlmock.NewRows([]string{"company_id", "business_name"}).
				AddRow("comp-1", "Test Company"))

		comp, err := repo.FindByID(context.Background(), "comp-1")
		assert.NoError(t, err)
		assert.NotNil(t, comp)
		assert.Equal(t, "comp-1", comp.CompanyID)
	})

	t.Run("NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "companies" WHERE (company_id = $1 OR user_id = $2) AND "companies"."deleted_at" IS NULL ORDER BY "companies"."company_id" LIMIT $3`)).
			WithArgs("notfound", "notfound", 1).
			WillReturnRows(sqlmock.NewRows([]string{"company_id"}))

		comp, err := repo.FindByID(context.Background(), "notfound")
		assert.Error(t, err)
		assert.Nil(t, comp)
		assert.Contains(t, err.Error(), "empresa no encontrada")
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "companies" WHERE (company_id = $1 OR user_id = $2) AND "companies"."deleted_at" IS NULL ORDER BY "companies"."company_id" LIMIT $3`)).
			WithArgs("error", "error", 1).
			WillReturnError(errors.New("db error"))

		comp, err := repo.FindByID(context.Background(), "error")
		assert.Error(t, err)
		assert.Nil(t, comp)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestCompanyRepository_FindAll(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewCompanyRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "companies" WHERE "companies"."deleted_at" IS NULL`)).
			WillReturnRows(sqlmock.NewRows([]string{"company_id", "business_name"}).
				AddRow("comp-1", "Company 1").
				AddRow("comp-2", "Company 2"))

		comps, err := repo.FindAll(context.Background())
		assert.NoError(t, err)
		assert.Len(t, comps, 2)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "companies" WHERE "companies"."deleted_at" IS NULL`)).
			WillReturnError(errors.New("db error"))

		comps, err := repo.FindAll(context.Background())
		assert.Error(t, err)
		assert.Nil(t, comps)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestCompanyRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewCompanyRepository(db)

	t.Run("Success", func(t *testing.T) {
		comp := &models.Company{
			CompanyID:    "comp-1",
			BusinessName: "Test",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "companies"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Create(context.Background(), comp)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		comp := &models.Company{
			CompanyID:    "comp-1",
			BusinessName: "Test",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "companies"`)).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Create(context.Background(), comp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestCompanyRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewCompanyRepository(db)

	t.Run("Success", func(t *testing.T) {
		comp := &models.Company{
			CompanyID:    "comp-1",
			BusinessName: "Test",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "companies" SET`)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(context.Background(), comp)
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		comp := &models.Company{
			CompanyID:    "comp-1",
			BusinessName: "Test",
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "companies" SET`)).
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Update(context.Background(), comp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestCompanyRepository_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewCompanyRepository(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "companies" SET "deleted_at"=$1 WHERE company_id = $2 AND "companies"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "comp-1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(context.Background(), "comp-1")
		assert.NoError(t, err)
	})

	t.Run("DBError", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "companies" SET "deleted_at"=$1 WHERE company_id = $2 AND "companies"."deleted_at" IS NULL`)).
			WithArgs(sqlmock.AnyArg(), "error").
			WillReturnError(errors.New("db error"))
		mock.ExpectRollback()

		err := repo.Delete(context.Background(), "error")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})
}

func TestCompanyRepository_FindByID_AceptaIDDeUsuario(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewCompanyRepository(db)

	// Inserción Laboral consulta la empresa con el id de usuario.
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "companies" WHERE (company_id = $1 OR user_id = $2) AND "companies"."deleted_at" IS NULL ORDER BY "companies"."company_id" LIMIT $3`)).
		WithArgs("usr-emp", "usr-emp", 1).
		WillReturnRows(sqlmock.NewRows([]string{"company_id", "user_id", "business_name"}).
			AddRow("comp-1", "usr-emp", "Vivero"))

	comp, err := repo.FindByID(context.Background(), "usr-emp")
	assert.NoError(t, err)
	assert.Equal(t, "comp-1", comp.CompanyID)
	assert.Equal(t, "usr-emp", comp.UserID)
}
