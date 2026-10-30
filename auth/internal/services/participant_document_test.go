package services_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"auth/internal/models"
	"auth/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockParticipantDocRepo struct {
	mock.Mock
}

func (m *MockParticipantDocRepo) Create(ctx context.Context, doc *models.ParticipantDocument) error {
	return m.Called(ctx, doc).Error(0)
}

func (m *MockParticipantDocRepo) GetByID(ctx context.Context, id string) (*models.ParticipantDocument, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ParticipantDocument), args.Error(1)
}

func (m *MockParticipantDocRepo) GetByParticipantID(ctx context.Context, participantID string) ([]models.ParticipantDocument, error) {
	args := m.Called(ctx, participantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.ParticipantDocument), args.Error(1)
}

func (m *MockParticipantDocRepo) Update(ctx context.Context, doc *models.ParticipantDocument) error {
	return m.Called(ctx, doc).Error(0)
}

func (m *MockParticipantDocRepo) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

type MockObjectStorageService struct {
	mock.Mock
}

func (m *MockObjectStorageService) UploadFile(ctx context.Context, objectName string, content io.Reader, contentType string) (string, error) {
	args := m.Called(ctx, objectName, content, contentType)
	return args.String(0), args.Error(1)
}

func (m *MockObjectStorageService) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	args := m.Called(ctx, objectName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockObjectStorageService) DeleteFile(ctx context.Context, objectName string) error {
	return m.Called(ctx, objectName).Error(0)
}

type MockStorageCleanupRepo struct {
	mock.Mock
}

func (m *MockStorageCleanupRepo) Enqueue(ctx context.Context, task *models.StorageCleanupTask) error {
	return m.Called(ctx, task).Error(0)
}

func (m *MockStorageCleanupRepo) GetDueTasks(ctx context.Context, limit int) ([]models.StorageCleanupTask, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.StorageCleanupTask), args.Error(1)
}

func (m *MockStorageCleanupRepo) MarkProcessing(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockStorageCleanupRepo) MarkCompleted(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockStorageCleanupRepo) MarkFailed(ctx context.Context, id string, lastErr string, nextRunAt time.Time, exhausted bool) error {
	return m.Called(ctx, id, lastErr, nextRunAt, exhausted).Error(0)
}

func (m *MockStorageCleanupRepo) FindByID(ctx context.Context, id string) (*models.StorageCleanupTask, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.StorageCleanupTask), args.Error(1)
}

func TestParticipantDocument_UploadDocument_CompensatingRollback(t *testing.T) {
	// Arrange
	mockPartRepo := new(MockParticipantRepository)
	mockDocRepo := new(MockParticipantDocRepo)
	mockStorage := new(MockObjectStorageService)
	mockCleanup := new(MockStorageCleanupRepo)

	svc := services.NewParticipantService(mockPartRepo, mockDocRepo, mockStorage, mockCleanup)

	ctx := context.Background()
	participant := &models.Participant{ParticipantID: "part-1"}
	mockPartRepo.On("FindByID", ctx, "part-1").Return(participant, nil)

	mockStorage.On("UploadFile", ctx, mock.MatchedBy(func(name string) bool {
		return true
	}), mock.Anything, "application/pdf").Return("https://storage.googleapis.com/test-bucket/documents/part-1/file.pdf", nil)

	// Simulate database failure on create
	dbErr := errors.New("db insert connection lost")
	mockDocRepo.On("Create", ctx, mock.AnythingOfType("*models.ParticipantDocument")).Return(dbErr)

	// Storage delete MUST be called to compensate and rollback the orphaned file
	mockStorage.On("DeleteFile", mock.Anything, mock.MatchedBy(func(name string) bool {
		return true
	})).Return(nil)

	// Act
	fileContent := bytes.NewReader([]byte("dummy content"))
	doc, err := svc.UploadDocument(ctx, "part-1", "DNI", fileContent, "application/pdf", "dni.pdf")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, doc)
	assert.Contains(t, err.Error(), "failed to save document in database")
	mockStorage.AssertCalled(t, "DeleteFile", mock.Anything, mock.MatchedBy(func(name string) bool {
		return true
	}))
}

