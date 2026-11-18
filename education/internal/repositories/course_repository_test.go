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

// newTestClient apunta el ContentfulClient a servidores httptest en vez
// de a los hosts reales de Contentful, para poder testear los
// repositories sin red ni credenciales reales.
func newTestClient(cdaURL, cmaURL string) *ContentfulClient {
	return &ContentfulClient{
		spaceID:     "space-1",
		environment: "testing",
		cdaToken:    "cda-token",
		cmaToken:    "cma-token",
		locale:      "en-US",
		httpClient:  http.DefaultClient,
		cdaBaseURL:  cdaURL,
		cmaBaseURL:  cmaURL,
	}
}

func TestCourseRepository_List(t *testing.T) {
	cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/spaces/space-1/environments/testing/entries", r.URL.Path)
		assert.Equal(t, "course", r.URL.Query().Get("content_type"))
		assert.Equal(t, "cda-token", r.URL.Query().Get("access_token"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"sys":    map[string]any{"id": "c1"},
					"fields": map[string]any{"title": "Curso 1", "mode": "Presencial", "duration": "4 semanas"},
				},
			},
			"includes": map[string]any{},
		})
	}))
	defer cda.Close()

	client := newTestClient(cda.URL, "")
	repo := NewCourseRepository(client)

	courses, err := repo.List(t.Context())
	require.NoError(t, err)
	require.Len(t, courses, 1)
	assert.Equal(t, "c1", courses[0].ID)
	assert.Equal(t, "Curso 1", courses[0].Title)
	assert.Equal(t, "Presencial", courses[0].Mode)
}

func TestCourseRepository_Get_NotFound(t *testing.T) {
	cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "c-missing", r.URL.Query().Get("sys.id"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items":    []map[string]any{},
			"includes": map[string]any{},
		})
	}))
	defer cda.Close()

	client := newTestClient(cda.URL, "")
	repo := NewCourseRepository(client)

	course, err := repo.Get(t.Context(), "c-missing")
	assert.Error(t, err)
	assert.Nil(t, course)
	assert.True(t, IsNotFound(err))
}

func TestCourseRepository_Create(t *testing.T) {
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/spaces/space-1/environments/testing/entries", r.URL.Path)
		assert.Equal(t, "course", r.Header.Get("X-Contentful-Content-Type"))
		assert.Equal(t, "Bearer cma-token", r.Header.Get("Authorization"))

		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		fields := body["fields"].(map[string]any)
		title := fields["title"].(map[string]any)
		assert.Equal(t, "Curso nuevo", title["en-US"])

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"sys": map[string]any{"id": "c-new", "version": 1},
		})
	}))
	defer cma.Close()

	client := newTestClient("", cma.URL)
	repo := NewCourseRepository(client)

	input := models.CourseInput{
		Title:              "Curso nuevo",
		Description:        json.RawMessage(`{"nodeType":"document"}`),
		ResponsibleTeacher: "teacher-1",
		Mode:               "Presencial",
		Duration:           "4 semanas",
		TrainingAreaID:     "area-1",
		MaterialIDs:        []string{"mat-1"},
	}

	course, err := repo.Create(t.Context(), input)
	require.NoError(t, err)
	assert.Equal(t, "c-new", course.ID)
	assert.Equal(t, "Curso nuevo", course.Title)
}

func TestCourseRepository_Update_UsesCurrentVersion(t *testing.T) {
	callCount := 0
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"sys": map[string]any{"id": "c1", "version": 3},
			})
		case http.MethodPut:
			assert.Equal(t, "3", r.Header.Get("X-Contentful-Version"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"sys": map[string]any{"id": "c1", "version": 4},
			})
		}
	}))
	defer cma.Close()

	client := newTestClient("", cma.URL)
	repo := NewCourseRepository(client)

	input := models.CourseInput{
		Title: "Curso editado", Description: json.RawMessage(`{}`),
		ResponsibleTeacher: "teacher-1", Mode: "Virtual", Duration: "2 semanas",
		TrainingAreaID: "area-1", MaterialIDs: []string{"mat-1"},
	}

	course, err := repo.Update(t.Context(), "c1", input)
	require.NoError(t, err)
	assert.Equal(t, "Curso editado", course.Title)
	assert.Equal(t, 2, callCount)
}

