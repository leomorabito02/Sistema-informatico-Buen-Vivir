package repositories

import (
	"context"
	"errors"

	"auth/internal/models"
	"gorm.io/gorm"
)

type TeacherRepository interface {
	FindByID(ctx context.Context, id string) (*models.Teacher, error)
	FindAll(ctx context.Context) ([]models.Teacher, error)
	Create(ctx context.Context, teacher *models.Teacher) error
	Update(ctx context.Context, teacher *models.Teacher) error
	Delete(ctx context.Context, id string) error
}

type teacherRepository struct {
	db *gorm.DB
}

func NewTeacherRepository(db *gorm.DB) TeacherRepository {
	return &teacherRepository{db: db}
}

func (r *teacherRepository) FindByID(ctx context.Context, id string) (*models.Teacher, error) {
	var teacher models.Teacher
	if err := r.db.WithContext(ctx).First(&teacher, "teacher_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profesor no encontrado")
		}
		return nil, err
	}
	return &teacher, nil
}

func (r *teacherRepository) FindAll(ctx context.Context) ([]models.Teacher, error) {
	var teachers []models.Teacher
	if err := r.db.WithContext(ctx).Find(&teachers).Error; err != nil {
		return nil, err
	}
	return teachers, nil
}

func (r *teacherRepository) Create(ctx context.Context, teacher *models.Teacher) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Create(teacher).Error
	})
}

func (r *teacherRepository) Update(ctx context.Context, teacher *models.Teacher) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Save(teacher).Error
	})
}

func (r *teacherRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Where("teacher_id = ?", id).Delete(&models.Teacher{}).Error
	})
}
