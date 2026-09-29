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