func TestCourseRepository_Publish(t *testing.T) {
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "c1", "version": 2}})
		case r.Method == http.MethodPut && r.URL.Path == "/spaces/space-1/environments/testing/entries/c1/published":
			assert.Equal(t, "2", r.Header.Get("X-Contentful-Version"))
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "c1", "version": 3, "publishedVersion": 3}})
		}
	}))
	defer cma.Close()

	client := newTestClient("", cma.URL)
	repo := NewCourseRepository(client)

	err := repo.Publish(t.Context(), "c1")
	assert.NoError(t, err)
}

func TestCourseRepository_Unpublish_Idempotent(t *testing.T) {
	// Contentful responde 404 si ya estaba despublicada; el repository
	// lo trata como éxito (idempotente), ver contentful_client.go.
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))
	defer cma.Close()

	client := newTestClient("", cma.URL)
	repo := NewCourseRepository(client)

	err := repo.Unpublish(t.Context(), "c1")
	assert.NoError(t, err)
}

func TestCourseRepository_Delete_UnpublishesBeforeDeleting(t *testing.T) {
	var methodsCalled []string
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodsCalled = append(methodsCalled, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "c1"}})
	}))
	defer cma.Close()

	client := newTestClient("", cma.URL)
	repo := NewCourseRepository(client)

	err := repo.Delete(t.Context(), "c1")
	require.NoError(t, err)
	require.Len(t, methodsCalled, 2)
	assert.Equal(t, "DELETE /spaces/space-1/environments/testing/entries/c1/published", methodsCalled[0])
	assert.Equal(t, "DELETE /spaces/space-1/environments/testing/entries/c1", methodsCalled[1])
}

