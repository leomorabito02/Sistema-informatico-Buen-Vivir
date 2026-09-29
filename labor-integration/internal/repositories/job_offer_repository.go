package repositories

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"buenvivir/labor-integration/internal/models"
)

type JobOfferRepository interface {
	Create(ctx context.Context, offer *models.JobOffer) error
	List(ctx context.Context) ([]*models.JobOffer, error)
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

func (r *jobOfferRepository) List(ctx context.Context) ([]*models.JobOffer, error) {
	iter := r.client.Collection("job_offers").Documents(ctx)
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
