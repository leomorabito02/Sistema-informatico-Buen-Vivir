package repositories

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidationMessage_DecisionTable validates ValidationMessage using a Decision Table covering status codes and error types.
func TestValidationMessage_DecisionTable(t *testing.T) {
	decisionTable := []struct {
		name          string
		err           error
		expectedOk    bool
		expectedSubstr string
	}{
		{
			name:       "Not a ContentfulAPIError",
			err:        errors.New("generic error"),
			expectedOk: false,
		},
		{
			name: "ContentfulAPIError with 404 status (not 422)",
			err: &ContentfulAPIError{
				StatusCode: http.StatusNotFound,
				Body:       `{"message":"Not Found"}`,
			},
			expectedOk: false,
		},
		{
			name: "ContentfulAPIError with 500 status (not 422)",
			err: &ContentfulAPIError{
				StatusCode: http.StatusInternalServerError,
				Body:       `{"message":"Server Error"}`,
			},
			expectedOk: false,
		},
		{
			name: "Contentful 422 with 'required' error",
			err: &ContentfulAPIError{
				StatusCode: http.StatusUnprocessableEntity,
				Body: `{
					"details": {
						"errors": [
							{"name": "required", "path": ["fields", "title"]}
						]
					}
				}`,
			},
			expectedOk:    true,
			expectedSubstr: "faltan campos obligatorios: título",
		},
		{
			name: "Contentful 422 with 'notResolvable' error",
			err: &ContentfulAPIError{
				StatusCode: http.StatusUnprocessableEntity,
				Body: `{
					"details": {
						"errors": [
							{"name": "notResolvable", "path": ["fields", "trainingArea"]}
						]
					}
				}`,
			},
			expectedOk:    true,
			expectedSubstr: "hay referencias borradas o sin publicar en: área formativa",
		},
		{
			name: "Contentful 422 with 'unique' error",
			err: &ContentfulAPIError{
				StatusCode: http.StatusUnprocessableEntity,
				Body: `{
					"details": {
						"errors": [
							{"name": "unique", "path": ["fields", "name"]}
						]
					}
				}`,
			},
			expectedOk:    true,
			expectedSubstr: "ya existe otro contenido con el mismo valor en: nombre",
		},
		{
			name: "Contentful 422 with unknown error name ('other')",
			err: &ContentfulAPIError{
				StatusCode: http.StatusUnprocessableEntity,
				Body: `{
					"details": {
						"errors": [
							{"name": "customValidation", "path": ["fields", "duration"]}
						]
					}
				}`,
			},
			expectedOk:    true,
			expectedSubstr: "hay valores inválidos en: duración",
		},
		{
			name: "Contentful 422 with combined multiple errors and duplicates",
			err: &ContentfulAPIError{
				StatusCode: http.StatusUnprocessableEntity,
				Body: `{
					"details": {
						"errors": [
							{"name": "required", "path": ["fields", "title"]},
							{"name": "required", "path": ["fields", "title"]},
							{"name": "unique", "path": ["fields", "url"]}
						]
					}
				}`,
			},
			expectedOk:    true,
			expectedSubstr: "faltan campos obligatorios: título; ya existe otro contenido con el mismo valor en: URL",
		},
		{
			name: "Contentful 422 with empty errors array",
			err: &ContentfulAPIError{
				StatusCode: http.StatusUnprocessableEntity,
				Body:       `{"details": {"errors": []}}`,
			},
			expectedOk:    true,
			expectedSubstr: "Contentful rechazó el contenido por validación",
		},
		{
			name: "Contentful 422 with malformed short path (<2 elements)",
			err: &ContentfulAPIError{
				StatusCode: http.StatusUnprocessableEntity,
				Body: `{
					"details": {
						"errors": [
							{"name": "required", "path": ["fields"]}
						]
					}
				}`,
			},
			expectedOk:    true,
			expectedSubstr: "faltan campos obligatorios: ?",
		},
	}

	for _, tt := range decisionTable {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			testErr := tt.err

			// Act
			msg, ok := ValidationMessage(testErr)

			// Assert
			assert.Equal(t, tt.expectedOk, ok)
			if tt.expectedOk {
				assert.Contains(t, msg, tt.expectedSubstr)
			}
		})
	}
}

// TestIsNotFound_EquivalenceClasses tests sentinel and Contentful 404 detection.
func TestIsNotFound_EquivalenceClasses(t *testing.T) {
	t.Run("Sentinel ErrNotFound is recognized", func(t *testing.T) {
		// Arrange & Act & Assert
		assert.True(t, IsNotFound(ErrNotFound))
	})

	t.Run("ContentfulAPIError with 404 is recognized", func(t *testing.T) {
		// Arrange
		err := &ContentfulAPIError{StatusCode: http.StatusNotFound}

		// Act & Assert
		assert.True(t, IsNotFound(err))
	})

	t.Run("ContentfulAPIError with non-404 is not treated as not found", func(t *testing.T) {
		// Arrange
		err := &ContentfulAPIError{StatusCode: http.StatusBadRequest}

		// Act & Assert
		assert.False(t, IsNotFound(err))
	})

	t.Run("Other error is not treated as not found", func(t *testing.T) {
		// Arrange & Act & Assert
		assert.False(t, IsNotFound(errors.New("other")))
		assert.False(t, IsNotFound(nil))
	})
}