// Un borrador nunca publicado: la CMA responde 400 al despublicar. Hay
// que confirmar que no tiene publishedVersion y borrarlo igual.
func TestCourseRepository_Delete_DraftNeverPublished(t *testing.T) {
	var methodsCalled []string
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodsCalled = append(methodsCalled, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/published"):
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"sys":{"type":"Error","id":"BadRequest"},"message":"Not published"}`))
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "c1", "version": 3}})
		}
	}))
	defer cma.Close()

	repo := NewCourseRepository(newTestClient("", cma.URL))

	err := repo.Delete(t.Context(), "c1")
	require.NoError(t, err)
	assert.Equal(t, []string{
		"DELETE /spaces/space-1/environments/testing/entries/c1/published",
		"GET /spaces/space-1/environments/testing/entries/c1",
		"DELETE /spaces/space-1/environments/testing/entries/c1",
	}, methodsCalled)
}

// Si el 400 al despublicar NO es porque sea un borrador (la entry sigue
// publicada), no se intenta borrar.
func TestCourseRepository_Delete_UnpublishFailsOnPublishedEntry(t *testing.T) {
	var methodsCalled []string
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methodsCalled = append(methodsCalled, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"message":"otro motivo"}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "c1", "version": 3, "publishedVersion": 2}})
	}))
	defer cma.Close()

	repo := NewCourseRepository(newTestClient("", cma.URL))

	err := repo.Delete(t.Context(), "c1")
	require.Error(t, err)
	assert.Len(t, methodsCalled, 2)
}

func TestValidationMessage_MissingFields(t *testing.T) {
	err := &ContentfulAPIError{StatusCode: http.StatusUnprocessableEntity, Body: `{
		"sys":{"type":"Error","id":"ValidationFailed"},
		"details":{"errors":[
			{"name":"required","path":["fields","mode"]},
			{"name":"required","path":["fields","trainingArea"]},
			{"name":"notResolvable","path":["fields","material","en-US",0]}
		]}}`}

	msg, ok := ValidationMessage(err)
	require.True(t, ok)
	assert.Contains(t, msg, "faltan campos obligatorios: modalidad, área formativa")
	assert.Contains(t, msg, "referencias borradas o sin publicar en: materiales")

	_, ok = ValidationMessage(&ContentfulAPIError{StatusCode: http.StatusBadRequest})
	assert.False(t, ok)
}

func TestCourseRepository_ListByResponsibleTeacher_FiltersByTeacher(t *testing.T) {
	cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "teacher-1", r.URL.Query().Get("fields.responsibleTeacher"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []map[string]any{}, "includes": map[string]any{}})
	}))
	defer cda.Close()

	client := newTestClient(cda.URL, "")
	repo := NewCourseRepository(client)

	courses, err := repo.ListByResponsibleTeacher(t.Context(), "teacher-1")
	require.NoError(t, err)
	assert.Len(t, courses, 0)
}

func TestCourseRepository_ListDraftsByResponsibleTeacher_OnlyUnpublished(t *testing.T) {
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "teacher-1", r.URL.Query().Get("fields.responsibleTeacher"))
		assert.Equal(t, "false", r.URL.Query().Get("sys.publishedVersion[exists]"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"sys": map[string]any{"id": "draft-1"},
					"fields": map[string]any{
						"title":              map[string]any{"en-US": "Borrador"},
						"responsibleTeacher": map[string]any{"en-US": "teacher-1"},
						"mode":               map[string]any{"en-US": "Virtual"},
					},
				},
			},
		})
	}))
	defer cma.Close()

	client := newTestClient("", cma.URL)
	repo := NewCourseRepository(client)

	courses, err := repo.ListDraftsByResponsibleTeacher(t.Context(), "teacher-1")
	require.NoError(t, err)
	require.Len(t, courses, 1)
	assert.Equal(t, "draft-1", courses[0].ID)
	assert.Equal(t, "Borrador", courses[0].Title)
	assert.Equal(t, "teacher-1", courses[0].ResponsibleTeacher)
}

func TestCourseRepository_ListAllDrafts(t *testing.T) {
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "false", r.URL.Query().Get("sys.publishedVersion[exists]"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"items": []map[string]any{}})
	}))
	defer cma.Close()

	client := newTestClient("", cma.URL)
	repo := NewCourseRepository(client)

	courses, err := repo.ListAllDrafts(t.Context())
	require.NoError(t, err)
	assert.Len(t, courses, 0)
}

func TestCourseRepository_Get_ResolvesIncludes(t *testing.T) {
	cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "c1", r.URL.Query().Get("sys.id"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"sys": map[string]any{"id": "c1"},
					"fields": map[string]any{
						"title": "Curso con área",
						"mode":  "Mixta",
						"trainingArea": map[string]any{
							"sys": map[string]any{"type": "Link", "linkType": "Entry", "id": "area-1"},
						},
					},
				},
			},
			"includes": map[string]any{
				"Entry": []map[string]any{
					{
						"sys":    map[string]any{"id": "area-1"},
						"fields": map[string]any{"name": "Oficios"},
					},
				},
			},
		})
	}))
	defer cda.Close()

	client := newTestClient(cda.URL, "")
	repo := NewCourseRepository(client)

	course, err := repo.Get(t.Context(), "c1")
	require.NoError(t, err)
	require.NotNil(t, course.TrainingArea)
	assert.Equal(t, "area-1", course.TrainingArea.ID)
	assert.Equal(t, "Oficios", course.TrainingArea.Name)
}

func TestCourseRepository_Create_APIErrorPropagatesStatusCode(t *testing.T) {
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"message":"validation failed"}`))
	}))
	defer cma.Close()

	client := newTestClient("", cma.URL)
	repo := NewCourseRepository(client)

	course, err := repo.Create(t.Context(), models.CourseInput{Title: "x"})
	assert.Error(t, err)
	assert.Nil(t, course)

	apiErr, ok := err.(*ContentfulAPIError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.StatusCode)
}

func TestCourseRepository_ExistsInManagement(t *testing.T) {
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/entries/existe"):
			_, _ = w.Write([]byte(`{"sys":{"id":"existe","version":1},"fields":{}}`))
		case strings.HasSuffix(r.URL.Path, "/entries/roto"):
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer cma.Close()
	repo := NewCourseRepository(newTestClient("", cma.URL))

	exists, err := repo.ExistsInManagement(t.Context(), "existe")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsInManagement(t.Context(), "no-existe")
	require.NoError(t, err)
	assert.False(t, exists)

	_, err = repo.ExistsInManagement(t.Context(), "roto")
	assert.Error(t, err)
}

