package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"buenvivir/education/internal/models"
)

type mockCourseRepoForSync struct {
	mock.Mock
}

func (m *mockCourseRepoForSync) List(ctx context.Context) ([]models.Course, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Course), args.Error(1)
}

func (m *mockCourseRepoForSync) Get(ctx context.Context, id string) (*models.Course, error) {
	return nil, nil
}
func (m *mockCourseRepoForSync) Create(ctx context.Context, input models.CourseInput) (*models.Course, error) {
	return nil, nil
}
func (m *mockCourseRepoForSync) Update(ctx context.Context, id string, input models.CourseInput) (*models.Course, error) {
	return nil, nil
}
func (m *mockCourseRepoForSync) Delete(ctx context.Context, id string) error  { return nil }
func (m *mockCourseRepoForSync) Publish(ctx context.Context, id string) error { return nil }
func (m *mockCourseRepoForSync) Unpublish(ctx context.Context, id string) error {
	return nil
}
func (m *mockCourseRepoForSync) ListByResponsibleTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	return nil, nil
}
func (m *mockCourseRepoForSync) ListDraftsByResponsibleTeacher(ctx context.Context, teacherID string) ([]models.Course, error) {
	return nil, nil
}
func (m *mockCourseRepoForSync) ExistsInManagement(ctx context.Context, id string) (bool, error) {
	return false, nil
}
func (m *mockCourseRepoForSync) ListAllDrafts(ctx context.Context) ([]models.Course, error) {
	return nil, nil
}
func (m *mockCourseRepoForSync) ListByResponsibleTeachers(ctx context.Context, teacherIDs []string) ([]models.Course, error) {
	return nil, nil
}

// TestKVSyncService_SyncCourseStatsAsync_GuardClauses tests boundary guard conditions.
func TestKVSyncService_SyncCourseStatsAsync_GuardClauses(t *testing.T) {
	t.Run("Nil service receiver is safe", func(t *testing.T) {
		// Arrange
		var svc *kvSyncService

		// Act & Assert (must not panic)
		svc.SyncCourseStatsAsync()
	})

	t.Run("Nil courseRepo is safe", func(t *testing.T) {
		// Arrange
		svc := &kvSyncService{courseRepo: nil}

		// Act & Assert (must not panic)
		svc.SyncCourseStatsAsync()
	})

	t.Run("Empty WORKER_IMPACT_URL returns early", func(t *testing.T) {
		// Arrange
		t.Setenv("WORKER_IMPACT_URL", "")
		repo := new(mockCourseRepoForSync)
		svc := NewKVSyncService(repo)

		// Act
		svc.SyncCourseStatsAsync()

		// Assert: repo.List should never be called
		repo.AssertNotCalled(t, "List", mock.Anything)
	})
}

// TestKVSyncService_PerformSync_EquivalenceClassesAndBoundaries tests performSync with various HTTP responses and boundaries.
func TestKVSyncService_PerformSync_EquivalenceClassesAndBoundaries(t *testing.T) {
	t.Run("Success equivalence class with auth token and payload verification", func(t *testing.T) {
		// Arrange
		receivedAuth := ""
		var receivedPayload CourseImpactUpdatePayload

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedAuth = r.Header.Get("Authorization")
			_ = json.NewDecoder(r.Body).Decode(&receivedPayload)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		repo := new(mockCourseRepoForSync)
		courses := []models.Course{{ID: "c1"}, {ID: "c2"}, {ID: "c3"}}
		repo.On("List", mock.Anything).Return(courses, nil)

		svc := &kvSyncService{courseRepo: repo}

		// Act
		svc.performSync(server.URL, "sync-secret-123")

		// Assert
		assert.Equal(t, "Bearer sync-secret-123", receivedAuth)
		assert.Equal(t, 3, receivedPayload.Courses)
		repo.AssertExpectations(t)
	})

	t.Run("Boundary: 0 courses count", func(t *testing.T) {
		// Arrange
		var receivedPayload CourseImpactUpdatePayload
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&receivedPayload)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		repo := new(mockCourseRepoForSync)
		repo.On("List", mock.Anything).Return([]models.Course{}, nil)

		svc := &kvSyncService{courseRepo: repo}

		// Act
		svc.performSync(server.URL, "")

		// Assert
		assert.Equal(t, 0, receivedPayload.Courses)
		repo.AssertExpectations(t)
	})

	t.Run("Repo error returns early without making HTTP request", func(t *testing.T) {
		// Arrange
		serverCalled := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serverCalled = true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		repo := new(mockCourseRepoForSync)
		repo.On("List", mock.Anything).Return(nil, errors.New("database failure"))

		svc := &kvSyncService{courseRepo: repo}

		// Act
		svc.performSync(server.URL, "")

		// Assert
		assert.False(t, serverCalled, "HTTP server should not be called when repo fails")
		repo.AssertExpectations(t)
	})

	t.Run("Server 500 error handled gracefully", func(t *testing.T) {
		// Arrange
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		repo := new(mockCourseRepoForSync)
		repo.On("List", mock.Anything).Return([]models.Course{{ID: "c1"}}, nil)

		svc := &kvSyncService{courseRepo: repo}

		// Act & Assert (must not panic)
		svc.performSync(server.URL, "")
		repo.AssertExpectations(t)
	})

	t.Run("Network connection error handled gracefully", func(t *testing.T) {
		// Arrange
		repo := new(mockCourseRepoForSync)
		repo.On("List", mock.Anything).Return([]models.Course{{ID: "c1"}}, nil)

		svc := &kvSyncService{courseRepo: repo}

		// Act & Assert (invalid URL must not panic)
		svc.performSync("http://127.0.0.1:0/invalid", "")
		repo.AssertExpectations(t)
	})

	t.Run("Invalid HTTP URL parsing error handled gracefully", func(t *testing.T) {
		// Arrange
		repo := new(mockCourseRepoForSync)
		repo.On("List", mock.Anything).Return([]models.Course{{ID: "c1"}}, nil)

		svc := &kvSyncService{courseRepo: repo}

		// Act & Assert (malformed URL must fail http.NewRequestWithContext and not panic)
		svc.performSync("://invalid-url", "")
		repo.AssertExpectations(t)
	})

	t.Run("Async invocation triggers background goroutine", func(t *testing.T) {
		// Arrange
		called := make(chan bool, 1)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called <- true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		t.Setenv("WORKER_IMPACT_URL", server.URL)
		t.Setenv("WORKER_SYNC_TOKEN", "token")

		repo := new(mockCourseRepoForSync)
		repo.On("List", mock.Anything).Return([]models.Course{{ID: "c1"}}, nil)

		svc := NewKVSyncService(repo)

		// Act
		svc.SyncCourseStatsAsync()

		// Assert: wait for goroutine completion
		select {
		case <-called:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for async sync")
		}
	})
}
