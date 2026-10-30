package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"buenvivir/education/internal/repositories"
	"buenvivir/education/internal/services"
)

// TestWriteError_DecisionTable tests all branches of writeError using a Decision Table.
func TestWriteError_DecisionTable(t *testing.T) {
	decisionMatrix := []struct {
		name         string
		err          error
		expectedCode int
		expectedBody string
	}{
		{
			name:         "Validation error maps to 400 Bad Request",
			err:          services.NewValidationError("campo obligatorio faltante"),
			expectedCode: http.StatusBadRequest,
			expectedBody: "campo obligatorio faltante",
		},
		{
			name:         "Forbidden error maps to 403 Forbidden",
			err:          services.NewForbiddenError("no tienes permisos"),
			expectedCode: http.StatusForbidden,
			expectedBody: "no tienes permisos",
		},
		{
			name:         "Unavailable error maps to 503 Service Unavailable",
			err:          services.NewUnavailableError("ia no disponible"),
			expectedCode: http.StatusServiceUnavailable,
			expectedBody: "ia no disponible",
		},
		{
			name:         "Not found error maps to 404 Not Found",
			err:          repositories.ErrNotFound,
			expectedCode: http.StatusNotFound,
			expectedBody: "recurso no encontrado",
		},
		{
			name:         "Conflict error maps to 409 Conflict",
			err:          &repositories.ContentfulAPIError{StatusCode: http.StatusConflict},
			expectedCode: http.StatusConflict,
			expectedBody: "el recurso fue modificado por otra edición",
		},
		{
			name: "ValidationMessage error maps to 422 Unprocessable Entity",
			err: &repositories.ContentfulAPIError{
				StatusCode: http.StatusUnprocessableEntity,
				Body:       `{"details":{"errors":[{"name":"required","path":["fields","title"]}]}}`,
			},
			expectedCode: http.StatusUnprocessableEntity,
			expectedBody: "faltan campos obligatorios: título",
		},
		{
			name:         "Default unknown error maps to 502 Bad Gateway",
			err:          errors.New("connection reset by peer"),
			expectedCode: http.StatusBadGateway,
			expectedBody: "error de infraestructura: connection reset by peer",
		},
	}

	for _, tt := range decisionMatrix {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			rec := httptest.NewRecorder()

			// Act
			writeError(rec, tt.err)

			// Assert
			assert.Equal(t, tt.expectedCode, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.expectedBody)
		})
	}
}

// TestHTTPHelpers_JSONEncodingAndDecoding tests JSON writing and decoding with boundaries.
func TestHTTPHelpers_JSONEncodingAndDecoding(t *testing.T) {
	t.Run("writeJSON writes valid JSON and headers", func(t *testing.T) {
		// Arrange
		rec := httptest.NewRecorder()
		payload := map[string]string{"key": "value"}

		// Act
		writeJSON(rec, http.StatusCreated, payload)

		// Assert
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
		assert.Contains(t, rec.Body.String(), `"key":"value"`)
	})

	t.Run("writeJSON with nil payload produces empty body", func(t *testing.T) {
		// Arrange
		rec := httptest.NewRecorder()

		// Act
		writeJSON(rec, http.StatusNoContent, nil)

		// Assert
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, rec.Body.String())
	})

	t.Run("decodeJSONBody decodes valid payload and handles boundary", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
		var dst struct {
			Name string `json:"name"`
		}

		// Act
		err := decodeJSONBody(req, &dst)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "test", dst.Name)
	})

	t.Run("decodeJSONBody returns error on malformed JSON boundary", func(t *testing.T) {
		// Arrange
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{invalid-json`))
		var dst map[string]any

		// Act
		err := decodeJSONBody(req, &dst)

		// Assert
		assert.Error(t, err)
	})
}

// TestUploadContentType_DecisionTable tests MIME type resolution based on declared type and extension.
func TestUploadContentType_DecisionTable(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		declared string
		expected string
	}{
		{"Explicit valid type preserved", "doc.pdf", "application/pdf", "application/pdf"},
		{"Explicit type with params stripped", "doc.pdf", "application/pdf; charset=utf-8", "application/pdf"},
		{"Octet-stream falls back to pdf extension", "doc.pdf", "application/octet-stream", "application/pdf"},
		{"Octet-stream falls back to docx extension", "doc.docx", "application/octet-stream", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"Empty declared falls back to png extension", "photo.png", "", "image/png"},
		{"Unknown extension with octet-stream returns empty", "file.xyz", "application/octet-stream", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			got := uploadContentType(tt.fileName, tt.declared)

			// Assert
			assert.Equal(t, tt.expected, got)
		})
	}
}