func TestCourseRepository_ListByResponsibleTeachers_BoundariesAndSuccess(t *testing.T) {
	t.Run("Boundary: empty teacherIDs slice returns empty without API calls", func(t *testing.T) {
		// Arrange
		repo := NewCourseRepository(newTestClient("", ""))

		// Act
		courses, err := repo.ListByResponsibleTeachers(t.Context(), []string{})

		// Assert
		require.NoError(t, err)
		assert.Empty(t, courses)
	})

	t.Run("Success: chunks queries and collects courses", func(t *testing.T) {
		// Arrange
		cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.Query().Get("fields.responsibleTeacher[in]"), "t1,t2")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{
						"sys":    map[string]any{"id": "c1"},
						"fields": map[string]any{"title": "Curso Docente", "responsibleTeacher": "t1"},
					},
				},
				"includes": map[string]any{},
			})
		}))
		defer cda.Close()

		repo := NewCourseRepository(newTestClient(cda.URL, ""))

		// Act
		courses, err := repo.ListByResponsibleTeachers(t.Context(), []string{"t1", "t2"})

		// Assert
		require.NoError(t, err)
		require.Len(t, courses, 1)
		assert.Equal(t, "c1", courses[0].ID)
	})

	t.Run("Error: CDA error propagates", func(t *testing.T) {
		// Arrange
		cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cda.Close()

		repo := NewCourseRepository(newTestClient(cda.URL, ""))

		// Act
		courses, err := repo.ListByResponsibleTeachers(t.Context(), []string{"t1"})

		// Assert
		assert.Error(t, err)
		assert.Nil(t, courses)
	})
}

// TestCourseInputToFields_EquivalenceClasses tests equivalence classes and boundary values for input mapping.
func TestCourseInputToFields_EquivalenceClasses(t *testing.T) {
	cap50 := 50
	t.Run("Class 1: Full fields with capacity and cover image", func(t *testing.T) {
		// Arrange
		input := models.CourseInput{
			Title:              "Full Course",
			Description:        json.RawMessage(`"Desc"`),
			ResponsibleTeacher: "t1",
			Mode:               "Presencial",
			Duration:           "2m",
			TrainingAreaID:     "ta1",
			CoverImageAssetID:  "img1",
			Capacity:           &cap50,
			MaterialIDs:        []string{"m1", "m2"},
		}

		// Act
		fields := courseInputToFields("en-US", input)

		// Assert
		assert.Equal(t, map[string]any{"en-US": "Full Course"}, fields["title"])
		assert.Equal(t, map[string]any{"en-US": 50}, fields["capacity"])
		assert.Contains(t, fields, "coverImage")
		assert.Contains(t, fields, "material")
	})

	t.Run("Class 2: Boundary with nil capacity and empty cover image", func(t *testing.T) {
		// Arrange
		input := models.CourseInput{
			Title:              "Minimal Course",
			Description:        json.RawMessage(`"Desc"`),
			ResponsibleTeacher: "t1",
			Mode:               "Online",
			Duration:           "1m",
			TrainingAreaID:     "ta1",
			CoverImageAssetID:  "",
			Capacity:           nil,
			MaterialIDs:        nil,
		}

		// Act
		fields := courseInputToFields("en-US", input)

		// Assert
		assert.NotContains(t, fields, "capacity")
		assert.NotContains(t, fields, "coverImage")
		assert.Empty(t, fields["material"].(map[string]any)["en-US"].([]models.ContentfulLink))
	})
}

// TestCourseRepository_Get_FullIncludes tests CDA mapping with all includes resolved (assets, trainingArea, material).
func TestCourseRepository_Get_FullIncludes(t *testing.T) {
	// Arrange
	cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"sys": map[string]any{"id": "c1"},
					"fields": map[string]any{
						"title":        "Curso Completo",
						"capacity":     float64(30),
						"coverImage":   map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Asset", "id": "asset1"}},
						"trainingArea": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Entry", "id": "area1"}},
						"material": []map[string]any{
							{"sys": map[string]any{"type": "Link", "linkType": "Entry", "id": "mat1"}},
						},
					},
				},
			},
			"includes": map[string]any{
				"Entry": []map[string]any{
					{"sys": map[string]any{"id": "area1"}, "fields": map[string]any{"name": "Área Tech"}},
					{"sys": map[string]any{"id": "mat1"}, "fields": map[string]any{"title": "Guía 1", "type": "Document"}},
				},
				"Asset": []map[string]any{
					{"sys": map[string]any{"id": "asset1"}, "fields": map[string]any{"title": "Cover", "file": map[string]any{"url": "//cdn.example.com/cover.png"}}},
				},
			},
		})
	}))
	defer cda.Close()

	repo := NewCourseRepository(newTestClient(cda.URL, ""))

	// Act
	course, err := repo.Get(t.Context(), "c1")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, course.Capacity)
	assert.Equal(t, 30, *course.Capacity)
	require.NotNil(t, course.CoverImage)
	assert.Equal(t, "https://cdn.example.com/cover.png", course.CoverImage.URL)
	require.NotNil(t, course.TrainingArea)
	assert.Equal(t, "Área Tech", course.TrainingArea.Name)
	require.Len(t, course.Material, 1)
	assert.Equal(t, "Guía 1", course.Material[0].Title)
}

