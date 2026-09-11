package services_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"auth/internal/services"
)

func TestKVSyncService_SyncImpactStatsAsync(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewKVSyncService(mockRepo)

	// Set temporary env vars for the worker
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_ = os.Setenv("WORKER_IMPACT_URL", server.URL)
	_ = os.Setenv("WORKER_SYNC_TOKEN", "test-token")
	defer func() {
		_ = os.Unsetenv("WORKER_IMPACT_URL")
		_ = os.Unsetenv("WORKER_SYNC_TOKEN")
	}()

	mockRepo.On("CountActiveParticipants", mock.Anything).Return(int64(10), nil)
	mockRepo.On("CountVolunteers", mock.Anything).Return(int64(5), nil)

	// Since it's async, we can just call it and wait a tiny bit for the goroutine
	svc.SyncImpactStatsAsync()

	time.Sleep(100 * time.Millisecond)
	mockRepo.AssertExpectations(t)
}

func TestKVSyncService_NoWorkerURL(t *testing.T) {
	_ = os.Unsetenv("WORKER_IMPACT_URL")
	mockRepo := new(MockUserRepository)
	svc := services.NewKVSyncService(mockRepo)

	svc.SyncImpactStatsAsync()
	time.Sleep(50 * time.Millisecond)

	// Should not call anything if URL is missing
	mockRepo.AssertNotCalled(t, "CountActiveParticipants")
}

func TestKVSyncService_RepoError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewKVSyncService(mockRepo)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_ = os.Setenv("WORKER_IMPACT_URL", server.URL)
	defer func() { _ = os.Unsetenv("WORKER_IMPACT_URL") }()

	mockRepo.On("CountActiveParticipants", mock.Anything).Return(int64(0), assert.AnError)

	svc.SyncImpactStatsAsync()
	time.Sleep(50 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}

func TestKVSyncService_HttpError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	svc := services.NewKVSyncService(mockRepo)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_ = os.Setenv("WORKER_IMPACT_URL", server.URL)
	defer func() { _ = os.Unsetenv("WORKER_IMPACT_URL") }()

	mockRepo.On("CountActiveParticipants", mock.Anything).Return(int64(10), nil)
	mockRepo.On("CountVolunteers", mock.Anything).Return(int64(5), nil)

	svc.SyncImpactStatsAsync()
	time.Sleep(50 * time.Millisecond)

	mockRepo.AssertExpectations(t)
}
