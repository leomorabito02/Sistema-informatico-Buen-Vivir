package repositories_test

import (
	"context"
	"regexp"
	"testing"

	"auth/internal/models"
	"auth/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestParticipantDocumentRepository_CRUD(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewParticipantDocumentRepository(db)

	doc := &models.ParticipantDocument{
		ID:            "pdoc-1",
		ParticipantID: "part-1",
		Name:          "DNI",
		DocURL:        "https://storage/documents/part-1/dni.pdf",
		FileType:      "application/pdf",
	}

	// Create
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO "participant_documents"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(context.Background(), doc)
	assert.NoError(t, err)

	// GetByID
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "participant_documents" WHERE id = $1`)).
		WithArgs("pdoc-1", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "participant_id", "name"}).AddRow("pdoc-1", "part-1", "DNI"))

	got, err := repo.GetByID(context.Background(), "pdoc-1")
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, "pdoc-1", got.ID)

	// GetByParticipantID
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "participant_documents" WHERE participant_id = $1`)).
		WithArgs("part-1").
		WillReturnRows(sqlmock.NewRows([]string{"id", "participant_id"}).AddRow("pdoc-1", "part-1"))

	docs, err := repo.GetByParticipantID(context.Background(), "part-1")
	assert.NoError(t, err)
	assert.Len(t, docs, 1)

	// Update
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "participant_documents"`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.Update(context.Background(), doc)
	assert.NoError(t, err)

	// Delete (soft delete)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "participant_documents" SET "deleted_at"=`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = repo.Delete(context.Background(), "pdoc-1")
	assert.NoError(t, err)
}
