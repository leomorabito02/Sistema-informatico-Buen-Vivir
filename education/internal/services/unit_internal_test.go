package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"buenvivir/education/internal/models"

	"github.com/stretchr/testify/assert"
)

// TestErrors_Methods verifies error interface implementations and type assertions.
func TestErrors_Methods(t *testing.T) {
	t.Run("ForbiddenError Error() and type check", func(t *testing.T) {
		// Arrange
		err := NewForbiddenError("prohibido")

		// Act
		msg := err.Error()

		// Assert
		assert.Equal(t, "prohibido", msg)
		assert.True(t, IsForbiddenError(err))
		assert.False(t, IsForbiddenError(errors.New("other")))
	})

	t.Run("UnavailableError Error() and type check", func(t *testing.T) {
		// Arrange
		err := NewUnavailableError("no disponible")

		// Act
		msg := err.Error()

		// Assert
		assert.Equal(t, "no disponible", msg)
		assert.True(t, IsUnavailableError(err))
		assert.False(t, IsUnavailableError(errors.New("other")))
	})

	t.Run("ValidationError type check negative case", func(t *testing.T) {
		// Arrange & Act & Assert
		assert.False(t, IsValidationError(errors.New("other")))
	})
}

// TestSleepCtx_DecisionTable tests sleepCtx behavior with expired timer vs canceled context.
func TestSleepCtx_DecisionTable(t *testing.T) {
	t.Run("Timer expires normally returns nil", func(t *testing.T) {
		// Arrange
		ctx := context.Background()

		// Act
		err := sleepCtx(ctx, 1*time.Millisecond)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("Context canceled returns context.Canceled error", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Act
		err := sleepCtx(ctx, 1*time.Hour)

		// Assert
		assert.ErrorIs(t, err, context.Canceled)
	})
}

// TestSafeFileName_BoundaryValues tests boundary values and sanitization for submission file names.
func TestSafeFileName_BoundaryValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty string defaults to archivo", "", "archivo"},
		{"Single dot defaults to archivo", ".", "archivo"},
		{"Double dot defaults to archivo", "..", "archivo"},
		{"Backslash path extracts base name", `C:\Users\Student\homework.pdf`, "homework.pdf"},
		{"Unix path extracts base name", "/tmp/uploads/homework.pdf", "homework.pdf"},
		{"Unsafe characters replaced with underscore", "my:file*name?.pdf", "my_file_name_.pdf"},
		{"Long file name truncated to last 100 characters", strings.Repeat("a", 120) + ".pdf", strings.Repeat("a", 96) + ".pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			got := safeFileName(tt.input)

			// Assert
			assert.Equal(t, tt.expected, got)
		})
	}
}

// TestBuildProgress_EquivalenceClasses tests progress calculation equivalence classes.
func TestBuildProgress_EquivalenceClasses(t *testing.T) {
	t.Run("Zero total activities results in 0 percent", func(t *testing.T) {
		// Arrange & Act
		p := buildProgress("p1", 0, 0)

		// Assert
		assert.Equal(t, 0.0, p.PorcentajeAsistencia)
		assert.Equal(t, 0, p.ActividadesTotal)
		assert.Equal(t, 0, p.ActividadesAsistidas)
	})

	t.Run("Partial completion calculates accurate percentage", func(t *testing.T) {
		// Arrange & Act
		p := buildProgress("p1", 4, 2)

		// Assert
		assert.Equal(t, 50.0, p.PorcentajeAsistencia)
		assert.Equal(t, 4, p.ActividadesTotal)
		assert.Equal(t, 2, p.ActividadesAsistidas)
	})

	t.Run("Full completion calculates 100 percent", func(t *testing.T) {
		// Arrange & Act
		p := buildProgress("p1", 5, 5)

		// Assert
		assert.Equal(t, 100.0, p.PorcentajeAsistencia)
	})
}

// TestSubmissionActivityPrefix tests prefix generation format.
func TestSubmissionActivityPrefix(t *testing.T) {
	// Arrange & Act
	prefix := submissionActivityPrefix("act-123")

	// Assert
	assert.Equal(t, "entregas/act-123/", prefix)
}

// TestCleanupUploads_NoOpWhenAnyFalse verifies early return when any is false.
func TestCleanupUploads_NoOpWhenAnyFalse(t *testing.T) {
	// Arrange
	s := &submissionService{}

	// Act & Assert (should not panic even with nil storage)
	s.cleanupUploads(context.Background(), "prefix", false)
}

// TestFilterSuggestions_BoundariesAndDuplicates tests duplicates and max limit capping.
func TestFilterSuggestions_BoundariesAndDuplicates(t *testing.T) {
	// Arrange
	catalogByKey := map[string]models.Skill{
		"excel": {ID: "s-excel", Nombre: "Excel", Categoria: models.SkillCategoryTecnologia},
	}
	excluded := map[string]bool{"forbidden": true}

	var suggestions []models.SkillSuggestion
	for i := 1; i <= 8; i++ {
		suggestions = append(suggestions, models.SkillSuggestion{
			Nombre: fmt.Sprintf("Skill %d", i),
		})
	}
	// Duplicate inside response
	suggestions = append([]models.SkillSuggestion{{Nombre: "Skill 1"}}, suggestions...)
	// Excluded item
	suggestions = append(suggestions, models.SkillSuggestion{Nombre: "forbidden"})

	// Act
	filtered := filterSuggestions(suggestions, excluded, catalogByKey)

	// Assert: capped at maxSuggestions (5)
	assert.Len(t, filtered, 5)
}
