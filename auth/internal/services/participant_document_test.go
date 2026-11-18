package services_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"auth/internal/middlewares"
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

func TestParticipantDocument_GetDocuments_Access(t *testing.T) {
	mockPartRepo := new(MockParticipantRepository)
	mockDocRepo := new(MockParticipantDocRepo)
	svc := services.NewParticipantService(mockPartRepo, mockDocRepo, nil, nil)

	// Participant not found
	mockPartRepo.On("FindByID", mock.Anything, "p-missing").Return(nil, errors.New("not found"))
	_, err := svc.GetDocuments(context.Background(), "p-missing")
	assert.Error(t, err)

	// Forbidden: participant tries to read another participant's docs
	part := &models.Participant{ParticipantID: "p-other", UserID: "usr-other"}
	mockPartRepo.On("FindByID", mock.Anything, "p-other").Return(part, nil)
	ctxPart := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_participant")
	ctxPart = context.WithValue(ctxPart, middlewares.UserIDKey, "usr-me")
	_, err = svc.GetDocuments(ctxPart, "p-other")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden")

	// Success: participant reads their own docs
	ctxOwn := context.WithValue(context.Background(), middlewares.RoleIDKey, "role_participant")
	ctxOwn = context.WithValue(ctxOwn, middlewares.UserIDKey, "usr-other")
	mockDocRepo.On("GetByParticipantID", ctxOwn, "p-other").Return([]models.ParticipantDocument{{ID: "doc-1"}}, nil)
	docs, err := svc.GetDocuments(ctxOwn, "p-other")
	assert.NoError(t, err)
	assert.Len(t, docs, 1)
}

func TestParticipantDocument_UploadDocument_Errors(t *testing.T) {
	mockPartRepo := new(MockParticipantRepository)
	mockDocRepo := new(MockParticipantDocRepo)
	mockStorage := new(MockObjectStorageService)
	mockCleanup := new(MockStorageCleanupRepo)
	svc := services.NewParticipantService(mockPartRepo, mockDocRepo, mockStorage, mockCleanup)

	ctx := context.Background()
	part := &models.Participant{ParticipantID: "p-1"}
	mockPartRepo.On("FindByID", ctx, "p-1").Return(part, nil)

	// Storage upload error
	mockStorage.On("UploadFile", ctx, mock.Anything, mock.Anything, "application/pdf").
		Return("", errors.New("gcs down")).Once()
	_, err := svc.UploadDocument(ctx, "p-1", "DNI", bytes.NewReader([]byte("data")), "application/pdf", "dni.pdf")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error uploading document")

	// DB Create error triggers rollback
	mockStorage.On("UploadFile", ctx, mock.Anything, mock.Anything, "application/pdf").
		Return("https://storage/documents/p-1/1_dni.pdf", nil).Once()
	mockDocRepo.On("Create", ctx, mock.AnythingOfType("*models.ParticipantDocument")).
		Return(errors.New("db insert fail")).Once()
	mockStorage.On("DeleteFile", mock.Anything, mock.Anything).
		Return(errors.New("delete failed too")).Once()
	mockCleanup.On("Enqueue", mock.Anything, mock.AnythingOfType("*models.StorageCleanupTask")).Return(nil).Once()

	_, err = svc.UploadDocument(ctx, "p-1", "DNI", bytes.NewReader([]byte("data")), "application/pdf", "dni.pdf")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save document in database")
}

func TestParticipantDocument_UpdateAndDelete_Errors(t *testing.T) {
	mockPartRepo := new(MockParticipantRepository)
	mockDocRepo := new(MockParticipantDocRepo)
	mockStorage := new(MockObjectStorageService)
	svc := services.NewParticipantService(mockPartRepo, mockDocRepo, mockStorage, nil)

	ctx := context.Background()
	part := &models.Participant{ParticipantID: "p-1"}
	mockPartRepo.On("FindByID", ctx, "p-1").Return(part, nil)

	// Update: doc does not belong to participant
	wrongDoc := &models.ParticipantDocument{ID: "d-wrong", ParticipantID: "p-other"}
	mockDocRepo.On("GetByID", ctx, "d-wrong").Return(wrongDoc, nil)
	_, err := svc.UpdateDocument(ctx, "p-1", "d-wrong", "Name", nil, "", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not belong to participant")

	// Update: name only, no new file
	goodDoc := &models.ParticipantDocument{ID: "d-good", ParticipantID: "p-1", Name: "Old"}
	mockDocRepo.On("GetByID", ctx, "d-good").Return(goodDoc, nil)
	mockDocRepo.On("Update", ctx, mock.MatchedBy(func(d *models.ParticipantDocument) bool {
		return d.Name == "NewName"
	})).Return(nil)
	updated, err := svc.UpdateDocument(ctx, "p-1", "d-good", "NewName", nil, "", "")
	assert.NoError(t, err)
	assert.Equal(t, "NewName", updated.Name)

	// Delete: doc does not belong to participant
	err = svc.DeleteDocument(ctx, "p-1", "d-wrong")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not belong to participant")

	// Delete: db delete error
	mockDocRepo.On("Delete", ctx, "d-good").Return(errors.New("db delete error"))
	err = svc.DeleteDocument(ctx, "p-1", "d-good")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete document from database")
}

func TestParticipantDocument_DownloadDocument(t *testing.T) {
	mockPartRepo := new(MockParticipantRepository)
	mockDocRepo := new(MockParticipantDocRepo)
	mockStorage := new(MockObjectStorageService)
	svc := services.NewParticipantService(mockPartRepo, mockDocRepo, mockStorage, nil)

	ctx := context.Background()
	part := &models.Participant{ParticipantID: "p-1"}
	mockPartRepo.On("FindByID", ctx, "p-1").Return(part, nil)

	// Invalid doc URL (missing documents/ prefix)
	badURLDoc := &models.ParticipantDocument{ID: "d-bad", ParticipantID: "p-1", DocURL: "https://badurl/file.pdf"}
	mockDocRepo.On("GetByID", ctx, "d-bad").Return(badURLDoc, nil)
	_, _, _, err := svc.DownloadDocument(ctx, "p-1", "d-bad")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid document url")

	// Success
	goodDoc := &models.ParticipantDocument{
		ID:            "d-ok",
		ParticipantID: "p-1",
		DocURL:        "https://storage/documents/p-1/file.pdf?query=1",
		FileType:      "application/pdf",
		Name:          "DNI",
	}
	mockDocRepo.On("GetByID", ctx, "d-ok").Return(goodDoc, nil)
	rc := io.NopCloser(strings.NewReader("pdf-bytes"))
	mockStorage.On("DownloadFile", ctx, "documents/p-1/file.pdf").Return(rc, nil)

	reader, cType, name, err := svc.DownloadDocument(ctx, "p-1", "d-ok")
	assert.NoError(t, err)
	assert.Equal(t, "application/pdf", cType)
	assert.Equal(t, "DNI", name)
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(reader)
	assert.Equal(t, "pdf-bytes", buf.String())
}
