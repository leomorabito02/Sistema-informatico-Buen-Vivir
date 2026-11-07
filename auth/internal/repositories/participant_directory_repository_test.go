package repositories_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"auth/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestParticipantDirectoryRepository(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repositories.NewParticipantDirectoryRepository(db)

	t.Run("FindProfileByParticipantID_Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT participants.participant_id, participants.user_id, users.first_name, users.last_name, participants.birthdate, participants.approval_status FROM "participants" JOIN users ON users.user_id = participants.user_id AND users.deleted_at IS NULL WHERE participants.deleted_at IS NULL AND participants.participant_id = $1 LIMIT $2`)).
			WithArgs("p-1", 1).
			WillReturnRows(sqlmock.NewRows([]string{"participant_id", "user_id", "first_name", "last_name", "birthdate", "approval_status"}).
				AddRow("p-1", "u-1", "Ana", "Gomez", time.Now(), "approved"))

		row, err := repo.FindProfileByParticipantID(context.Background(), "p-1")
		assert.NoError(t, err)
		assert.NotNil(t, row)
		assert.Equal(t, "p-1", row.ParticipantID)
	})

	t.Run("FindProfileByParticipantID_NotFound", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT participants.participant_id, participants.user_id, users.first_name, users.last_name, participants.birthdate, participants.approval_status FROM "participants" JOIN users ON users.user_id = participants.user_id AND users.deleted_at IS NULL WHERE participants.deleted_at IS NULL AND participants.participant_id = $1 LIMIT $2`)).
			WithArgs("p-none", 1).
			WillReturnRows(sqlmock.NewRows([]string{"participant_id"}))

		row, err := repo.FindProfileByParticipantID(context.Background(), "p-none")
		assert.ErrorIs(t, err, repositories.ErrParticipantProfileNotFound)
		assert.Nil(t, row)
	})

	t.Run("FindProfileByUserID_Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT participants.participant_id, participants.user_id, users.first_name, users.last_name, participants.birthdate, participants.approval_status FROM "participants" JOIN users ON users.user_id = participants.user_id AND users.deleted_at IS NULL WHERE participants.deleted_at IS NULL AND participants.user_id = $1 LIMIT $2`)).
			WithArgs("u-1", 1).
			WillReturnRows(sqlmock.NewRows([]string{"participant_id", "user_id", "first_name", "last_name", "birthdate", "approval_status"}).
				AddRow("p-1", "u-1", "Ana", "Gomez", time.Now(), "approved"))

		row, err := repo.FindProfileByUserID(context.Background(), "u-1")
		assert.NoError(t, err)
		assert.NotNil(t, row)
		assert.Equal(t, "u-1", row.UserID)
	})

	t.Run("SearchProfiles_Success", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT participants.participant_id, participants.user_id, users.first_name, users.last_name, participants.birthdate, participants.approval_status FROM "participants" JOIN users ON users.user_id = participants.user_id AND users.deleted_at IS NULL WHERE participants.deleted_at IS NULL AND (users.first_name ILIKE $1 ESCAPE '\' OR users.last_name ILIKE $2 ESCAPE '\' OR (users.first_name || ' ' || users.last_name) ILIKE $3 ESCAPE '\') ORDER BY users.last_name, users.first_name LIMIT $4`)).
			WithArgs("%ana%", "%ana%", "%ana%", 10).
			WillReturnRows(sqlmock.NewRows([]string{"participant_id", "first_name"}).AddRow("p-1", "Ana"))

		rows, err := repo.SearchProfiles(context.Background(), "ana", 10)
		assert.NoError(t, err)
		assert.Len(t, rows, 1)
	})

	t.Run("FindUserSummaries_EmptyAndSuccess", func(t *testing.T) {
		// Empty slice returns empty without query
		rows, err := repo.FindUserSummaries(context.Background(), nil)
		assert.NoError(t, err)
		assert.Empty(t, rows)

		// Non-empty
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT user_id, first_name, last_name FROM "users" WHERE user_id IN ($1) AND deleted_at IS NULL`)).
			WithArgs("u-1").
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "first_name", "last_name"}).AddRow("u-1", "Ana", "Gomez"))

		rows, err = repo.FindUserSummaries(context.Background(), []string{"u-1"})
		assert.NoError(t, err)
		assert.Len(t, rows, 1)
		assert.Equal(t, "Ana", rows[0].FirstName)
	})
}
