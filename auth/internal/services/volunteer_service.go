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
	userRepo      repositories.UserRepository
}

func NewVolunteerService(volunteerRepo repositories.VolunteerRepository, opts ...repositories.UserRepository) VolunteerService {
	var userRepo repositories.UserRepository
	if len(opts) > 0 {
		userRepo = opts[0]
	}
	return &volunteerService{
		volunteerRepo: volunteerRepo,
		userRepo:      userRepo,
	}
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

	NewKVSyncService(s.userRepo).SyncImpactStatsAsync()

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
			volunteer.ApprovedBy = resolveApprover(ctx, req.ApprovedBy)
		}
	}

	if err := s.volunteerRepo.Update(ctx, volunteer); err != nil {
		return nil, err
	}

	NewKVSyncService(s.userRepo).SyncImpactStatsAsync()

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

	if err := s.volunteerRepo.Delete(ctx, id); err != nil {
		return err
	}

	NewKVSyncService(s.userRepo).SyncImpactStatsAsync()

	return nil
}
