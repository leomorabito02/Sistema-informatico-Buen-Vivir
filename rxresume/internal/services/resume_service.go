package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"buenvivir/rxresume/internal/dto"
	"buenvivir/rxresume/internal/models"
	"buenvivir/rxresume/internal/renderer"
	"buenvivir/rxresume/internal/repositories"
	"buenvivir/rxresume/internal/storage"
)

type ResumeService interface {
	GetResumeByUserID(ctx context.Context, userID string) (*models.Resume, error)
	// GetPublicResume solo lee: si el CV no existe devuelve ErrResumeNotFound
	// (no lo crea, a diferencia de GetResumeByUserID).
	GetPublicResume(ctx context.Context, userID string) (*models.Resume, error)
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

const (
	picturesStoragePrefix = "pictures/"
	contentTypeJPEG       = "image/jpeg"
	contentTypePNG        = "image/png"
)

var ErrPDFNotReady = errors.New("pdf not ready")

// ErrResumeNotFound: el usuario no tiene CV. Las rutas públicas responden 404 (no
// crean un CV vacío).
var ErrResumeNotFound = errors.New("resume not found")

type resumeService struct {
	repo     repositories.ResumeRepository
	storage  storage.ObjectStorageService
	renderer renderer.Client
}

func NewResumeService(
	repo repositories.ResumeRepository,
	storage storage.ObjectStorageService,
	rendererClient renderer.Client,
) ResumeService {
	return &resumeService{
		repo:     repo,
		storage:  storage,
		renderer: rendererClient,
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

	// Always ensure standard accessible URLs
	resume.PDFURL = fmt.Sprintf("/api/v1/rxresume/public/%s/pdf", userID)
	resume.PublicURL = fmt.Sprintf("/cv-public/%s", userID)

	return resume, nil
}

// GetAllResumes returns all participants' resumes for staff.
func (s *resumeService) GetAllResumes(ctx context.Context) ([]models.Resume, error) {
	resumes, err := s.repo.FindAllParticipants(ctx)
	if err != nil {
		return nil, err
	}
	for i := range resumes {
		resumes[i].PDFURL = fmt.Sprintf("/api/v1/rxresume/public/%s/pdf", resumes[i].UserID)
		resumes[i].PublicURL = fmt.Sprintf("/cv-public/%s", resumes[i].UserID)
	}
	return resumes, nil
}

// UpdateResume handles full updates to a resume and invalidates the cached PDF.
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

	// Invalidate cached PDF
	_ = s.storage.DeleteFile(ctx, fmt.Sprintf("pdfs/%s.pdf", userID))

	return resume, nil
}

// GenerateResume updates resume with profile info and invalidates cached PDF.
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
	if profileData.Picture != "" {
		resume.Data.Basics.Picture = profileData.Picture
	}

	resume.UpdatedAt = time.Now()

	err = s.repo.Transaction(ctx, func(txRepo repositories.ResumeRepository) error {
		return txRepo.Save(ctx, resume)
	})
	if err != nil {
		return nil, err
	}

	// Invalidate cached PDF
	_ = s.storage.DeleteFile(ctx, fmt.Sprintf("pdfs/%s.pdf", userID))

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

	resume.PDFURL = gcsURL
	resume.UpdatedAt = time.Now()

	err = s.repo.Transaction(ctx, func(txRepo repositories.ResumeRepository) error {
		return txRepo.Save(ctx, resume)
	})
	if err != nil {
		return nil, err
	}

	_ = s.storage.DeleteFile(ctx, fmt.Sprintf("pdfs/%s.pdf", userID))
	return resume, nil
}

