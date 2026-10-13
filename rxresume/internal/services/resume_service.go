package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"buenvivir/rxresume/internal/client"
	"buenvivir/rxresume/internal/dto"
	"buenvivir/rxresume/internal/models"
	"buenvivir/rxresume/internal/parser"
	"buenvivir/rxresume/internal/repositories"
	"buenvivir/rxresume/internal/storage"
	"buenvivir/rxresume/internal/worker"
)

type ResumeService interface {
	GetResumeByUserID(ctx context.Context, userID string) (*models.Resume, error)
	GetAllResumes(ctx context.Context) ([]models.Resume, error)
	UpdateResume(ctx context.Context, userID string, data *models.ResumeData) (*models.Resume, error)
	ParseAndUploadPDF(ctx context.Context, userID string, filename string, reader io.Reader) (*models.Resume, error)
	DeleteSectionItem(ctx context.Context, userID string, section string, itemID string) (*models.Resume, error)
	GenerateResume(ctx context.Context, userID string, profileData *dto.GenerateResumeRequest) (*models.Resume, error)
	UploadPicture(ctx context.Context, userID string, filename string, reader io.Reader) (*models.Resume, error)
	ExportPDF(ctx context.Context, userID string) ([]byte, error)
	DownloadMedia(ctx context.Context, path string) (io.ReadCloser, error)
	GetPDFStatus(ctx context.Context, userID string) bool
}

var ErrPDFNotReady = errors.New("pdf not ready")

type resumeService struct {
	repo      repositories.ResumeRepository
	storage   storage.ObjectStorageService
	parser    parser.PDFParser
	rxClient  client.ReactiveResumeClient
	syncQueue worker.SyncQueue
}

func NewResumeService(
	repo repositories.ResumeRepository,
	storage storage.ObjectStorageService,
	parser parser.PDFParser,
	rxClient client.ReactiveResumeClient,
	syncQueue worker.SyncQueue,
) ResumeService {
	return &resumeService{
		repo:      repo,
		storage:   storage,
		parser:    parser,
		rxClient:  rxClient,
		syncQueue: syncQueue,
	}
}

// GetResumeByUserID retrieves or initializes a participant's resume.
func (s *resumeService) GetResumeByUserID(ctx context.Context, userID string) (*models.Resume, error) {
	resume, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if resume == nil {
		resume = &models.Resume{
			ID:        "rsm-" + userID,
			UserID:    userID,
			Title:     "Curriculum Vitae",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Data: models.ResumeData{
				Basics: models.Basics{
					Name:     "",
					Email:    "",
					Phone:    "",
					Headline: "",
					Location: "",
					Summary:  "",
				},
				Work:           []models.WorkExperience{},
				Education:      []models.EducationItem{},
				Skills:         []models.Skill{},
				Certifications: []models.Certification{},
				Languages:      []models.LanguageItem{},
				Awards:         []models.Award{},
				Volunteer:      []models.VolunteerExperience{},
				Projects:       []models.Project{},
				References:     []models.Reference{},
				Interests:      []models.Interest{},
			},
		}
		if err := s.repo.Save(ctx, resume); err != nil {
			return nil, err
		}
	}
	return resume, nil
}

// GetAllResumes returns all participants' resumes for staff.
func (s *resumeService) GetAllResumes(ctx context.Context) ([]models.Resume, error) {
	return s.repo.FindAllParticipants(ctx)
}

// UpdateResume handles full updates to a resume, syncing with self-hosted Reactive Resume instance.
func (s *resumeService) UpdateResume(ctx context.Context, userID string, data *models.ResumeData) (*models.Resume, error) {
	resume, err := s.GetResumeByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	resume.Data = *data
	resume.UpdatedAt = time.Now()

	err = s.repo.Transaction(ctx, func(txRepo repositories.ResumeRepository) error {
		return txRepo.Save(ctx, resume)
	})
	if err != nil {
		return nil, err
	}
	_ = s.storage.DeleteFile(ctx, fmt.Sprintf("pdfs/%s.pdf", userID))
	s.syncQueue.Enqueue(userID)
	return resume, nil
}

// GenerateResume synchronizes the resume with Reactive Resume explicitly and updates profile info.
func (s *resumeService) GenerateResume(ctx context.Context, userID string, profileData *dto.GenerateResumeRequest) (*models.Resume, error) {
	resume, err := s.GetResumeByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if profileData.Name != "" {
		resume.Data.Basics.Name = profileData.Name
	}
	if profileData.Email != "" {
		resume.Data.Basics.Email = profileData.Email
	}
	if profileData.Phone != "" {
		resume.Data.Basics.Phone = profileData.Phone
	}
	if profileData.Location != "" {
		resume.Data.Basics.Location = profileData.Location
	}
	if profileData.Headline != "" {
		resume.Data.Basics.Headline = profileData.Headline
	}

	resume.UpdatedAt = time.Now()

	err = s.repo.Transaction(ctx, func(txRepo repositories.ResumeRepository) error {
		return txRepo.Save(ctx, resume)
	})
	if err != nil {
		return nil, err
	}
	_ = s.storage.DeleteFile(ctx, fmt.Sprintf("pdfs/%s.pdf", userID))
	s.syncQueue.Enqueue(userID)
	return resume, nil
}

