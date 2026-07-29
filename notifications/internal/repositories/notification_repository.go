package repositories

import (
	"context"
	"time"

	"buenvivir/notifications/internal/models"
)

type NotificationRepository interface {
	FindByUserID(ctx context.Context, userID string) ([]*models.Notification, error)
}

type notificationRepository struct{}

func NewNotificationRepository() NotificationRepository {
	return &notificationRepository{}
}

func (r *notificationRepository) FindByUserID(ctx context.Context, userID string) ([]*models.Notification, error) {
	return []*models.Notification{
		{ID: "ntf-1", UserID: userID, Message: "Bienvenido a la PWA Buen Vivir", Read: false, CreatedAt: time.Now()},
	}, nil
}
