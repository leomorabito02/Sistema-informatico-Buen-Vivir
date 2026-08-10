package services

import (
	"context"
	"errors"
	"time"

	"auth/internal/dto"
	"auth/internal/middlewares"
	"auth/internal/models"
	"auth/internal/repositories"
	"auth/internal/utils"
)

type VolunteerService interface {
	CreateVolunteer(ctx context.Context, req *dto.CreateVolunteerRequest) (*dto.VolunteerResponse, error)
	GetVolunteerByID(ctx context.Context, id string) (*dto.VolunteerResponse, error)
	GetAllVolunteers(ctx context.Context) ([]dto.VolunteerResponse, error)
	UpdateVolunteer(ctx context.Context, id string, req *dto.UpdateVolunteerRequest) (*dto.VolunteerResponse, error)
	DeleteVolunteer(ctx context.Context, id string) error
}

type volunteerService struct {
	volunteerRepo repositories.VolunteerRepository
}

func NewVolunteerService(volunteerRepo repositories.VolunteerRepository) VolunteerService {
	return &volunteerService{volunteerRepo: volunteerRepo}
}

func (s *volunteerService) CreateVolunteer(ctx context.Context, req *dto.CreateVolunteerRequest) (*dto.VolunteerResponse, error) {
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	status := req.ApprovalStatus
	if status == "" {
		status = "pending"
	}

	volunteer := &models.Volunteer{
		VolunteerID:    "vol-" + utils.GenerateRandomString(10),
		UserID:         req.UserID,
		ApprovalStatus: status,
	}

	if err := s.volunteerRepo.Create(ctx, volunteer); err != nil {
		return nil, err
	}

	return dto.ToVolunteerResponse(volunteer), nil
}

func (s *volunteerService) GetVolunteerByID(ctx context.Context, id string) (*dto.VolunteerResponse, error) {
	volunteer, err := s.volunteerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return dto.ToVolunteerResponse(volunteer), nil
}

func (s *volunteerService) GetAllVolunteers(ctx context.Context) ([]dto.VolunteerResponse, error) {
	volunteers, err := s.volunteerRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	var resp []dto.VolunteerResponse
	for _, v := range volunteers {
		resp = append(resp, *dto.ToVolunteerResponse(&v))
	}
	return resp, nil
}

func (s *volunteerService) UpdateVolunteer(ctx context.Context, id string, req *dto.UpdateVolunteerRequest) (*dto.VolunteerResponse, error) {
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	volunteer, err := s.volunteerRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_volunteer" {
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if volunteer.UserID != reqUserID {
			return nil, errors.New("forbidden: access denied to edit other volunteer profiles")
		}
	}

	if req.ApprovalStatus != "" {
		volunteer.ApprovalStatus = req.ApprovalStatus
		if req.ApprovalStatus == "approved" {
			now := time.Now()
			volunteer.ApprovalDate = &now
			volunteer.ApprovedBy = &req.ApprovedBy
		}
	}

	if err := s.volunteerRepo.Update(ctx, volunteer); err != nil {
		return nil, err
	}

	return dto.ToVolunteerResponse(volunteer), nil
}

func (s *volunteerService) DeleteVolunteer(ctx context.Context, id string) error {
	volunteer, err := s.volunteerRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_volunteer" {
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if volunteer.UserID != reqUserID {
			return errors.New("forbidden: access denied to delete other volunteer profiles")
		}
	}

	return s.volunteerRepo.Delete(ctx, id)
}
