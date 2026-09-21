package repositories

import (
	"context"
	"errors"

	"buenvivir/rxresume/internal/models"
	"gorm.io/gorm"
)

type ResumeRepository interface {
	FindByUserID(ctx context.Context, userID string) (*models.Resume, error)
	FindAllParticipants(ctx context.Context) ([]models.Resume, error)
	Save(ctx context.Context, resume *models.Resume) error
	Delete(ctx context.Context, id string) error
}

type resumeRepository struct {
	db *gorm.DB
}

func NewResumeRepository(db *gorm.DB) ResumeRepository {
	return &resumeRepository{db: db}
}

// FindByUserID retrieves the resume belonging to the specified user_id.
func (r *resumeRepository) FindByUserID(ctx context.Context, userID string) (*models.Resume, error) {
	var resume models.Resume
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&resume).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &resume, nil
}

// FindAllParticipants returns all stored resumes in the system.
func (r *resumeRepository) FindAllParticipants(ctx context.Context) ([]models.Resume, error) {
	var resumes []models.Resume
	err := r.db.WithContext(ctx).Order("updated_at desc").Find(&resumes).Error
	return resumes, err
}

// Save creates or updates a resume entry in the database.
func (r *resumeRepository) Save(ctx context.Context, resume *models.Resume) error {
	return r.db.WithContext(ctx).Save(resume).Error
}

// Delete removes a resume entry by ID.
func (r *resumeRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Resume{}).Error
}
