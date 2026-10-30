package repositories

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"buenvivir/labor-integration/internal/models"
)

type CompanyRepository interface {
	Get(ctx context.Context, id string) (*models.Company, error)
	Create(ctx context.Context, company *models.Company) error
}

type companyRepository struct {
	client *firestore.Client
}

func NewCompanyRepository(client *firestore.Client) CompanyRepository {
	return &companyRepository{client: client}
}

func (r *companyRepository) Get(ctx context.Context, id string) (*models.Company, error) {
	doc, err := r.client.Collection("companies").Doc(id).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil // Return nil if not found
		}
		return nil, err
	}

	var company models.Company
	if err := doc.DataTo(&company); err != nil {
		return nil, err
	}
	company.ID = doc.Ref.ID
	return &company, nil
}

func (r *companyRepository) Create(ctx context.Context, company *models.Company) error {
	ref := r.client.Collection("companies").NewDoc()
	company.ID = ref.ID
	_, err := ref.Set(ctx, company)
	return err
}
