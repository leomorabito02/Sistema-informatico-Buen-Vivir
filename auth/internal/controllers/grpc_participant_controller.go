package controllers

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"auth/api/proto/authpb"
	"auth/internal/repositories"
	"auth/internal/services"
)

type GrpcParticipantController struct {
	authpb.UnimplementedAuthServiceServer
	participantService services.ParticipantService
	directoryService   services.ParticipantDirectoryService
}

func NewGrpcParticipantController(participantService services.ParticipantService, directoryService services.ParticipantDirectoryService) *GrpcParticipantController {
	return &GrpcParticipantController{
		participantService: participantService,
		directoryService:   directoryService,
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

// GetParticipantProfile devuelve nombre, apellido y fecha de nacimiento
// del participante, buscándolo por participant_id o por user_id.
func (c *GrpcParticipantController) GetParticipantProfile(ctx context.Context, req *authpb.GetParticipantProfileRequest) (*authpb.ParticipantProfile, error) {
	var (
		row *repositories.ParticipantProfileRow
		err error
	)
	switch lookup := req.GetLookup().(type) {
	case *authpb.GetParticipantProfileRequest_ParticipantId:
		row, err = c.directoryService.GetProfileByParticipantID(ctx, lookup.ParticipantId)
	case *authpb.GetParticipantProfileRequest_UserId:
		row, err = c.directoryService.GetProfileByUserID(ctx, lookup.UserId)
	default:
		return nil, status.Error(codes.InvalidArgument, "participant_id or user_id is required")
	}
	if err != nil {
		return nil, directoryError(err)
	}
	return toParticipantProfile(row), nil
}

// SearchParticipants busca participantes por coincidencia parcial de
// nombre o apellido.
func (c *GrpcParticipantController) SearchParticipants(ctx context.Context, req *authpb.SearchParticipantsRequest) (*authpb.SearchParticipantsResponse, error) {
	rows, err := c.directoryService.SearchParticipants(ctx, req.GetQuery(), int(req.GetLimit()))
	if err != nil {
		return nil, directoryError(err)
	}
	out := make([]*authpb.ParticipantProfile, 0, len(rows))
	for i := range rows {
		out = append(out, toParticipantProfile(&rows[i]))
	}
	return &authpb.SearchParticipantsResponse{Participants: out}, nil
}

// GetUsersByIDs devuelve nombre y apellido de cada usuario encontrado;
// los ids inexistentes simplemente no aparecen en la respuesta.
func (c *GrpcParticipantController) GetUsersByIDs(ctx context.Context, req *authpb.GetUsersByIDsRequest) (*authpb.GetUsersByIDsResponse, error) {
	rows, err := c.directoryService.GetUserSummaries(ctx, req.GetUserIds())
	if err != nil {
		return nil, directoryError(err)
	}
	out := make([]*authpb.UserSummary, 0, len(rows))
	for _, row := range rows {
		out = append(out, &authpb.UserSummary{
			UserId:    row.UserID,
			FirstName: row.FirstName,
			LastName:  row.LastName,
		})
	}
	return &authpb.GetUsersByIDsResponse{Users: out}, nil
}

func toParticipantProfile(row *repositories.ParticipantProfileRow) *authpb.ParticipantProfile {
	birthdate := ""
	if !row.Birthdate.IsZero() {
		birthdate = row.Birthdate.Format("2006-01-02")
	}
	return &authpb.ParticipantProfile{
		ParticipantId:  row.ParticipantID,
		UserId:         row.UserID,
		FirstName:      row.FirstName,
		LastName:       row.LastName,
		Birthdate:      birthdate,
		ApprovalStatus: row.ApprovalStatus,
	}
}

// directoryError traduce los errores del directorio a códigos gRPC sin
// exponer detalles internos de la base.
func directoryError(err error) error {
	switch {
	case errors.Is(err, services.ErrInvalidDirectoryRequest):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, repositories.ErrParticipantProfileNotFound):
		return status.Error(codes.NotFound, "participant not found")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
