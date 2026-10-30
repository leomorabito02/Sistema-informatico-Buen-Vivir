package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestDocIDs_Formatting tests composite document ID constructors.
func TestDocIDs_Formatting(t *testing.T) {
	t.Run("ActivitySkillDocID combines activity and skill IDs", func(t *testing.T) {
		// Arrange & Act
		id := ActivitySkillDocID("act-1", "skill-a")

		// Assert
		assert.Equal(t, "act-1_skill-a", id)
	})

	t.Run("ParticipantSkillDocID combines participant and skill IDs", func(t *testing.T) {
		// Arrange & Act
		id := ParticipantSkillDocID("part-1", "skill-a")

		// Assert
		assert.Equal(t, "part-1_skill-a", id)
	})

	t.Run("SubmissionDocID combines activity and participant IDs", func(t *testing.T) {
		// Arrange & Act
		id := SubmissionDocID("act-1", "part-1")

		// Assert
		assert.Equal(t, "act-1_part-1", id)
	})
}

// TestNewGCSFileStorage_EmptyBucketReturnsNil tests boundary where bucket is empty.
func TestNewGCSFileStorage_EmptyBucketReturnsNil(t *testing.T) {
	// Arrange & Act
	storage, client, err := NewGCSFileStorage(t.Context(), "", nil)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, storage)
	assert.Nil(t, client)
}
