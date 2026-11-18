package events_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/events"
)

// TestSameTypes_DecisionTable tests slicing equality across length, ordering, and content permutations.
func TestSameTypes_DecisionTable(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{"Both empty", []string{}, []string{}, true},
		{"Different lengths", []string{"a"}, []string{"a", "b"}, false},
		{"Same elements same order", []string{"a", "b"}, []string{"a", "b"}, true},
		{"Same elements different order", []string{"b", "a"}, []string{"a", "b"}, true},
		{"Same length different elements", []string{"a", "b"}, []string{"a", "c"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			res := events.SameTypes(tt.a, tt.b)

			// Assert
			assert.Equal(t, tt.expected, res)
		})
	}
}

// TestInfraFilters_EquivalenceClasses tests file parsing error handling and regex extraction.
func TestInfraFilters_EquivalenceClasses(t *testing.T) {
	t.Run("Class 1: Non-existent file returns error", func(t *testing.T) {
		// Act
		filters, err := events.InfraFilters("non-existent-file.sh")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, filters)
	})

	t.Run("Class 2: File without subscriptions returns error", func(t *testing.T) {
		// Arrange
		tmp := filepath.Join(t.TempDir(), "setup.sh")
		require.NoError(t, os.WriteFile(tmp, []byte("#!/bin/bash\necho hello"), 0o600))

		// Act
		filters, err := events.InfraFilters(tmp)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no se encontraron suscripciones")
		assert.Nil(t, filters)
	})

	t.Run("Class 3: File with matching push subscription returns parsed filters", func(t *testing.T) {
		// Arrange
		tmp := filepath.Join(t.TempDir(), "setup.sh")
		content := "ensure_push_sub educacion-sub endpoint \\\n  \"$(filter evento.b evento.a)\"\n"
		require.NoError(t, os.WriteFile(tmp, []byte(content), 0o600))

		// Act
		filters, err := events.InfraFilters(tmp)

		// Assert
		require.NoError(t, err)
		require.Len(t, filters, 1)
		assert.Equal(t, []string{"evento.a", "evento.b"}, filters["educacion-sub"])
	})
}
