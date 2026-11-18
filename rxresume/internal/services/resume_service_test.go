package services

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"buenvivir/rxresume/internal/dto"
	"buenvivir/rxresume/internal/models"
	"buenvivir/rxresume/internal/repositories"
)

type mockRepo struct {
	resumes map[string]*models.Resume
}

func (m *mockRepo) FindByUserID(ctx context.Context, userID string) (*models.Resume, error) {
	return m.resumes[userID], nil
}

func (m *mockRepo) FindAllParticipants(ctx context.Context) ([]models.Resume, error) {
	var list []models.Resume
	for _, r := range m.resumes {
		list = append(list, *r)
	}
	return list, nil
}

func (m *mockRepo) Save(ctx context.Context, resume *models.Resume) error {
	m.resumes[resume.UserID] = resume
	return nil
}

func (m *mockRepo) Delete(ctx context.Context, userID string) error {
	delete(m.resumes, userID)
	return nil
}

func (m *mockRepo) Transaction(ctx context.Context, fn func(txRepo repositories.ResumeRepository) error) error {
	return fn(m)
}

type mockStorage struct {
	files map[string][]byte
}

func (s *mockStorage) UploadFile(ctx context.Context, objectName string, reader io.Reader, contentType string) (string, error) {
	data, _ := io.ReadAll(reader)
	s.files[objectName] = data
	return "https://storage.googleapis.com/" + objectName, nil
}