func TestParticipantDocument_UpdateDocument_SafeOrderingAndRollback(t *testing.T) {
	t.Run("DB Failure rolls back new upload and preserves old file", func(t *testing.T) {
		// Arrange
		mockPartRepo := new(MockParticipantRepository)
		mockDocRepo := new(MockParticipantDocRepo)
		mockStorage := new(MockObjectStorageService)
		mockCleanup := new(MockStorageCleanupRepo)

		svc := services.NewParticipantService(mockPartRepo, mockDocRepo, mockStorage, mockCleanup)

		ctx := context.Background()
		participant := &models.Participant{ParticipantID: "part-1"}
		mockPartRepo.On("FindByID", ctx, "part-1").Return(participant, nil)

		oldDoc := &models.ParticipantDocument{
			ID:            "pdoc-1",
			ParticipantID: "part-1",
			Name:          "Old Name",
			DocURL:        "https://storage.googleapis.com/test-bucket/documents/part-1/old.pdf",
		}
		mockDocRepo.On("GetByID", ctx, "pdoc-1").Return(oldDoc, nil)

		mockStorage.On("UploadFile", ctx, mock.MatchedBy(func(name string) bool {
			return true
		}), mock.Anything, "application/pdf").Return("https://storage.googleapis.com/test-bucket/documents/part-1/new.pdf", nil)

		// Simulate DB update failure
		mockDocRepo.On("Update", ctx, mock.AnythingOfType("*models.ParticipantDocument")).Return(errors.New("db update error"))

		// Rollback must delete the NEW file, never the old file
		mockStorage.On("DeleteFile", mock.Anything, mock.MatchedBy(func(name string) bool {
			return name != "documents/part-1/old.pdf"
		})).Return(nil)

		// Act
		fileContent := bytes.NewReader([]byte("new content"))
		updated, err := svc.UpdateDocument(ctx, "part-1", "pdoc-1", "New Name", fileContent, "application/pdf", "new.pdf")

		// Assert
		assert.Error(t, err)
		assert.Nil(t, updated)
		mockStorage.AssertNotCalled(t, "DeleteFile", mock.Anything, "documents/part-1/old.pdf")
	})

	t.Run("DB Success purges old file and enqueues cleanup on storage delete failure", func(t *testing.T) {
		// Arrange
		mockPartRepo := new(MockParticipantRepository)
		mockDocRepo := new(MockParticipantDocRepo)
		mockStorage := new(MockObjectStorageService)
		mockCleanup := new(MockStorageCleanupRepo)

		svc := services.NewParticipantService(mockPartRepo, mockDocRepo, mockStorage, mockCleanup)

		ctx := context.Background()
		participant := &models.Participant{ParticipantID: "part-1"}
		mockPartRepo.On("FindByID", ctx, "part-1").Return(participant, nil)

		oldDoc := &models.ParticipantDocument{
			ID:            "pdoc-1",
			ParticipantID: "part-1",
			Name:          "Old Name",
			DocURL:        "https://storage.googleapis.com/test-bucket/documents/part-1/old.pdf",
		}
		mockDocRepo.On("GetByID", ctx, "pdoc-1").Return(oldDoc, nil)

		mockStorage.On("UploadFile", ctx, mock.Anything, mock.Anything, "application/pdf").Return("https://storage.googleapis.com/test-bucket/documents/part-1/new.pdf", nil)
		mockDocRepo.On("Update", ctx, mock.AnythingOfType("*models.ParticipantDocument")).Return(nil)

		// Old file deletion fails due to network glitch
		mockStorage.On("DeleteFile", ctx, "documents/part-1/old.pdf").Return(errors.New("storage network timeout"))

		// Outbox cleanup task MUST be enqueued
		mockCleanup.On("Enqueue", ctx, mock.MatchedBy(func(task *models.StorageCleanupTask) bool {
			return task.ObjectName == "documents/part-1/old.pdf" && task.Status == models.CleanupStatusPending
		})).Return(nil)

		// Act
		fileContent := bytes.NewReader([]byte("new content"))
		updated, err := svc.UpdateDocument(ctx, "part-1", "pdoc-1", "New Name", fileContent, "application/pdf", "new.pdf")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		mockCleanup.AssertCalled(t, "Enqueue", ctx, mock.AnythingOfType("*models.StorageCleanupTask"))
	})
}

func TestParticipantDocument_DeleteDocument_SafeOrdering(t *testing.T) {
	// Arrange
	mockPartRepo := new(MockParticipantRepository)
	mockDocRepo := new(MockParticipantDocRepo)
	mockStorage := new(MockObjectStorageService)
	mockCleanup := new(MockStorageCleanupRepo)

	svc := services.NewParticipantService(mockPartRepo, mockDocRepo, mockStorage, mockCleanup)

	ctx := context.Background()
	participant := &models.Participant{ParticipantID: "part-1"}
	mockPartRepo.On("FindByID", ctx, "part-1").Return(participant, nil)

	doc := &models.ParticipantDocument{
		ID:            "pdoc-1",
		ParticipantID: "part-1",
		DocURL:        "https://storage.googleapis.com/test-bucket/documents/part-1/file.pdf",
	}
	mockDocRepo.On("GetByID", ctx, "pdoc-1").Return(doc, nil)
	mockDocRepo.On("Delete", ctx, "pdoc-1").Return(nil)
	mockStorage.On("DeleteFile", ctx, "documents/part-1/file.pdf").Return(nil)

	// Act
	err := svc.DeleteDocument(ctx, "part-1", "pdoc-1")

	// Assert
	assert.NoError(t, err)
	mockDocRepo.AssertCalled(t, "Delete", ctx, "pdoc-1")
	mockStorage.AssertCalled(t, "DeleteFile", ctx, "documents/part-1/file.pdf")
}
