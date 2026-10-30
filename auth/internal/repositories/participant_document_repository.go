package repositories

import (
	"context"

	"auth/internal/models"

	"gorm.io/gorm"
)

type ParticipantDocumentRepository interface {
	Create(ctx context.Context, doc *models.ParticipantDocument) error
	GetByID(ctx context.Context, id string) (*models.ParticipantDocument, error)
	GetByParticipantID(ctx context.Context, participantID string) ([]models.ParticipantDocument, error)
	Update(ctx context.Context, doc *models.ParticipantDocument) error
	Delete(ctx context.Context, id string) error
}

type participantDocumentRepository struct {
	db *gorm.DB
}

func NewParticipantDocumentRepository(db *gorm.DB) ParticipantDocumentRepository {
	return &participantDocumentRepository{db: db}
}

func (r *participantDocumentRepository) Create(ctx context.Context, doc *models.ParticipantDocument) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

func (r *participantDocumentRepository) GetByID(ctx context.Context, id string) (*models.ParticipantDocument, error) {
	var doc models.ParticipantDocument
	err := r.db.WithContext(ctx).First(&doc, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

func (r *participantDocumentRepository) GetByParticipantID(ctx context.Context, participantID string) ([]models.ParticipantDocument, error) {
	var docs []models.ParticipantDocument
	err := r.db.WithContext(ctx).Where("participant_id = ?", participantID).Find(&docs).Error
	return docs, err
}

func (r *participantDocumentRepository) Update(ctx context.Context, doc *models.ParticipantDocument) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

func (r *participantDocumentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&models.ParticipantDocument{}, "id = ?", id).Error
}
