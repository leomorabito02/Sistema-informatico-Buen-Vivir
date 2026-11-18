package repositories

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"buenvivir/education/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaterialRepository_List_ResolvesFile(t *testing.T) {
	cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "material", r.URL.Query().Get("content_type"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"sys":    map[string]any{"id": "m1"},
					"fields": map[string]any{"title": "Guía", "type": "PDF", "file": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Asset", "id": "asset-1"}}},
				},
			},
			"includes": map[string]any{
				"Asset": []map[string]any{
					{
						"sys": map[string]any{"id": "asset-1"},
						"fields": map[string]any{
							"title": "Guía PDF",
							"file": map[string]any{
								"url":         "//images.ctfassets.net/x.pdf",
								"fileName":    "x.pdf",
								"contentType": "application/pdf",
							},
						},
					},
				},
			},
		})
	}))
	defer cda.Close()

	repo := NewMaterialRepository(newTestClient(cda.URL, ""))

	materials, err := repo.List(t.Context())
	require.NoError(t, err)
	require.Len(t, materials, 1)
	require.NotNil(t, materials[0].File)
	assert.Equal(t, "https://images.ctfassets.net/x.pdf", materials[0].File.URL)
}

func TestMaterialRepository_Create_AutoPublishes(t *testing.T) {
	var methodsCalled []string
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodsCalled = append(methodsCalled, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "m-new", "version": 1}})
	}))
	defer cma.Close()

	repo := NewMaterialRepository(newTestClient("", cma.URL))

	material, err := repo.Create(t.Context(), models.MaterialInput{Title: "Guía", Type: "PDF", FileAssetID: "asset-1"})
	require.NoError(t, err)
	assert.Equal(t, "m-new", material.ID)

	require.Len(t, methodsCalled, 2)
	assert.Contains(t, methodsCalled[1], "/published")
}

func TestMaterialRepository_Delete(t *testing.T) {
	var methodsCalled []string
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodsCalled = append(methodsCalled, r.Method)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "m1"}})
	}))
	defer cma.Close()

	repo := NewMaterialRepository(newTestClient("", cma.URL))

	err := repo.Delete(t.Context(), "m1")
	require.NoError(t, err)
	require.Len(t, methodsCalled, 2) // unpublish + delete
}

func TestMaterialRepository_ErrorHandling(t *testing.T) {
	t.Run("List error propagates", func(t *testing.T) {
		cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cda.Close()

		repo := NewMaterialRepository(newTestClient(cda.URL, ""))
		materials, err := repo.List(t.Context())
		assert.Error(t, err)
		assert.Nil(t, materials)
	})

	t.Run("Create error propagates when CreateEntry fails", func(t *testing.T) {
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cma.Close()

		repo := NewMaterialRepository(newTestClient("", cma.URL))
		material, err := repo.Create(t.Context(), models.MaterialInput{Title: "Guía", Type: "Video", URL: "https://example.com"})
		assert.Error(t, err)
		assert.Nil(t, material)
	})

	t.Run("Create error propagates when PublishEntry fails", func(t *testing.T) {
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "m1", "version": 1}})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cma.Close()

		repo := NewMaterialRepository(newTestClient("", cma.URL))
		material, err := repo.Create(t.Context(), models.MaterialInput{Title: "Guía", Type: "Video", URL: "https://example.com"})
		assert.Error(t, err)
		assert.Nil(t, material)
	})

	t.Run("Update error propagates when UpdateEntry fails", func(t *testing.T) {
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "m1", "version": 1}})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cma.Close()

		repo := NewMaterialRepository(newTestClient("", cma.URL))
		material, err := repo.Update(t.Context(), "m1", models.MaterialInput{Title: "Guía", Type: "PDF", FileAssetID: "asset-1", URL: "https://example.com"})
		assert.Error(t, err)
		assert.Nil(t, material)
	})

	t.Run("Update error propagates when PublishEntry fails", func(t *testing.T) {
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet:
				_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "m1", "version": 1}})
			case r.Method == http.MethodPut && !strings.Contains(r.URL.Path, "/published"):
				_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "m1", "version": 2}})
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
		}))
		defer cma.Close()

		repo := NewMaterialRepository(newTestClient("", cma.URL))
		material, err := repo.Update(t.Context(), "m1", models.MaterialInput{Title: "Guía", Type: "PDF", FileAssetID: "asset-1", URL: "https://example.com"})
		assert.Error(t, err)
		assert.Nil(t, material)
	})
}

func TestMaterialRepository_Delete_Error(t *testing.T) {
	// Arrange
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer cma.Close()

	repo := NewMaterialRepository(newTestClient("", cma.URL))

	// Act
	err := repo.Delete(t.Context(), "m1")

	// Assert
	assert.Error(t, err)
}

func TestMaterialRepository_UploadAndCreate_Errors(t *testing.T) {
	t.Run("CreateEntry failure returns error", func(t *testing.T) {
		// Arrange
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "/assets") {
				if r.Method == http.MethodPost {
					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "asset-1", "version": 1}})
					return
				}
				if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/processing") {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "asset-1", "version": 2}})
					return
				}
				if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/published") {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "asset-1", "version": 3}})
					return
				}
				if r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(map[string]any{
						"sys": map[string]any{"id": "asset-1", "version": 2},
						"fields": map[string]any{
							"file": map[string]any{
								"en-US": map[string]any{"url": "//images.ctfassets.net/test.pdf"},
							},
						},
					})
					return
				}
			}
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cma.Close()

		repo := NewMaterialRepository(newTestClient("", cma.URL))

		// Act
		mat, err := repo.UploadAndCreate(t.Context(), "Title", "PDF", "test.pdf", "application/pdf", []byte("%PDF"))

		// Assert
		assert.Error(t, err)
		assert.Nil(t, mat)
	})
}
