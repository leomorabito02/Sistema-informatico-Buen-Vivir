package repositories

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
