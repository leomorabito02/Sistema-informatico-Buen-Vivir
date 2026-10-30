package controllers

import (
	"context"
	"errors"

	"buenvivir/notifications/api/proto/notificationpb"
	"buenvivir/notifications/internal/dto"
	"buenvivir/notifications/internal/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcNotificationController struct {
	notificationpb.UnimplementedNotificationIngestionServiceServer
	service services.NotificationService
}

func NewGrpcNotificationController(service services.NotificationService) *GrpcNotificationController {
	return &GrpcNotificationController{service: service}
}

func (c *GrpcNotificationController) SendNotification(ctx context.Context, req *notificationpb.SendNotificationRequest) (*notificationpb.SendNotificationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	dtoReq := dto.SendNotificationRequestDTO{
		SourceService:    req.SourceService,
		EventType:        req.EventType,
		Category:         req.Category,
		Priority:         req.Priority,
		SenderType:       req.SenderType,
		SenderID:         req.SenderId,
		TargetType:       req.TargetType,
		RecipientUserIDs: req.RecipientUserIds,
		Title:            req.Title,
		Message:          req.Message,
		Data:             req.DataJson,
		IsCritical:       req.IsCritical,
		SendEmail:        req.SendEmail,
		SendPush:         req.SendPush,
		SendInApp:        req.SendInApp,
	}
	if req.TargetRole != "" {
		dtoReq.TargetRole = &req.TargetRole
	}
	if req.ActionUrl != "" {
		dtoReq.ActionURL = &req.ActionUrl
	}

	notifID, count, err := c.service.IngestNotification(ctx, dtoReq)
	if err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &notificationpb.SendNotificationResponse{
		NotificationId:  notifID,
		RecipientsCount: int32(count),
		Status:          "created",
	}, nil
}

func (c *GrpcNotificationController) SendBatchNotifications(ctx context.Context, req *notificationpb.SendBatchNotificationsRequest) (*notificationpb.SendBatchNotificationsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	createdIDs := make([]string, 0, len(req.Notifications))
	var totalCount int32

	for _, item := range req.Notifications {
		res, err := c.SendNotification(ctx, item)
		if err != nil {
			return nil, err
		}
		createdIDs = append(createdIDs, res.NotificationId)
		totalCount += res.RecipientsCount
	}

	return &notificationpb.SendBatchNotificationsResponse{
		CreatedCount:    totalCount,
		NotificationIds: createdIDs,
	}, nil
}

// InitDefaultPreferences initializes default notification preferences via gRPC for a user.
func (c *GrpcNotificationController) InitDefaultPreferences(ctx context.Context, req *notificationpb.InitDefaultPreferencesRequest) (*notificationpb.InitDefaultPreferencesResponse, error) {
	if req == nil || req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if err := c.service.InitDefaultPreferences(ctx, req.UserId); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &notificationpb.InitDefaultPreferencesResponse{
		UserId:  req.UserId,
		Success: true,
		Count:   5,
	}, nil
}