// TestCourseRepository_ListDrafts_FullReferences tests CMA mapping with raw link references.
func TestCourseRepository_ListDrafts_FullReferences(t *testing.T) {
	// Arrange
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"items": []map[string]any{
				{
					"sys": map[string]any{"id": "d1"},
					"fields": map[string]any{
						"title":        map[string]any{"en-US": "Draft 1"},
						"capacity":     map[string]any{"en-US": float64(20)},
						"coverImage":   map[string]any{"en-US": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Asset", "id": "asset1"}}},
						"trainingArea": map[string]any{"en-US": map[string]any{"sys": map[string]any{"type": "Link", "linkType": "Entry", "id": "area1"}}},
						"material": map[string]any{"en-US": []map[string]any{
							{"sys": map[string]any{"type": "Link", "linkType": "Entry", "id": "mat1"}},
						}},
					},
				},
			},
		})
	}))
	defer cma.Close()

	repo := NewCourseRepository(newTestClient("", cma.URL))

	// Act
	drafts, err := repo.ListDraftsByResponsibleTeacher(t.Context(), "t1")

	// Assert
	require.NoError(t, err)
	require.Len(t, drafts, 1)
	require.NotNil(t, drafts[0].Capacity)
	assert.Equal(t, 20, *drafts[0].Capacity)
	require.NotNil(t, drafts[0].CoverImage)
	assert.Equal(t, "asset1", drafts[0].CoverImage.ID)
	require.NotNil(t, drafts[0].TrainingArea)
	assert.Equal(t, "area1", drafts[0].TrainingArea.ID)
	require.Len(t, drafts[0].Material, 1)
	assert.Equal(t, "mat1", drafts[0].Material[0].ID)
}

// TestCourseRepository_Publish_DecisionTable tests Publish success and error propagation.
func TestCourseRepository_Publish_DecisionTable(t *testing.T) {
	t.Run("Success publishes entry", func(t *testing.T) {
		// Arrange
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"sys": map[string]any{"id": "c1", "version": 2},
				})
				return
			}
			if r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/published") {
				assert.Equal(t, "2", r.Header.Get("X-Contentful-Version"))
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{
					"sys": map[string]any{"id": "c1", "version": 3},
				})
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer cma.Close()

		repo := NewCourseRepository(newTestClient("", cma.URL))

		// Act
		err := repo.Publish(t.Context(), "c1")

		// Assert
		require.NoError(t, err)
	})

	t.Run("Error: GetManagementEntry failure propagates", func(t *testing.T) {
		// Arrange
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cma.Close()

		repo := NewCourseRepository(newTestClient("", cma.URL))

		// Act
		err := repo.Publish(t.Context(), "c1")

		// Assert
		assert.Error(t, err)
	})
}

// TestCourseRepository_ListByResponsibleTeacher tests filtering by a single teacher.
func TestCourseRepository_ListByResponsibleTeacher(t *testing.T) {
	t.Run("Success lists courses for teacher", func(t *testing.T) {
		// Arrange
		cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "t1", r.URL.Query().Get("fields.responsibleTeacher"))
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"items": []map[string]any{
					{"sys": map[string]any{"id": "c1"}, "fields": map[string]any{"title": "Curso 1"}},
				},
				"includes": map[string]any{},
			})
		}))
		defer cda.Close()

		repo := NewCourseRepository(newTestClient(cda.URL, ""))

		// Act
		courses, err := repo.ListByResponsibleTeacher(t.Context(), "t1")

		// Assert
		require.NoError(t, err)
		require.Len(t, courses, 1)
		assert.Equal(t, "c1", courses[0].ID)
	})

	t.Run("Error: CDA error propagates", func(t *testing.T) {
		// Arrange
		cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cda.Close()

		repo := NewCourseRepository(newTestClient(cda.URL, ""))

		// Act
		courses, err := repo.ListByResponsibleTeacher(t.Context(), "t1")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, courses)
	})
}

