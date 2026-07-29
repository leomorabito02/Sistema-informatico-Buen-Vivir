package services

import (
	"context"

	"buenvivir/notifications/internal/models"
	"buenvivir/notifications/internal/repositories"
)

type NotificationService interface {
	GetNotifications(ctx context.Context, userID string) ([]*models.Notification, error)
}

type notificationService struct {
	repo repositories.NotificationRepository
}

func NewNotificationService(repo repositories.NotificationRepository) NotificationService {
	return &notificationService{repo: repo}
}

func (s *notificationService) GetNotifications(ctx context.Context, userID string) ([]*models.Notification, error) {
	return s.repo.FindByUserID(ctx, userID)
}
