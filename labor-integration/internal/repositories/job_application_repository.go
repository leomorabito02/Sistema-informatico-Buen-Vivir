package repositories

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"

	"buenvivir/labor-integration/internal/models"
)

type JobApplicationRepository interface {
	Create(ctx context.Context, application *models.JobApplication) error
	ListByOffer(ctx context.Context, offerID string) ([]*models.JobApplication, error)
	HasApplicationsForOffer(ctx context.Context, offerID string) (bool, error)
	HasApplied(ctx context.Context, offerID, participantID string) (bool, error)
	ListByParticipant(ctx context.Context, participantID string) ([]*models.JobApplication, error)
	UpdateStatus(ctx context.Context, applicationID string, status string) error
	GetByID(ctx context.Context, applicationID string) (*models.JobApplication, error)
	Delete(ctx context.Context, applicationID string) error
}

type jobApplicationRepository struct {
	client *firestore.Client
}

func NewJobApplicationRepository(client *firestore.Client) JobApplicationRepository {
	return &jobApplicationRepository{client: client}
}

func (r *jobApplicationRepository) Create(ctx context.Context, application *models.JobApplication) error {
	ref := r.client.Collection("job_applications").NewDoc()
	application.ID = ref.ID
	_, err := ref.Set(ctx, application)
	return err
}

func (r *jobApplicationRepository) HasApplicationsForOffer(ctx context.Context, offerID string) (bool, error) {
	iter := r.client.Collection("job_applications").Where("oferta_id", "==", offerID).Limit(1).Documents(ctx)
	defer iter.Stop()
	_, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *jobApplicationRepository) HasApplied(ctx context.Context, offerID, participantID string) (bool, error) {
	iter := r.client.Collection("job_applications").
		Where("oferta_id", "==", offerID).
		Where("participante_id", "==", participantID).
		Limit(1).Documents(ctx)
	defer iter.Stop()
	_, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *jobApplicationRepository) ListByOffer(ctx context.Context, offerID string) ([]*models.JobApplication, error) {
	iter := r.client.Collection("job_applications").Where("oferta_id", "==", offerID).Documents(ctx)
	var applications []*models.JobApplication
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var app models.JobApplication
		if err := doc.DataTo(&app); err != nil {
			continue
		}
		applications = append(applications, &app)
	}
	return applications, nil
}

func (r *jobApplicationRepository) ListByParticipant(ctx context.Context, participantID string) ([]*models.JobApplication, error) {
	iter := r.client.Collection("job_applications").Where("participante_id", "==", participantID).Documents(ctx)
	var applications []*models.JobApplication
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		var app models.JobApplication
		if err := doc.DataTo(&app); err != nil {
			continue
		}
		app.ID = doc.Ref.ID
		applications = append(applications, &app)
	}
	return applications, nil
}

func (r *jobApplicationRepository) GetByID(ctx context.Context, applicationID string) (*models.JobApplication, error) {
	doc, err := r.client.Collection("job_applications").Doc(applicationID).Get(ctx)
	if err != nil {
		return nil, err
	}
	var app models.JobApplication
	if err := doc.DataTo(&app); err != nil {
		return nil, err
	}
	app.ID = doc.Ref.ID
	return &app, nil
}

func (r *jobApplicationRepository) UpdateStatus(ctx context.Context, applicationID string, status string) error {
	_, err := r.client.Collection("job_applications").Doc(applicationID).Update(ctx, []firestore.Update{
		{Path: "estado", Value: status},
		{Path: "updated_at", Value: time.Now()},
	})
	return err
}

func (r *jobApplicationRepository) Delete(ctx context.Context, applicationID string) error {
	_, err := r.client.Collection("job_applications").Doc(applicationID).Delete(ctx)
	return err
}

