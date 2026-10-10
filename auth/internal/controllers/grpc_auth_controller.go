package controllers

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"auth/api/proto/authpb"
	"auth/internal/middlewares"
	"auth/internal/repositories"
	"auth/internal/services"
)

// serviceCtx marca una llamada gRPC como interna entre servicios. En gRPC no
// viaja el usuario que originó el pedido, así que los servicios que chequean
// permisos por rol (GetUserByID) la rechazaban siempre. Quien llama ya
// validó al usuario en su propio endpoint (Inserción Laboral, por el
// gateway); acá se lee con permisos de sistema.
func serviceCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, middlewares.RoleIDKey, "role_dev")
}

type GrpcAuthController struct {
	authpb.UnimplementedAuthServiceServer
	participantService services.ParticipantService
	companyService     services.CompanyService
	userService        services.UserService
	directoryService   services.ParticipantDirectoryService
}

func NewGrpcAuthController(participantService services.ParticipantService, companyService services.CompanyService, userService services.UserService, directoryService services.ParticipantDirectoryService) *GrpcAuthController {
	return &GrpcAuthController{
		participantService: participantService,
		companyService:     companyService,
		userService:        userService,
		directoryService:   directoryService,
	}
}

func (c *GrpcAuthController) GetParticipant(ctx context.Context, req *authpb.GetParticipantRequest) (*authpb.GetParticipantResponse, error) {
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

func (c *GrpcAuthController) GetCompany(ctx context.Context, req *authpb.GetCompanyRequest) (*authpb.GetCompanyResponse, error) {
	if req.CompanyId == "" {
		return nil, status.Error(codes.InvalidArgument, "company_id is required")
	}

	company, err := c.companyService.GetCompanyByID(ctx, req.CompanyId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if company == nil {
		return nil, status.Error(codes.NotFound, "company not found")
	}

	user, err := c.userService.GetUserByID(serviceCtx(ctx), company.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &authpb.GetCompanyResponse{
		CompanyId:      company.CompanyID,
		UserId:         company.UserID,
		BusinessName:   company.BusinessName,
		Cuit:           company.Cuit,
		ApprovalStatus: company.ApprovalStatus,
		UserEmail:      user.Email,
		UserPhone:      user.Phone,
		UserStatus:     user.Status,
	}, nil
}

func (c *GrpcAuthController) GetParticipantProfile(ctx context.Context, req *authpb.GetParticipantProfileRequest) (*authpb.ParticipantProfile, error) {
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

func (c *GrpcAuthController) SearchParticipants(ctx context.Context, req *authpb.SearchParticipantsRequest) (*authpb.SearchParticipantsResponse, error) {
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

func (c *GrpcAuthController) GetUsersByIDs(ctx context.Context, req *authpb.GetUsersByIDsRequest) (*authpb.GetUsersByIDsResponse, error) {
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
