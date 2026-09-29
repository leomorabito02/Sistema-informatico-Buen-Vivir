package controllers

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"auth/api/proto/authpb"
	"auth/internal/services"
)

type GrpcParticipantController struct {
	authpb.UnimplementedAuthServiceServer
	participantService services.ParticipantService
}

func NewGrpcParticipantController(participantService services.ParticipantService) *GrpcParticipantController {
	return &GrpcParticipantController{
		participantService: participantService,
	}
}

func (c *GrpcParticipantController) GetParticipant(ctx context.Context, req *authpb.GetParticipantRequest) (*authpb.GetParticipantResponse, error) {
	if req.ParticipantId == "" {
		return nil, status.Error(codes.InvalidArgument, "participant_id is required")
	}

	participant, err := c.participantService.GetParticipantByID(ctx, req.ParticipantId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if participant == nil {
		return nil, status.Error(codes.NotFound, "participant not found")
	}

	return &authpb.GetParticipantResponse{
		Id:             participant.ParticipantID,
		UserId:         participant.UserID,
		ApprovalStatus: participant.ApprovalStatus,
	}, nil
}
