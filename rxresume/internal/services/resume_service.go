package services

import (
	"context"
	"fmt"
	"io"
	"time"

	"buenvivir/rxresume/internal/client"
	"buenvivir/rxresume/internal/models"
	"buenvivir/rxresume/internal/parser"
	"buenvivir/rxresume/internal/repositories"
	"buenvivir/rxresume/internal/storage"
)

type ResumeService interface {
	GetResumeByUserID(ctx context.Context, userID string) (*models.Resume, error)
	GetAllResumes(ctx context.Context) ([]models.Resume, error)
	UpdateResume(ctx context.Context, userID string, data *models.ResumeData) (*models.Resume, error)
	ParseAndUploadPDF(ctx context.Context, userID string, filename string, reader io.Reader) (*models.Resume, error)
	DeleteSectionItem(ctx context.Context, userID string, section string, itemID string) (*models.Resume, error)
	GenerateResume(ctx context.Context, userID string, profileData *models.ProfileInput) (*models.Resume, error)
	UploadPicture(ctx context.Context, userID string, filename string, reader io.Reader) (*models.Resume, error)
	ExportPDF(ctx context.Context, userID string) ([]byte, error)
}

type resumeService struct {
	repo     repositories.ResumeRepository
	storage  storage.ObjectStorageService
	parser   parser.PDFParser
	rxClient client.ReactiveResumeClient
}

func NewResumeService(
	repo repositories.ResumeRepository,
	storage storage.ObjectStorageService,
	parser parser.PDFParser,
	rxClient client.ReactiveResumeClient,
) ResumeService {
	return &resumeService{
		repo:     repo,
		storage:  storage,
		parser:   parser,
		rxClient: rxClient,
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

	// Sync with self-hosted Reactive Resume API instance
	err = s.rxClient.SyncResume(ctx, resume)
	if err != nil {
		return nil, fmt.Errorf("sync failed: %w", err)
	}

	if err := s.repo.Save(ctx, resume); err != nil {
		return nil, err
	}
	return resume, nil
}

// GenerateResume synchronizes the resume with Reactive Resume explicitly and updates profile info.
func (s *resumeService) GenerateResume(ctx context.Context, userID string, profileData *models.ProfileInput) (*models.Resume, error) {
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

	err = s.rxClient.SyncResume(ctx, resume)
	if err != nil {
		return nil, fmt.Errorf("sync failed: %w", err)
	}

	if err := s.repo.Save(ctx, resume); err != nil {
		return nil, err
	}
	return resume, nil
}

// ParseAndUploadPDF processes an uploaded PDF and uploads it to object storage.
func (s *resumeService) ParseAndUploadPDF(ctx context.Context, userID string, filename string, reader io.Reader) (*models.Resume, error) {
	resume, err := s.GetResumeByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	objectName := fmt.Sprintf("resumes/%s/%d_%s", userID, time.Now().Unix(), filename)
	gcsURL, err := s.storage.UploadFile(ctx, objectName, reader, "application/pdf")
	if err != nil {
		return nil, fmt.Errorf("failed to upload pdf to storage: %w", err)
	}

	parsedData, err := s.parser.ParsePDFToResume(reader, filename)
	if err == nil && parsedData != nil {
		resume.Data = *parsedData
	}

	resume.PDFURL = gcsURL
	resume.UpdatedAt = time.Now()

	_ = s.rxClient.SyncResume(ctx, resume)

	if err := s.repo.Save(ctx, resume); err != nil {
		return nil, err
	}

	return resume, nil
}

// UploadPicture uploads the picture to Reactive Resume and saves the URL in our DB.
func (s *resumeService) UploadPicture(ctx context.Context, userID string, filename string, reader io.Reader) (*models.Resume, error) {
	resume, err := s.GetResumeByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	url, err := s.rxClient.UploadImage(ctx, userID, reader, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image to reactive resume: %w", err)
	}

	resume.Data.Basics.Picture = url
	resume.UpdatedAt = time.Now()

	_ = s.rxClient.SyncResume(ctx, resume)

	if err := s.repo.Save(ctx, resume); err != nil {
		return nil, err
	}

	return resume, nil
}

// DeleteSectionItem removes an item from a section of a resume.
func (s *resumeService) DeleteSectionItem(ctx context.Context, userID string, section string, itemID string) (*models.Resume, error) {
	resume, err := s.GetResumeByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	deleteItemFromSection(resume, section, itemID)

	resume.UpdatedAt = time.Now()
	
	_ = s.rxClient.SyncResume(ctx, resume)

	if err := s.repo.Save(ctx, resume); err != nil {
		return nil, err
	}

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

// ExportPDF downloads the generated PDF from Reactive Resume for a specific user.
func (s *resumeService) ExportPDF(ctx context.Context, userID string) ([]byte, error) {
	resume, err := s.GetResumeByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if resume.RxResumeID == "" {
		return nil, fmt.Errorf("resume not synchronized with Reactive Resume yet")
	}

	return s.rxClient.ExportResumePDF(ctx, userID)
}
