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

func TestMaterialRepository_Update_UsesCurrentVersionAndRepublishes(t *testing.T) {
	var calls []string
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "m1", "version": 7}})
		case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/published"):
			assert.Equal(t, "8", r.Header.Get("X-Contentful-Version"))
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "m1", "version": 9}})
		case r.Method == http.MethodPut:
			assert.Equal(t, "7", r.Header.Get("X-Contentful-Version"))
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "m1", "version": 8}})
		}
	}))
	defer cma.Close()

	repo := NewMaterialRepository(newTestClient("", cma.URL))

	material, err := repo.Update(t.Context(), "m1", models.MaterialInput{
		Title: "Guía v2", Type: "PDF", FileAssetID: "asset-9", URL: "",
	})
	require.NoError(t, err)
	assert.Equal(t, "m1", material.ID)
	assert.Equal(t, "Guía v2", material.Title)
	require.NotNil(t, material.File)
	assert.Equal(t, "asset-9", material.File.ID)

	require.Len(t, calls, 3)
	assert.Equal(t, http.MethodGet+" /spaces/space-1/environments/testing/entries/m1", calls[0])
	assert.Contains(t, calls[2], "/published")
}

func TestMaterialRepository_Update_PropagatesGetError(t *testing.T) {
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer cma.Close()

	repo := NewMaterialRepository(newTestClient("", cma.URL))

	material, err := repo.Update(t.Context(), "no-existe", models.MaterialInput{Title: "x"})
	assert.Nil(t, material)
	require.Error(t, err)
	assert.True(t, IsNotFound(err))
}

func TestMaterialRepository_UploadAndCreate_FullAssetFlow(t *testing.T) {
	upload := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/spaces/space-1/uploads", r.URL.Path)
		assert.Equal(t, "Bearer cma-token", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "upload-1"}})
	}))
	defer upload.Close()

	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		p := r.URL.Path
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(p, "/assets"):
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "asset-1", "version": 1}})
		case r.Method == http.MethodPut && strings.Contains(p, "/assets/asset-1/files/"):
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "asset-1", "version": 2}})
		case r.Method == http.MethodGet && strings.HasSuffix(p, "/assets/asset-1"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"sys":    map[string]any{"version": 2},
				"fields": map[string]any{"file": map[string]any{"en-US": map[string]any{"url": "//images.ctfassets.net/guia.pdf"}}},
			})
		case r.Method == http.MethodPut && strings.HasSuffix(p, "/assets/asset-1/published"):
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "asset-1", "version": 3}})
		case r.Method == http.MethodPost && strings.HasSuffix(p, "/entries"):
			assert.Equal(t, "material", r.Header.Get("X-Contentful-Content-Type"))
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "mat-1", "version": 1}})
		case r.Method == http.MethodPut && strings.HasSuffix(p, "/entries/mat-1/published"):
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "mat-1", "version": 2}})
		default:
			t.Errorf("ruta CMA inesperada: %s %s", r.Method, p)
			w.WriteHeader(http.StatusTeapot)
		}
	}))
	defer cma.Close()

	client := newTestClient("", cma.URL)
	client.uploadBaseURL = upload.URL
	repo := NewMaterialRepository(client)

	material, err := repo.UploadAndCreate(t.Context(), "Guía", "PDF", "guia.pdf", "application/pdf", []byte("%PDF-1.4 ..."))
	require.NoError(t, err)
	assert.Equal(t, "mat-1", material.ID)
	assert.Equal(t, "Guía", material.Title)
	assert.Equal(t, "PDF", material.Type)
	require.NotNil(t, material.File)
	assert.Equal(t, "asset-1", material.File.ID)
}

func TestMaterialRepository_UploadAndCreate_UploadFailurePropagates(t *testing.T) {
	upload := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"message":"upstream"}`))
	}))
	defer upload.Close()

	client := newTestClient("", "http://127.0.0.1:1") // CMA no debería llegar a usarse
	client.uploadBaseURL = upload.URL
	repo := NewMaterialRepository(client)

	material, err := repo.UploadAndCreate(t.Context(), "Guía", "PDF", "guia.pdf", "application/pdf", []byte("x"))
	assert.Nil(t, material)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no se pudo subir el archivo")
}