func (s *mockStorage) DownloadFile(ctx context.Context, objectName string) (io.ReadCloser, error) {
	data, exists := s.files[objectName]
	if !exists {
		return nil, io.EOF
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *mockStorage) DeleteFile(ctx context.Context, objectName string) error {
	delete(s.files, objectName)
	return nil
}

type mockRenderer struct {
	renderCount int
	output      []byte
	err         error
	lastResume  *models.Resume
}

func (r *mockRenderer) RenderPDF(ctx context.Context, resume *models.Resume) ([]byte, error) {
	r.renderCount++
	r.lastResume = resume
	return r.output, r.err
}

func TestResumeService_ExportPDF_Cached(t *testing.T) {
	// Arrange
	ctx := context.Background()
	cachedPDF := []byte("%PDF-cached")
	storage := &mockStorage{
		files: map[string][]byte{
			"pdfs/usr-1.pdf": cachedPDF,
		},
	}
	repo := &mockRepo{resumes: make(map[string]*models.Resume)}
	renderer := &mockRenderer{output: []byte("%PDF-new")}
	service := NewResumeService(repo, storage, renderer)

	// Act
	pdf, err := service.ExportPDF(ctx, "usr-1")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(pdf) != string(cachedPDF) {
		t.Errorf("expected cached PDF, got: %s", string(pdf))
	}
	if renderer.renderCount != 0 {
		t.Errorf("expected renderer not to be called, called %d times", renderer.renderCount)
	}
}

func TestResumeService_ExportPDF_OnDemandRenderAndCache(t *testing.T) {
	// Arrange
	ctx := context.Background()
	freshPDF := []byte("%PDF-fresh-render")
	storage := &mockStorage{files: make(map[string][]byte)}
	repo := &mockRepo{
		resumes: map[string]*models.Resume{
			"usr-2": {
				ID:     "rsm-2",
				UserID: "usr-2",
				Data:   models.ResumeData{Basics: models.Basics{Name: "Ana"}},
			},
		},
	}
	renderer := &mockRenderer{output: freshPDF}
	service := NewResumeService(repo, storage, renderer)

	// Act
	pdf, err := service.ExportPDF(ctx, "usr-2")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(pdf) != string(freshPDF) {
		t.Errorf("expected fresh PDF, got %s", string(pdf))
	}
	if renderer.renderCount != 1 {
		t.Errorf("expected renderer to be called once, called %d times", renderer.renderCount)
	}
	// Verify it was cached in storage
	cached, exists := storage.files["pdfs/usr-2.pdf"]
	if !exists || string(cached) != string(freshPDF) {
		t.Errorf("expected PDF to be cached in storage")
	}
}

func TestResumeService_UpdateResume_InvalidatesCache(t *testing.T) {
	// Arrange
	ctx := context.Background()
	storage := &mockStorage{
		files: map[string][]byte{
			"pdfs/usr-3.pdf": []byte("%PDF-old"),
		},
	}
	repo := &mockRepo{
		resumes: map[string]*models.Resume{
			"usr-3": {ID: "rsm-3", UserID: "usr-3"},
		},
	}
	renderer := &mockRenderer{}
	service := NewResumeService(repo, storage, renderer)

	// Act
	updated, err := service.UpdateResume(ctx, "usr-3", &models.ResumeData{
		Basics: models.Basics{Name: "Carlos Nuevo"},
	})

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Data.Basics.Name != "Carlos Nuevo" {
		t.Errorf("expected name Carlos Nuevo, got %s", updated.Data.Basics.Name)
	}
	// Verify cache was deleted
	if _, exists := storage.files["pdfs/usr-3.pdf"]; exists {
		t.Errorf("expected cached pdfs/usr-3.pdf to be deleted on update")
	}
}

func TestResumeService_ExportPDF_ResolvesPictureToDataURL(t *testing.T) {
	// Arrange
	ctx := context.Background()
	storage := &mockStorage{
		files: map[string][]byte{
			"pictures/usr-4/avatar.png": []byte("png-file-content"),
		},
	}
	repo := &mockRepo{
		resumes: map[string]*models.Resume{
			"usr-4": {
				ID:     "rsm-4",
				UserID: "usr-4",
				Data: models.ResumeData{
					Basics: models.Basics{
						Name:    "Maria Test",
						Picture: "/api/v1/rxresume/public/media?path=pictures/usr-4/avatar.png",
					},
				},
			},
		},
	}
	renderer := &mockRenderer{output: []byte("%PDF-rendered")}
	service := NewResumeService(repo, storage, renderer)

	// Act
	pdf, err := service.ExportPDF(ctx, "usr-4")

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(pdf) != "%PDF-rendered" {
		t.Errorf("unexpected pdf output: %s", string(pdf))
	}
	if renderer.lastResume == nil {
		t.Fatalf("expected renderer to receive resume, got nil")
	}
	expectedPrefix := "data:image/png;base64,"
	if !strings.HasPrefix(renderer.lastResume.Data.Basics.Picture, expectedPrefix) {
		t.Errorf("expected picture to be data URL with prefix %s, got %s", expectedPrefix, renderer.lastResume.Data.Basics.Picture)
	}
	// Verify original repository resume wasn't mutated
	if repo.resumes["usr-4"].Data.Basics.Picture != "/api/v1/rxresume/public/media?path=pictures/usr-4/avatar.png" {
		t.Errorf("expected repo resume picture to remain unchanged, got %s", repo.resumes["usr-4"].Data.Basics.Picture)
	}
}

func TestResumeService_GenerateResume_SetsPicture(t *testing.T) {
	// Arrange
	ctx := context.Background()
	storage := &mockStorage{files: make(map[string][]byte)}
	repo := &mockRepo{
		resumes: map[string]*models.Resume{
			"usr-5": {
				ID:     "rsm-5",
				UserID: "usr-5",
				Data: models.ResumeData{
					Basics: models.Basics{
						Name:    "Old Name",
						Picture: "",
					},
				},
			},
		},
	}
	service := NewResumeService(repo, storage, nil)

	req := &dto.GenerateResumeRequest{
		Name:    "New Name",
		Email:   "new@example.com",
		Picture: "https://example.com/avatar.jpg",
	}

	// Act
	res, err := service.GenerateResume(ctx, "usr-5", req)

	// Assert
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Data.Basics.Picture != "https://example.com/avatar.jpg" {
		t.Errorf("expected picture to be set, got %s", res.Data.Basics.Picture)
	}
	if res.Data.Basics.Name != "New Name" {
		t.Errorf("expected name to be 'New Name', got %s", res.Data.Basics.Name)
	}
}

// Las rutas públicas no crean CVs: un id inexistente da ErrResumeNotFound
// y la base queda igual (escaneo ZAP del 17/11/2026).
func TestResumeService_PublicoNoCreaCV(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{resumes: make(map[string]*models.Resume)}
	renderer := &mockRenderer{output: []byte("%PDF")}
	service := NewResumeService(repo, &mockStorage{files: map[string][]byte{}}, renderer)

	if _, err := service.GetPublicResume(ctx, "id-inventado"); !errors.Is(err, ErrResumeNotFound) {
		t.Fatalf("esperaba ErrResumeNotFound, fue %v", err)
	}
	if _, err := service.ExportPDF(ctx, "id-inventado"); !errors.Is(err, ErrResumeNotFound) {
		t.Fatalf("PDF: esperaba ErrResumeNotFound, fue %v", err)
	}
	if len(repo.resumes) != 0 || renderer.renderCount != 0 {
		t.Fatalf("no debía crear CVs ni renderizar: %d CVs, %d renders", len(repo.resumes), renderer.renderCount)
	}

	repo.resumes["usr-9"] = &models.Resume{UserID: "usr-9"}
	r, err := service.GetPublicResume(ctx, "usr-9")
	if err != nil || r.PublicURL != "/cv-public/usr-9" {
		t.Fatalf("CV existente: %+v, %v", r, err)
	}
}