func determineImageContentType(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return contentTypeJPEG
	case ".png":
		return contentTypePNG
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

func extractStoragePath(rawURL string) string {
	if idx := strings.Index(rawURL, "path="); idx != -1 {
		raw := rawURL[idx+5:]
		if amp := strings.Index(raw, "&"); amp != -1 {
			raw = raw[:amp]
		}
		if unescaped, err := url.QueryUnescape(raw); err == nil && unescaped != "" {
			return unescaped
		}
		return raw
	}
	if u, err := url.Parse(rawURL); err == nil {
		if p := u.Query().Get("path"); p != "" {
			return p
		}
	}
	return rawURL
}

func (s *resumeService) deleteOldPicture(ctx context.Context, pictureURL string) {
	if pictureURL == "" {
		return
	}
	oldPath := extractStoragePath(pictureURL)
	if strings.HasPrefix(oldPath, picturesStoragePrefix) {
		_ = s.storage.DeleteFile(ctx, oldPath)
	}
}

// UploadPicture uploads the picture to storage and saves the URL in our DB.
func (s *resumeService) UploadPicture(ctx context.Context, userID string, filename string, reader io.Reader) (*models.Resume, error) {
	resume, err := s.GetResumeByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	cleanName := filepath.Base(filename)
	cleanName = strings.ReplaceAll(cleanName, " ", "_")
	ext := strings.ToLower(filepath.Ext(cleanName))
	contentType := determineImageContentType(ext)

	objectName := fmt.Sprintf("%s%s/%d_%s", picturesStoragePrefix, userID, time.Now().Unix(), cleanName)
	if _, err = s.storage.UploadFile(ctx, objectName, reader, contentType); err != nil {
		return nil, fmt.Errorf("failed to upload picture to storage: %w", err)
	}

	s.deleteOldPicture(ctx, resume.Data.Basics.Picture)

	proxyURL := fmt.Sprintf("/api/v1/rxresume/public/media?path=%s", url.QueryEscape(objectName))
	resume.Data.Basics.Picture = proxyURL
	resume.UpdatedAt = time.Now()

	err = s.repo.Transaction(ctx, func(txRepo repositories.ResumeRepository) error {
		return txRepo.Save(ctx, resume)
	})
	if err != nil {
		return nil, err
	}

	_ = s.storage.DeleteFile(ctx, fmt.Sprintf("pdfs/%s.pdf", userID))
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

// GetPDFStatus checks whether the cached PDF exists or can be rendered.
func (s *resumeService) GetPDFStatus(ctx context.Context, userID string) bool {
	rc, err := s.storage.DownloadFile(ctx, fmt.Sprintf("pdfs/%s.pdf", userID))
	if err == nil {
		_ = rc.Close()
		return true
	}
	// Resume exists and can be rendered on demand
	resume, err := s.repo.FindByUserID(ctx, userID)
	return err == nil && resume != nil
}

// GetPublicResume devuelve el CV sin crearlo si no existe.
func (s *resumeService) GetPublicResume(ctx context.Context, userID string) (*models.Resume, error) {
	resume, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if resume == nil {
		return nil, ErrResumeNotFound
	}
	resume.PDFURL = fmt.Sprintf("/api/v1/rxresume/public/%s/pdf", userID)
	resume.PublicURL = fmt.Sprintf("/cv-public/%s", userID)
	return resume, nil
}

// ExportPDF returns the generated PDF, using GCS cache or rendering on-demand via cv-renderer.
func (s *resumeService) ExportPDF(ctx context.Context, userID string) ([]byte, error) {
	objectName := fmt.Sprintf("pdfs/%s.pdf", userID)

	if rc, err := s.storage.DownloadFile(ctx, objectName); err == nil {
		defer rc.Close()
		if pdfBytes, err := io.ReadAll(rc); err == nil && len(pdfBytes) > 0 {
			return pdfBytes, nil
		}
	}

	// 2. Fetch resume data (ruta pública: solo lectura, no crea el CV)
	resume, err := s.GetPublicResume(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Prepare resume for renderer: embed local picture as data URL if available
	resumeForRender := *resume
	if resumeForRender.Data.Basics.Picture != "" {
		if dataURL := s.resolvePictureDataURL(ctx, resumeForRender.Data.Basics.Picture); dataURL != "" {
			resumeForRender.Data.Basics.Picture = dataURL
		}
	}

	pdfBytes, err := s.renderer.RenderPDF(ctx, &resumeForRender)
	if err != nil {
		return nil, fmt.Errorf("failed to render PDF: %w", err)
	}

	// 4. Cache in storage asynchronously/best-effort
	_, _ = s.storage.UploadFile(ctx, objectName, bytes.NewReader(pdfBytes), "application/pdf")

	return pdfBytes, nil
}

func parseStoragePath(picURL string) string {
	storagePath := extractStoragePath(picURL)
	if idx := strings.Index(storagePath, picturesStoragePrefix); idx != -1 {
		storagePath = storagePath[idx:]
	}
	if unescaped, err := url.QueryUnescape(storagePath); err == nil && unescaped != "" {
		return unescaped
	}
	return storagePath
}

func (s *resumeService) resolveStoragePictureDataURL(ctx context.Context, path string) string {
	rc, err := s.storage.DownloadFile(ctx, path)
	if err != nil {
		return ""
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil || len(data) == 0 {
		return ""
	}

	mimeType := contentTypeJPEG
	if strings.ToLower(filepath.Ext(path)) == ".png" {
		mimeType = contentTypePNG
	}
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
}

func fetchRemotePictureDataURL(ctx context.Context, rawURL string) string {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return ""
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		return ""
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil || len(data) == 0 {
		return ""
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = contentTypeJPEG
	}
	if idx := strings.Index(ct, ";"); idx != -1 {
		ct = ct[:idx]
	}
	return fmt.Sprintf("data:%s;base64,%s", ct, base64.StdEncoding.EncodeToString(data))
}

func (s *resumeService) resolvePictureDataURL(ctx context.Context, picURL string) string {
	if strings.HasPrefix(picURL, "data:image/") {
		return picURL
	}

	storagePath := parseStoragePath(picURL)
	if strings.HasPrefix(storagePath, picturesStoragePrefix) {
		if dataURL := s.resolveStoragePictureDataURL(ctx, storagePath); dataURL != "" {
			return dataURL
		}
	}

	if strings.HasPrefix(picURL, "http://") || strings.HasPrefix(picURL, "https://") {
		if dataURL := fetchRemotePictureDataURL(ctx, picURL); dataURL != "" {
			return dataURL
		}
	}

	return picURL
}
