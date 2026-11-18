package repositories

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"buenvivir/education/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContentfulAPIError_Error(t *testing.T) {
	err := &ContentfulAPIError{StatusCode: http.StatusUnprocessableEntity, Body: `{"message":"invalid"}`}
	assert.Equal(t, `contentful respondió 422: {"message":"invalid"}`, err.Error())
}

func TestIsConflict(t *testing.T) {
	assert.True(t, IsConflict(&ContentfulAPIError{StatusCode: http.StatusConflict}))
	assert.False(t, IsConflict(&ContentfulAPIError{StatusCode: http.StatusNotFound}))
	assert.False(t, IsConflict(errors.New("otro error")))
	assert.False(t, IsConflict(nil))
}

func TestIsNotFound_RecognisesBothSources(t *testing.T) {
	assert.True(t, IsNotFound(ErrNotFound))
	assert.True(t, IsNotFound(&ContentfulAPIError{StatusCode: http.StatusNotFound}))
	assert.False(t, IsNotFound(&ContentfulAPIError{StatusCode: http.StatusConflict}))
	assert.False(t, IsNotFound(errors.New("cualquier cosa")))
}

func TestNewContentfulClient_DefaultsAndLocaleFallback(t *testing.T) {
	c := NewContentfulClient(&config.Config{
		ContentfulSpaceID:     "space-1",
		ContentfulEnvironment: config.ContentfulEnvTesting,
		ContentfulCDAToken:    "cda",
		ContentfulCMAToken:    "cma",
		ContentfulLocale:      "es-AR",
	})

	assert.Equal(t, "es-AR", c.Locale())
	assert.Equal(t, defaultCDABaseURL, c.cdaBaseURL)
	assert.Equal(t, defaultCMABaseURL, c.cmaBaseURL)
	assert.Equal(t, defaultUploadBaseURL, c.uploadBaseURL)
	assert.NotNil(t, c.httpClient)
}

func TestContentfulClient_DeleteEntryCascade_DecisionTable(t *testing.T) {
	t.Run("Published entry: unpublish 200 then delete 200 succeeds", func(t *testing.T) {
		// Arrange
		var unpublished, deleted bool
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/published") {
				unpublished = true
				w.WriteHeader(http.StatusOK)
				return
			}
			if r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/entries/e1") {
				deleted = true
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := newTestClient("", server.URL)

		// Act
		err := client.DeleteEntryCascade(t.Context(), "e1")

		// Assert
		require.NoError(t, err)
		assert.True(t, unpublished)
		assert.True(t, deleted)
	})

	t.Run("Draft entry: unpublish 400 with PublishedVersion==0 proceeds to delete", func(t *testing.T) {
		// Arrange
		var checkedManagement, deleted bool
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/published") {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"message":"Not published"}`))
				return
			}
			if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/entries/e2") {
				checkedManagement = true
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"sys": map[string]any{"id": "e2", "version": 1, "publishedVersion": 0},
				})
				return
			}
			if r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/entries/e2") {
				deleted = true
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := newTestClient("", server.URL)

		// Act
		err := client.DeleteEntryCascade(t.Context(), "e2")

		// Assert
		require.NoError(t, err)
		assert.True(t, checkedManagement)
		assert.True(t, deleted)
	})

	t.Run("Unpublish 400 but PublishedVersion!=0 returns error", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/published") {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"message":"Cannot unpublish"}`))
				return
			}
			if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/entries/e3") {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"sys": map[string]any{"id": "e3", "version": 3, "publishedVersion": 2},
				})
				return
			}
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := newTestClient("", server.URL)

		// Act
		err := client.DeleteEntryCascade(t.Context(), "e3")

		// Assert
		assert.Error(t, err)
	})

	t.Run("Unpublish 500 fails fast", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client := newTestClient("", server.URL)

		// Act
		err := client.DeleteEntryCascade(t.Context(), "e4")

		// Assert
		assert.Error(t, err)
	})
}

func TestContentfulClient_CreateUpdatePublishEntry(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/entries"):
			assert.Equal(t, "course", r.Header.Get("X-Contentful-Content-Type"))
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "c1", "version": 1}})
		case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/entries/c1"):
			assert.Equal(t, "1", r.Header.Get("X-Contentful-Version"))
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "c1", "version": 2}})
		case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/entries/c1/published"):
			assert.Equal(t, "2", r.Header.Get("X-Contentful-Version"))
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "c1", "version": 3}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := newTestClient("", server.URL)

	// Act & Assert
	entry, err := client.CreateEntry(t.Context(), "course", map[string]any{"title": "Curso"})
	require.NoError(t, err)
	assert.Equal(t, "c1", entry.Sys.ID)

	updated, err := client.UpdateEntry(t.Context(), "c1", 1, map[string]any{"title": "Curso Modificado"})
	require.NoError(t, err)
	assert.Equal(t, 2, updated.Sys.Version)

	published, err := client.PublishEntry(t.Context(), "c1", 2)
	require.NoError(t, err)
	assert.Equal(t, 3, published.Sys.Version)
}

