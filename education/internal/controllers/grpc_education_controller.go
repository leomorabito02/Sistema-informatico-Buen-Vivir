package controllers

import (
	"context"
	"strings"

	"buenvivir/education/api/proto/educationpb"
	"buenvivir/education/internal/services"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcEducationController struct {
	educationpb.UnimplementedEducationServiceServer
	skillService services.SkillService
}

func NewGrpcEducationController(skillService services.SkillService) *GrpcEducationController {
	return &GrpcEducationController{skillService: skillService}
}

func (c *GrpcEducationController) GetGrantedSkills(ctx context.Context, _ *educationpb.GetGrantedSkillsRequest) (*educationpb.GetGrantedSkillsResponse, error) {
	if c.skillService == nil {
		return nil, status.Error(codes.Unavailable, "skill service not configured")
	}
	granted, err := c.skillService.ListGrantedSkills(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	items := make([]*educationpb.GrantedSkillItem, 0, len(granted))
	for _, g := range granted {
		items = append(items, &educationpb.GrantedSkillItem{
			ParticipanteId: g.ParticipanteID,
			Nombre:         g.Nombre,
		})
	}
	return &educationpb.GetGrantedSkillsResponse{Skills: items}, nil
}

func (c *GrpcEducationController) GetParticipantSkills(ctx context.Context, req *educationpb.GetParticipantSkillsRequest) (*educationpb.GetParticipantSkillsResponse, error) {
	if c.skillService == nil {
		return nil, status.Error(codes.Unavailable, "skill service not configured")
	}
	participanteID := strings.TrimSpace(req.GetParticipanteId())
	if participanteID == "" {
		return nil, status.Error(codes.InvalidArgument, "participante_id is required")
	}
	skills, err := c.skillService.ListParticipantSkills(ctx, participanteID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	nombres := make([]string, 0, len(skills))
	for _, s := range skills {
		if s.Nombre != "" {
			nombres = append(nombres, s.Nombre)
		}
	}
	return &educationpb.GetParticipantSkillsResponse{Nombres: nombres}, nil
}