func TestCourseRepository_Update_Failure(t *testing.T) {
	// Arrange
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "c1", "version": 1}})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer cma.Close()

	repo := NewCourseRepository(newTestClient("", cma.URL))

	// Act
	course, err := repo.Update(t.Context(), "c1", models.CourseInput{Title: "Title"})

	// Assert
	assert.Error(t, err)
	assert.Nil(t, course)
}

func TestCourseRepository_Delete_Error(t *testing.T) {
	// Arrange
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer cma.Close()

	repo := NewCourseRepository(newTestClient("", cma.URL))

	// Act
	err := repo.Delete(t.Context(), "c1")

	// Assert
	assert.Error(t, err)
}

func TestCourseRepository_Publish_Errors(t *testing.T) {
	t.Run("GetManagementEntry error returns error", func(t *testing.T) {
		// Arrange
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer cma.Close()

		repo := NewCourseRepository(newTestClient("", cma.URL))

		// Act
		err := repo.Publish(t.Context(), "c1")

		// Assert
		assert.Error(t, err)
	})

	t.Run("PublishEntry error returns error", func(t *testing.T) {
		// Arrange
		cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "c1", "version": 1}})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer cma.Close()

		repo := NewCourseRepository(newTestClient("", cma.URL))

		// Act
		err := repo.Publish(t.Context(), "c1")

		// Assert
		assert.Error(t, err)
	})
}

func TestCourseRepository_ListAllDrafts_Error(t *testing.T) {
	// Arrange
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer cma.Close()

	repo := NewCourseRepository(newTestClient("", cma.URL))

	// Act
	drafts, err := repo.ListAllDrafts(t.Context())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, drafts)
}

func TestCourseRepository_ListDraftsByResponsibleTeacher_Error(t *testing.T) {
	// Arrange
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer cma.Close()

	repo := NewCourseRepository(newTestClient("", cma.URL))

	// Act
	drafts, err := repo.ListDraftsByResponsibleTeacher(t.Context(), "t1")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, drafts)
}

func TestCourseRepository_CourseFromInput_Complete(t *testing.T) {
	// Arrange
	sys := models.ContentfulSys{
		ID:        "c-complete",
		CreatedAt: "2026-10-01T00:00:00Z",
		UpdatedAt: "2026-10-02T00:00:00Z",
	}
	capacity := 30
	input := models.CourseInput{
		Title:              "Complete Course",
		Description:        json.RawMessage(`{"nodeType":"document"}`),
		ResponsibleTeacher: "teacher-1",
		Mode:               "Presencial",
		Duration:           "4 semanas",
		Capacity:           &capacity,
		CoverImageAssetID:  "asset-cover",
		TrainingAreaID:     "area-formativa-1",
		MaterialIDs:        []string{"mat-1", "mat-2"},
	}

	// Act
	course := courseFromInput(sys, input)

	// Assert
	assert.Equal(t, "c-complete", course.ID)
	assert.Equal(t, "Complete Course", course.Title)
	require.NotNil(t, course.CoverImage)
	assert.Equal(t, "asset-cover", course.CoverImage.ID)
	require.NotNil(t, course.TrainingArea)
	assert.Equal(t, "area-formativa-1", course.TrainingArea.ID)
	require.Len(t, course.Material, 2)
	assert.Equal(t, "mat-1", course.Material[0].ID)
	assert.Equal(t, "mat-2", course.Material[1].ID)
}

