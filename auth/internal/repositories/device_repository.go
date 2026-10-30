package repositories

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
)

type DeviceToken struct {
	Token     string    `json:"token" firestore:"token"`
	UserAgent string    `json:"userAgent" firestore:"userAgent"`
	UpdatedAt time.Time `json:"updatedAt" firestore:"updatedAt"`
}

type DeviceRepository interface {
	RegisterDevice(ctx context.Context, userID, token, userAgent string) error
	GetUserTokens(ctx context.Context, userID string) ([]string, error)
	DeleteDeviceTokens(ctx context.Context, userID string, tokens []string) error
}

type firestoreDeviceRepository struct {
	client *firestore.Client
}

func NewFirestoreDeviceRepository(client *firestore.Client) DeviceRepository {
	return &firestoreDeviceRepository{client: client}
}

func (r *firestoreDeviceRepository) RegisterDevice(ctx context.Context, userID, token, userAgent string) error {
	if r.client == nil {
		return nil
	}
	device := DeviceToken{
		Token:     token,
		UserAgent: userAgent,
		UpdatedAt: time.Now(),
	}
	_, err := r.client.Collection("users").Doc(userID).Collection("devices").Doc(token).Set(ctx, device)
	return err
}

func (r *firestoreDeviceRepository) GetUserTokens(ctx context.Context, userID string) ([]string, error) {
	if r.client == nil || userID == "" {
		return nil, nil
	}
	docs, err := r.client.Collection("users").Doc(userID).Collection("devices").Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	tokens := make([]string, 0, len(docs))
	for _, doc := range docs {
		tokens = append(tokens, doc.Ref.ID)
	}
	return tokens, nil
}

func (r *firestoreDeviceRepository) DeleteDeviceTokens(ctx context.Context, userID string, tokens []string) error {
	if r.client == nil || userID == "" || len(tokens) == 0 {
		return nil
	}
	return r.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		for _, t := range tokens {
			if t != "" {
				docRef := r.client.Collection("users").Doc(userID).Collection("devices").Doc(t)
				if err := tx.Delete(docRef); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
