package repositories

import (
	"context"
	"errors"

	"auth/internal/models"
	"gorm.io/gorm"
)

type VolunteerRepository interface {
	FindByID(ctx context.Context, id string) (*models.Volunteer, error)
	FindAll(ctx context.Context) ([]models.Volunteer, error)
	Create(ctx context.Context, volunteer *models.Volunteer) error
	Update(ctx context.Context, volunteer *models.Volunteer) error
	Delete(ctx context.Context, id string) error
}

type volunteerRepository struct {
	db *gorm.DB
}

func NewVolunteerRepository(db *gorm.DB) VolunteerRepository {
	return &volunteerRepository{db: db}
}

func (r *volunteerRepository) FindByID(ctx context.Context, id string) (*models.Volunteer, error) {
	var volunteer models.Volunteer
	if err := r.db.WithContext(ctx).First(&volunteer, "volunteer_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("voluntario no encontrado")
		}
		return nil, err
	}
	return &volunteer, nil
}

func (r *volunteerRepository) FindAll(ctx context.Context) ([]models.Volunteer, error) {
	var volunteers []models.Volunteer
	if err := r.db.WithContext(ctx).Find(&volunteers).Error; err != nil {
		return nil, err
	}
	return volunteers, nil
}

func (r *volunteerRepository) Create(ctx context.Context, volunteer *models.Volunteer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Create(volunteer).Error
	})
}

func (r *volunteerRepository) Update(ctx context.Context, volunteer *models.Volunteer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Save(volunteer).Error
	})
}

func (r *volunteerRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Where("volunteer_id = ?", id).Delete(&models.Volunteer{}).Error
	})
}