// TestContentfulClient_MarshalAndUnmarshalErrors tests boundary conditions when serializing or deserializing Contentful entries.
func TestContentfulClient_MarshalAndUnmarshalErrors(t *testing.T) {
	t.Run("CreateEntry with unmarshalable fields returns error", func(t *testing.T) {
		// Arrange
		client := newTestClient("", "")
		badFields := map[string]any{"bad": make(chan int)}

		// Act
		entry, err := client.CreateEntry(t.Context(), "course", badFields)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, entry)
	})

	t.Run("UpdateEntry with unmarshalable fields returns error", func(t *testing.T) {
		// Arrange
		client := newTestClient("", "")
		badFields := map[string]any{"bad": make(chan int)}

		// Act
		entry, err := client.UpdateEntry(t.Context(), "c1", 1, badFields)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, entry)
	})

	t.Run("GetManagementEntry with malformed response JSON returns error", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{malformed-json`))
		}))
		defer server.Close()

		client := newTestClient("", server.URL)

		// Act
		entry, err := client.GetManagementEntry(t.Context(), "c1")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, entry)
	})

	t.Run("CreateEntry with malformed response JSON returns error", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{malformed-json`))
		}))
		defer server.Close()

		client := newTestClient("", server.URL)

		// Act
		entry, err := client.CreateEntry(t.Context(), "course", map[string]any{"title": "Valid"})

		// Assert
		assert.Error(t, err)
		assert.Nil(t, entry)
	})

	t.Run("UpdateEntry with malformed response JSON returns error", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{malformed-json`))
		}))
		defer server.Close()

		client := newTestClient("", server.URL)

		// Act
		entry, err := client.UpdateEntry(t.Context(), "c1", 1, map[string]any{"title": "Valid"})

		// Assert
		assert.Error(t, err)
		assert.Nil(t, entry)
	})

	t.Run("PublishEntry with malformed response JSON returns error", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{malformed-json`))
		}))
		defer server.Close()

		client := newTestClient("", server.URL)

		// Act
		entry, err := client.PublishEntry(t.Context(), "c1", 1)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, entry)
	})

	t.Run("DeleteEntryCascade fails when GetManagementEntry returns error on draft check", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/published") {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"message":"Not published"}`))
				return
			}
			if r.Method == http.MethodGet {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		client := newTestClient("", server.URL)

		// Act
		err := client.DeleteEntryCascade(t.Context(), "e-draft")

		// Assert
		assert.Error(t, err)
	})
}

// TestContentfulClient_ListAndNetworkErrors tests malformed JSON and network error branches.
func TestContentfulClient_ListAndNetworkErrors(t *testing.T) {
	t.Run("ListEntriesWhere returns error on malformed JSON", func(t *testing.T) {
		// Arrange
		cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{malformed-cda`))
		}))
		defer cda.Close()

		client := newTestClient(cda.URL, "")

		// Act
		coll, err := client.ListEntriesWhere(t.Context(), "course", 1, nil)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, coll)
	})

	t.Run("ListManagementEntriesWhere returns error on malformed JSON", func(t *testing.T) {
		// Arrange
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{malformed-cma`))
		}))
		defer cma.Close()

		client := newTestClient("", cma.URL)

		// Act
		coll, err := client.ListManagementEntriesWhere(t.Context(), "course", nil)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, coll)
	})

	t.Run("doCDA invalid URL returns error", func(t *testing.T) {
		// Arrange
		client := newTestClient("://invalid-url", "")

		// Act
		_, _, err := client.doCDA(t.Context(), http.MethodGet, "/entries", nil)

		// Assert
		assert.Error(t, err)
	})

	t.Run("doCMA invalid URL returns error", func(t *testing.T) {
		// Arrange
		client := newTestClient("", "://invalid-url")

		// Act
		_, _, err := client.doCMA(t.Context(), http.MethodGet, "/entries", nil, nil)

		// Assert
		assert.Error(t, err)
	})

	t.Run("do network error wraps and returns error", func(t *testing.T) {
		// Arrange
		client := newTestClient("", "http://127.0.0.1:1")

		// Act
		_, err := client.GetManagementEntry(t.Context(), "c1")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error de red hablando con Contentful")
	})
}