func TestCourseRepository_CourseFromCMAEntry_Complete(t *testing.T) {
	// Arrange
	locale := "es-AR"
	rawJSON := `{
		"sys": {
			"id": "cma-course-1",
			"createdAt": "2026-10-01T00:00:00Z",
			"updatedAt": "2026-10-02T00:00:00Z"
		},
		"fields": {
			"title": {"es-AR": "CMA Course Title"},
			"description": {"es-AR": "CMA Course Description"},
			"responsibleTeacher": {"es-AR": "teacher-cma-1"},
			"mode": {"es-AR": "Virtual"},
			"duration": {"es-AR": "8 semanas"},
			"capacity": {"es-AR": 25},
			"coverImage": {"es-AR": {"sys": {"type": "Link", "linkType": "Asset", "id": "asset-cma-cover"}}},
			"trainingArea": {"es-AR": {"sys": {"type": "Link", "linkType": "Entry", "id": "ta-cma-1"}}},
			"material": {"es-AR": [
				{"sys": {"type": "Link", "linkType": "Entry", "id": "mat-cma-1"}},
				{"sys": {"type": "Link", "linkType": "Entry", "id": "mat-cma-2"}}
			]}
		}
	}`
	var entry models.ContentfulEntry
	require.NoError(t, json.Unmarshal([]byte(rawJSON), &entry))

	// Act
	course := courseFromCMAEntry(locale, entry)

	// Assert
	assert.Equal(t, "cma-course-1", course.ID)
	assert.Equal(t, "CMA Course Title", course.Title)
	assert.Equal(t, json.RawMessage(`"CMA Course Description"`), course.Description)
	assert.Equal(t, "teacher-cma-1", course.ResponsibleTeacher)
	assert.Equal(t, "Virtual", course.Mode)
	assert.Equal(t, "8 semanas", course.Duration)
	require.NotNil(t, course.Capacity)
	assert.Equal(t, 25, *course.Capacity)
	require.NotNil(t, course.CoverImage)
	assert.Equal(t, "asset-cma-cover", course.CoverImage.ID)
	require.NotNil(t, course.TrainingArea)
	assert.Equal(t, "ta-cma-1", course.TrainingArea.ID)
	require.Len(t, course.Material, 2)
	assert.Equal(t, "mat-cma-1", course.Material[0].ID)
	assert.Equal(t, "mat-cma-2", course.Material[1].ID)
}

func TestCourseRepository_List_Error(t *testing.T) {
	// Arrange
	cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer cda.Close()

	client := newTestClient(cda.URL, "")
	repo := NewCourseRepository(client)

	// Act
	courses, err := repo.List(t.Context())

	// Assert
	assert.Error(t, err)
	assert.Nil(t, courses)
}

func TestCourseRepository_Get_Error(t *testing.T) {
	// Arrange
	cda := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer cda.Close()

	client := newTestClient(cda.URL, "")
	repo := NewCourseRepository(client)

	// Act
	course, err := repo.Get(t.Context(), "c1")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, course)
}

// Editar un curso publicado lo vuelve a publicar con la versión nueva,
// así el catálogo (CDA) muestra el cambio, por ejemplo el docente nuevo
// (17/11/2026). El borrador (test anterior) no se publica.
func TestCourseRepository_Update_RepublicaSiEstabaPublicado(t *testing.T) {
	publicado := false
	cma := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "c1", "version": 6, "publishedVersion": 5}})
		case r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/entries/c1/published"):
			assert.Equal(t, "7", r.Header.Get("X-Contentful-Version"))
			publicado = true
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "c1", "version": 8, "publishedVersion": 7}})
		case r.Method == http.MethodPut:
			assert.Equal(t, "6", r.Header.Get("X-Contentful-Version"))
			_ = json.NewEncoder(w).Encode(map[string]any{"sys": map[string]any{"id": "c1", "version": 7, "publishedVersion": 5}})
		}
	}))
	defer cma.Close()

	repo := NewCourseRepository(newTestClient("", cma.URL))
	input := models.CourseInput{
		Title: "Curso", Description: json.RawMessage(`{}`),
		ResponsibleTeacher: "docente-nuevo", Mode: "Virtual", Duration: "2 semanas",
		TrainingAreaID: "area-1",
	}
	course, err := repo.Update(t.Context(), "c1", input)
	require.NoError(t, err)
	assert.True(t, publicado, "debía volver a publicar el curso")
	assert.Equal(t, "docente-nuevo", course.ResponsibleTeacher)
}
