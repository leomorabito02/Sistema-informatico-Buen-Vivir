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
	"io"
	"auth/internal/storage"
)

type ParticipantService interface {
	CreateParticipant(ctx context.Context, req *dto.CreateParticipantRequest) (*dto.ParticipantResponse, error)
	GetParticipantByID(ctx context.Context, id string) (*dto.ParticipantResponse, error)
	GetAllParticipants(ctx context.Context) ([]dto.ParticipantResponse, error)
	UpdateParticipant(ctx context.Context, id string, req *dto.UpdateParticipantRequest) (*dto.ParticipantResponse, error)
	DeleteParticipant(ctx context.Context, id string) error

	UploadDocument(ctx context.Context, participantID string, docName string, file io.Reader, contentType string, originalName string) (*models.ParticipantDocument, error)
	GetDocuments(ctx context.Context, participantID string) ([]models.ParticipantDocument, error)
	UpdateDocument(ctx context.Context, participantID string, docID string, newName string, file io.Reader, contentType string, originalName string) (*models.ParticipantDocument, error)
	DeleteDocument(ctx context.Context, participantID string, docID string) error
	DownloadDocument(ctx context.Context, participantID string, docID string) (io.ReadCloser, string, string, error)
}

type participantService struct {
	participantRepo repositories.ParticipantRepository
	userRepo        repositories.UserRepository
	docRepo         repositories.ParticipantDocumentRepository
	storage         storage.ObjectStorageService
	cleanupRepo     repositories.StorageCleanupRepository
}

func NewParticipantService(
	participantRepo repositories.ParticipantRepository,
	docRepo repositories.ParticipantDocumentRepository,
	storage storage.ObjectStorageService,
	opts ...any,
) ParticipantService {
	var userRepo repositories.UserRepository
	var cleanupRepo repositories.StorageCleanupRepository

	for _, opt := range opts {
		switch v := opt.(type) {
		case repositories.UserRepository:
			userRepo = v
		case repositories.StorageCleanupRepository:
			cleanupRepo = v
		}
	}

	return &participantService{
		participantRepo: participantRepo,
		docRepo:         docRepo,
		storage:         storage,
		userRepo:        userRepo,
		cleanupRepo:     cleanupRepo,
	}
}

func (s *participantService) CreateParticipant(ctx context.Context, req *dto.CreateParticipantRequest) (*dto.ParticipantResponse, error) {
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	status := req.ApprovalStatus
	if status == "" {
		status = "pending"
	}

	participant := &models.Participant{
		ParticipantID:  "part-" + utils.GenerateRandomString(10),
		UserID:         req.UserID,
		Birthdate:      req.Birthdate,
		DNI:            req.DNI,
		Address:        req.Address,
		Gender:         req.Gender,
		EmergencyContactName: req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
		ImageURL:       req.ImageURL,
		DocURL:         req.DocURL,
		ApprovalStatus: status,
	}

	if err := s.participantRepo.Create(ctx, participant); err != nil {
		return nil, err
	}

	NewKVSyncService(s.userRepo).SyncImpactStatsAsync()

	return dto.ToParticipantResponse(participant), nil
}

func (s *participantService) GetParticipantByID(ctx context.Context, id string) (*dto.ParticipantResponse, error) {
	participant, err := s.participantRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_participant" {
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if participant.UserID != reqUserID {
			return nil, errors.New("forbidden: access denied to other participant profiles")
		}
	}

	return dto.ToParticipantResponse(participant), nil
}

func (s *participantService) GetAllParticipants(ctx context.Context) ([]dto.ParticipantResponse, error) {
	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_participant" {
		return nil, errors.New("forbidden: cannot list all participants")
	}

	participants, err := s.participantRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	var resp []dto.ParticipantResponse
	for _, p := range participants {
		resp = append(resp, *dto.ToParticipantResponse(&p))
	}
	return resp, nil
}

func checkParticipantUpdateAccess(ctx context.Context, participant *models.Participant) error {
	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_participant" {
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if participant.UserID != reqUserID {
			return errors.New("forbidden: access denied to edit other participant profiles")
		}
	}
	return nil
}

func applyUpdates(ctx context.Context, participant *models.Participant, req *dto.UpdateParticipantRequest) {
	if req.Birthdate != nil {
		participant.Birthdate = *req.Birthdate
	}
	if req.DNI != "" {
		participant.DNI = req.DNI
	}
	if req.Address != "" {
		participant.Address = req.Address
	}
	if req.Gender != "" {
		participant.Gender = req.Gender
	}
	if req.EmergencyContactName != "" {
		participant.EmergencyContactName = req.EmergencyContactName
	}
	if req.EmergencyContactPhone != "" {
		participant.EmergencyContactPhone = req.EmergencyContactPhone
	}
	if req.ImageURL != "" {
		participant.ImageURL = req.ImageURL
	}
	if req.DocURL != "" {
		participant.DocURL = req.DocURL
	}

	if req.ApprovalStatus != "" {
		if req.ApprovalStatus == "approved" {
			participant.ApprovalStatus = approvalStatusApproved
			now := time.Now()
			participant.ApprovalDate = &now
			participant.ApprovedBy = resolveApprover(ctx, req.ApprovedBy)
		} else {
			participant.ApprovalStatus = req.ApprovalStatus
		}
	}
}

func (s *participantService) UpdateParticipant(ctx context.Context, id string, req *dto.UpdateParticipantRequest) (*dto.ParticipantResponse, error) {
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	participant, err := s.participantRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := checkParticipantUpdateAccess(ctx, participant); err != nil {
		return nil, err
	}

	applyUpdates(ctx, participant, req)

	if err := s.participantRepo.Update(ctx, participant); err != nil {
		return nil, err
	}

	NewKVSyncService(s.userRepo).SyncImpactStatsAsync()

	return dto.ToParticipantResponse(participant), nil
}

func (s *participantService) DeleteParticipant(ctx context.Context, id string) error {
	participant, err := s.participantRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_participant" {
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if participant.UserID != reqUserID {
			return errors.New("forbidden: access denied to delete other participant profiles")
		}
	}

	if err := s.participantRepo.Delete(ctx, id); err != nil {
		return err
	}

	NewKVSyncService(s.userRepo).SyncImpactStatsAsync()

	return nil
}
