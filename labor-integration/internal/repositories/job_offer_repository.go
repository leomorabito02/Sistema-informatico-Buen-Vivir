package repositories

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"buenvivir/labor-integration/internal/models"
)

type JobOfferRepository interface {
	Create(ctx context.Context, offer *models.JobOffer) error
	List(ctx context.Context, companyID string) ([]*models.JobOffer, error)
	Update(ctx context.Context, id string, offer *models.JobOffer) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status string) error
	UpdateStatusAndApprover(ctx context.Context, id string, status string, approvedBy string) error
	GetByID(ctx context.Context, id string) (*models.JobOffer, error)
}

type jobOfferRepository struct {
	client *firestore.Client
}

func NewJobOfferRepository(client *firestore.Client) JobOfferRepository {
	return &jobOfferRepository{client: client}
}

func (r *jobOfferRepository) Create(ctx context.Context, offer *models.JobOffer) error {
	ref := r.client.Collection("job_offers").NewDoc()
	offer.ID = ref.ID
	_, err := ref.Set(ctx, offer)
	return err
}

func (r *jobOfferRepository) List(ctx context.Context, companyID string) ([]*models.JobOffer, error) {
	col := r.client.Collection("job_offers")
	var iter *firestore.DocumentIterator
	if companyID != "" {
		iter = col.Where("empresa_id", "==", companyID).Documents(ctx)
	} else {
		iter = col.Documents(ctx)
	}
	var offers []*models.JobOffer
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var offer models.JobOffer
		if err := doc.DataTo(&offer); err != nil {
			continue // skip invalid docs
		}
		offers = append(offers, &offer)
	}
	return offers, nil
}

func (r *jobOfferRepository) Update(ctx context.Context, id string, offer *models.JobOffer) error {
	_, err := r.client.Collection("job_offers").Doc(id).Set(ctx, offer)
	return err
}

func (r *jobOfferRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Collection("job_offers").Doc(id).Delete(ctx)
	return err
}

func (r *jobOfferRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.client.Collection("job_offers").Doc(id).Update(ctx, []firestore.Update{
		{Path: "estado", Value: status},
	})
	return err
}

func (r *jobOfferRepository) UpdateStatusAndApprover(ctx context.Context, id string, status string, approvedBy string) error {
	_, err := r.client.Collection("job_offers").Doc(id).Update(ctx, []firestore.Update{
		{Path: "estado", Value: status},
		{Path: "aprobado_por", Value: approvedBy},
	})
	return err
}

func (r *jobOfferRepository) GetByID(ctx context.Context, id string) (*models.JobOffer, error) {
	doc, err := r.client.Collection("job_offers").Doc(id).Get(ctx)
	if err != nil {
		return nil, err
	}
	var offer models.JobOffer
	if err := doc.DataTo(&offer); err != nil {
		return nil, err
	}
	offer.ID = doc.Ref.ID
	return &offer, nil
}

