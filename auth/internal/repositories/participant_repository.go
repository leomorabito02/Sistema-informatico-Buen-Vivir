package repositories

import (
	"context"
	"errors"

	"auth/internal/models"
	"gorm.io/gorm"
)

type ParticipantRepository interface {
	FindByID(ctx context.Context, id string) (*models.Participant, error)
	FindAll(ctx context.Context) ([]models.Participant, error)
	Create(ctx context.Context, participant *models.Participant) error
	Update(ctx context.Context, participant *models.Participant) error
	Delete(ctx context.Context, id string) error
}

type participantRepository struct {
	db *gorm.DB
}

func NewParticipantRepository(db *gorm.DB) ParticipantRepository {
	return &participantRepository{db: db}
}

func (r *participantRepository) FindByID(ctx context.Context, id string) (*models.Participant, error) {
	var participant models.Participant
	if err := r.db.WithContext(ctx).First(&participant, "participant_id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("participante no encontrado")
		}
		return nil, err
	}
	return &participant, nil
}

func (r *participantRepository) FindAll(ctx context.Context) ([]models.Participant, error) {
	var participants []models.Participant
	if err := r.db.WithContext(ctx).Find(&participants).Error; err != nil {
		return nil, err
	}
	return participants, nil
}

func (r *participantRepository) Create(ctx context.Context, participant *models.Participant) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Create(participant).Error
	})
}

func (r *participantRepository) Update(ctx context.Context, participant *models.Participant) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Save(participant).Error
	})
}

func (r *participantRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Where("participant_id = ?", id).Delete(&models.Participant{}).Error
	})
}
