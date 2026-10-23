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

func TestTrainingAreaRepository_List(t *testing.T) {
	cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "trainingArea", r.URL.Query().Get("content_type"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{"sys": map[string]any{"id": "ta1"}, "fields": map[string]any{"name": "Oficios"}},
			},
		})
	}))
	defer cda.Close()

	repo := NewTrainingAreaRepository(newTestClient(cda.URL, ""))

	areas, err := repo.List(t.Context())
	require.NoError(t, err)
	require.Len(t, areas, 1)
	assert.Equal(t, "Oficios", areas[0].Name)
}

// Create en trainingArea publica automáticamente: dos round trips
// (POST /entries, luego PUT /entries/{id}/published).
func TestTrainingAreaRepository_Create_AutoPublishes(t *testing.T) {
	var methodsCalled []string
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodsCalled = append(methodsCalled, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "ta-new", "version": 1}})
	}))
	defer cma.Close()

	repo := NewTrainingAreaRepository(newTestClient("", cma.URL))

	area, err := repo.Create(t.Context(), models.TrainingAreaInput{Name: "Oficios", Description: "desc"})
	require.NoError(t, err)
	assert.Equal(t, "ta-new", area.ID)
	assert.Equal(t, "Oficios", area.Name)

	require.Len(t, methodsCalled, 2)
	assert.Equal(t, "POST /spaces/space-1/environments/testing/entries", methodsCalled[0])
	assert.Equal(t, "PUT /spaces/space-1/environments/testing/entries/ta-new/published", methodsCalled[1])
}

func TestTrainingAreaRepository_Update_AutoPublishes(t *testing.T) {
	var methodsCalled []string
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodsCalled = append(methodsCalled, r.Method)
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "ta1", "version": 2}})
		case http.MethodPut:
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "ta1", "version": 3}})
		}
	}))
	defer cma.Close()

	repo := NewTrainingAreaRepository(newTestClient("", cma.URL))

	area, err := repo.Update(t.Context(), "ta1", models.TrainingAreaInput{Name: "Oficios 2"})
	require.NoError(t, err)
	assert.Equal(t, "Oficios 2", area.Name)
	// GET (versión actual) -> PUT (update) -> PUT (publish)
	require.Len(t, methodsCalled, 3)
}

func TestTrainingAreaRepository_Delete_UnpublishesBeforeDeleting(t *testing.T) {
	var methodsCalled []string
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodsCalled = append(methodsCalled, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "ta1"}})
	}))
	defer cma.Close()

	repo := NewTrainingAreaRepository(newTestClient("", cma.URL))

	err := repo.Delete(t.Context(), "ta1")
	require.NoError(t, err)
	require.Len(t, methodsCalled, 2)
	assert.Contains(t, methodsCalled[0], "/published")
}

func TestTrainingAreaRepository_ErrorHandling(t *testing.T) {
	t.Run("List error propagates", func(t *testing.T) {
		cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cda.Close()

		repo := NewTrainingAreaRepository(newTestClient(cda.URL, ""))
		areas, err := repo.List(t.Context())
		assert.Error(t, err)
		assert.Nil(t, areas)
	})

	t.Run("Create error propagates when CreateEntry fails", func(t *testing.T) {
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cma.Close()

		repo := NewTrainingAreaRepository(newTestClient("", cma.URL))
		area, err := repo.Create(t.Context(), models.TrainingAreaInput{Name: "Area"})
		assert.Error(t, err)
		assert.Nil(t, area)
	})

	t.Run("Create error propagates when PublishEntry fails", func(t *testing.T) {
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "ta1", "version": 1}})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cma.Close()

		repo := NewTrainingAreaRepository(newTestClient("", cma.URL))
		area, err := repo.Create(t.Context(), models.TrainingAreaInput{Name: "Area"})
		assert.Error(t, err)
		assert.Nil(t, area)
	})

	t.Run("Update error propagates when GetManagementEntry fails", func(t *testing.T) {
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cma.Close()

		repo := NewTrainingAreaRepository(newTestClient("", cma.URL))
		area, err := repo.Update(t.Context(), "ta1", models.TrainingAreaInput{Name: "Area"})
		assert.Error(t, err)
		assert.Nil(t, area)
	})

	t.Run("Update error propagates when UpdateEntry fails", func(t *testing.T) {
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "ta1", "version": 1}})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cma.Close()

		repo := NewTrainingAreaRepository(newTestClient("", cma.URL))
		area, err := repo.Update(t.Context(), "ta1", models.TrainingAreaInput{Name: "Area"})
		assert.Error(t, err)
		assert.Nil(t, area)
	})

	t.Run("Update error propagates when PublishEntry fails", func(t *testing.T) {
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case r.Method == http.MethodGet:
				_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "ta1", "version": 1}})
			case r.Method == http.MethodPut && !strings.Contains(r.URL.Path, "/published"):
				_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "ta1", "version": 2}})
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
		}))
		defer cma.Close()

		repo := NewTrainingAreaRepository(newTestClient("", cma.URL))
		area, err := repo.Update(t.Context(), "ta1", models.TrainingAreaInput{Name: "Area"})
		assert.Error(t, err)
		assert.Nil(t, area)
	})
}
