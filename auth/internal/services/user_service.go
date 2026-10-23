package services

import (
	"context"
	"errors"
	"fmt"

	"auth/internal/dto"
	"auth/internal/middlewares"
	"auth/internal/models"
	"auth/internal/repositories"
	"auth/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*dto.UserResponse, error)
	GetAllUsers(ctx context.Context) ([]dto.UserResponse, error)
	UpdateUser(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
	GetParticipantsHistory(ctx context.Context) ([]dto.ParticipantHistoryResponse, error)
}

type userService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// 1. Validar request
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	// 2. Hashear password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error encrypting password: %w", err)
	}

	// 3. Crear modelo
	user := &models.User{
		UserID:       "usr-" + utils.GenerateRandomString(12), // helper pseudo-id o UUID
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Phone:        req.Phone,
		Status:       "active",
		RoleID:       req.RoleID,
	}

	// 4. Guardar en Base de Datos (a través del repositorio con transacciones)
	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	creatorID, _ := ctx.Value(middlewares.UserIDKey).(string)

	// Admin y dev crean cualquier rol; el voluntario, solo participantes
	// (HU-07). El resto no crea cuentas: antes cualquier usuario con
	// sesión podía crear un administrador.
	if !isAdminRequest(ctx) && !(reqRole == "role_volunteer" && user.RoleID == "role_participant") {
		return nil, errors.New("forbidden: access denied to create accounts with this role")
	}

	if err := s.userRepo.CreateWithRoleProfile(ctx, user, user.RoleID, creatorID); err != nil {
		return nil, err
	}

	// Sincronizar stats de impacto y preferencias de notificaciones
	NewKVSyncService(s.userRepo).SyncImpactStatsAsync()
	NewNotificationSyncService().InitUserPreferencesAsync(user.UserID)

	return dto.ToUserResponse(user), nil
}

func (s *userService) GetUserByID(ctx context.Context, id string) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole != "role_admin" && reqRole != "role_dev" {
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if user.UserID != reqUserID {
			if reqRole == "role_volunteer" && user.RoleID == "role_participant" {
				// Volunteers can view participants
			} else {
				return nil, errors.New("forbidden: access denied to other user accounts")
			}
		}
	}

	return dto.ToUserResponse(user), nil
}

func (s *userService) GetAllUsers(ctx context.Context) ([]dto.UserResponse, error) {
	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole != "role_admin" && reqRole != "role_dev" && reqRole != "role_volunteer" {
		return nil, errors.New("forbidden: insufficient permissions to list users")
	}

	users, err := s.userRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}
	var resp []dto.UserResponse
	for _, u := range users {
		if reqRole == "role_volunteer" && u.RoleID != "role_participant" {
			continue
		}
		resp = append(resp, *dto.ToUserResponse(&u))
	}
	return resp, nil
}

func (s *userService) UpdateUser(ctx context.Context, id string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	// 1. Validar request
	val := utils.GetValidator()
	if err := val.Struct(req); err != nil {
		return nil, err
	}

	// 2. Obtener usuario existente
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.checkUpdatePermissions(ctx, user, req); err != nil {
		return nil, err
	}

	// 3. Modificar campos permitidos
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.RoleID != "" {
		user.RoleID = req.RoleID
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	// 4. Actualizar
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Sincronizar stats de impacto (por si cambió de rol)
	NewKVSyncService(s.userRepo).SyncImpactStatsAsync()

	return dto.ToUserResponse(user), nil
}

func (s *userService) DeleteUser(ctx context.Context, id string) error {
	// Verificar si existe primero
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole != "role_admin" && reqRole != "role_dev" {
		reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
		if user.UserID != reqUserID {
			return errors.New("forbidden: access denied to delete other user accounts")
		}
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	NewKVSyncService(s.userRepo).SyncImpactStatsAsync()

	return nil
}

func (s *userService) checkUpdatePermissions(ctx context.Context, user *models.User, req *dto.UpdateUserRequest) error {
	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole == "role_admin" || reqRole == "role_dev" {
		return nil
	}

	reqUserID, _ := ctx.Value(middlewares.UserIDKey).(string)
	if user.UserID != reqUserID {
		if reqRole == "role_volunteer" && user.RoleID == "role_participant" {
			// Volunteers can edit participants
		} else {
			return errors.New("forbidden: access denied to edit other user accounts")
		}
	}
	
	// Si no es admin, tampoco puede cambiar su rol ni status
	req.RoleID = ""
	req.Status = ""
	return nil
}

func (s *userService) GetParticipantsHistory(ctx context.Context) ([]dto.ParticipantHistoryResponse, error) {
	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	if reqRole != "role_admin" && reqRole != "role_dev" {
		return nil, errors.New("forbidden: only administrators can see the participants history")
	}
	rows, err := s.userRepo.FindParticipantsHistory(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ParticipantHistoryResponse, len(rows))
	for i, r := range rows {
		result[i] = dto.ParticipantHistoryResponse{
			UserID:         r.UserID,
			Status:         r.Status,
			ApprovalStatus: r.ApprovalStatus,
			ApprovalDate:   r.ApprovalDate,
			CreatedAt:      r.CreatedAt,
			UpdatedAt:      r.UpdatedAt,
			DeletedAt:      r.DeletedAt,
		}
	}
	return result, nil
}