// ParseAndUploadPDF processes an uploaded PDF and uploads it to object storage.
func (s *resumeService) ParseAndUploadPDF(ctx context.Context, userID string, filename string, reader io.Reader) (*models.Resume, error) {
	resume, err := s.GetResumeByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	cleanName := filepath.Base(filename)
	objectName := fmt.Sprintf("resumes/%s/%d_%s", userID, time.Now().Unix(), cleanName)
	gcsURL, err := s.storage.UploadFile(ctx, objectName, reader, "application/pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to upload pdf to storage: %w", err)
	}

	parsedData, err := s.parser.ParsePDFToResume(reader, cleanName)
	if err == nil && parsedData != nil {
		resume.Data = *parsedData
	}

	resume.PDFURL = gcsURL
	resume.UpdatedAt = time.Now()

	err = s.repo.Transaction(ctx, func(txRepo repositories.ResumeRepository) error {
		return txRepo.Save(ctx, resume)
	})
	if err != nil {
		return nil, err
	}
	_ = s.storage.DeleteFile(ctx, fmt.Sprintf("pdfs/%s.pdf", userID))
	s.syncQueue.Enqueue(userID)

	return resume, nil
}

// UploadPicture uploads the picture to Reactive Resume and saves the URL in our DB.
func (s *resumeService) UploadPicture(ctx context.Context, userID string, filename string, reader io.Reader) (*models.Resume, error) {
	resume, err := s.GetResumeByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	cleanName := filepath.Base(filename)
	ext := strings.ToLower(filepath.Ext(cleanName))
	contentType := "application/octet-stream"
	switch ext {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	objectName := fmt.Sprintf("pictures/%s/%d_%s", userID, time.Now().Unix(), cleanName)
	_, err = s.storage.UploadFile(ctx, objectName, reader, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload picture to storage: %w", err)
	}

	publicAPIURL := os.Getenv("PUBLIC_API_URL")
	if publicAPIURL == "" {
		publicAPIURL = "http://localhost:8080"
	}

	// Delete old picture if it exists and is hosted in our bucket
	if oldPictureURL := resume.Data.Basics.Picture; oldPictureURL != "" {
		if u, err := url.Parse(oldPictureURL); err == nil {
			if oldPath := u.Query().Get("path"); strings.HasPrefix(oldPath, "pictures/") {
				_ = s.storage.DeleteFile(ctx, oldPath)
			}
		}
	}

	proxyURL := fmt.Sprintf("/api/v1/rxresume/public/media?path=%s", objectName)
	resume.Data.Basics.Picture = proxyURL
	resume.UpdatedAt = time.Now()

	err = s.repo.Transaction(ctx, func(txRepo repositories.ResumeRepository) error {
		return txRepo.Save(ctx, resume)
	})
	if err != nil {
		return nil, err
	}
	_ = s.storage.DeleteFile(ctx, fmt.Sprintf("pdfs/%s.pdf", userID))
	s.syncQueue.Enqueue(userID)

	return resume, nil
}

// DownloadMedia streams a media file from GCP storage.
func (s *resumeService) DownloadMedia(ctx context.Context, path string) (io.ReadCloser, error) {
	return s.storage.DownloadFile(ctx, path)
}

// DeleteSectionItem removes an item from a section of a resume.
func (s *resumeService) DeleteSectionItem(ctx context.Context, userID string, section string, itemID string) (*models.Resume, error) {
	resume, err := s.GetResumeByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	deleteItemFromSection(resume, section, itemID)

	resume.UpdatedAt = time.Now()

	err = s.repo.Transaction(ctx, func(txRepo repositories.ResumeRepository) error {
		return txRepo.Save(ctx, resume)
	})
	if err != nil {
		return nil, err
	}
	_ = s.storage.DeleteFile(ctx, fmt.Sprintf("pdfs/%s.pdf", userID))
	s.syncQueue.Enqueue(userID)

	return resume, nil
}

func deleteItemFromSection(resume *models.Resume, section, itemID string) {
	switch section {
	case "work":
		deleteWorkItem(resume, itemID)
	case "education":
		deleteEducationItem(resume, itemID)
	case "skills":
		deleteSkillItem(resume, itemID)
	}
}

func deleteWorkItem(resume *models.Resume, itemID string) {
	var updated []models.WorkExperience
	for _, w := range resume.Data.Work {
		if w.ID != itemID {
			updated = append(updated, w)
		}
	}
	resume.Data.Work = updated
}

func deleteEducationItem(resume *models.Resume, itemID string) {
	var updated []models.EducationItem
	for _, e := range resume.Data.Education {
		if e.ID != itemID {
			updated = append(updated, e)
		}
	}
	resume.Data.Education = updated
}

func deleteSkillItem(resume *models.Resume, itemID string) {
	var updated []models.Skill
	for _, sk := range resume.Data.Skills {
		if sk.ID != itemID {
			updated = append(updated, sk)
		}
	}
	resume.Data.Skills = updated
}

// GetPDFStatus checks whether the cached PDF for a user exists in GCS.
func (s *resumeService) GetPDFStatus(ctx context.Context, userID string) bool {
	rc, err := s.storage.DownloadFile(ctx, fmt.Sprintf("pdfs/%s.pdf", userID))
	if err != nil {
		return false
	}
	rc.Close()
	return true
}

// ExportPDF downloads the generated PDF from Reactive Resume for a specific user, with GCS caching.
func (s *resumeService) ExportPDF(ctx context.Context, userID string) ([]byte, error) {
	resume, err := s.GetResumeByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	objectName := fmt.Sprintf("pdfs/%s.pdf", userID)

	// 1. Try to serve cached PDF from GCS
	if rc, err := s.storage.DownloadFile(ctx, objectName); err == nil {
		defer rc.Close()
		if pdfBytes, err := io.ReadAll(rc); err == nil && len(pdfBytes) > 0 {
			return pdfBytes, nil
		}
	}

	// 2. Fallback check - if we're here, the worker hasn't finished or failed
	// We won't block waiting for it, we just return an error so the FE can poll
	if resume.RxResumeID != "" {
		return nil, ErrPDFNotReady
	}
	
	// Edge case: first time, but normally enqueue should handle this
	return nil, ErrPDFNotReady
}
