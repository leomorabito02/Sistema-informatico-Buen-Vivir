package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"auth/internal/dto"
	"auth/internal/events"
	"auth/internal/events/producers"
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
	ApproveUser(ctx context.Context, id string) (*dto.UserResponse, error)
	// RejectUser rechaza la cuenta y avisa por email (HU-12, HU-38).
	RejectUser(ctx context.Context, id string) (*dto.UserResponse, error)
	GetCampaignEligibleUsers(ctx context.Context) ([]repositories.CampaignEligibleUser, error)
	IsEmailVerified(ctx context.Context, identifier string) (bool, error)
	GetActiveUserIDsByRoles(ctx context.Context, roles []string) ([]string, error)
}

type userService struct {
	userRepo            repositories.UserRepository
	querier             repositories.RoleProfileQuerier
	notificationSync    NotificationSyncService
	verificationService VerificationService
	producer            producers.UsuariosProducer
}

func NewUserService(userRepo repositories.UserRepository, opts ...any) UserService {
	svc := &userService{userRepo: userRepo}
	for _, opt := range opts {
		switch v := opt.(type) {
		case repositories.RoleProfileQuerier:
			svc.querier = v
		case NotificationSyncService:
			svc.notificationSync = v
		case VerificationService:
			svc.verificationService = v
		case producers.UsuariosProducer:
			svc.producer = v
		}
	}
	return svc
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

	if strings.TrimPrefix(req.RoleID, "role_") == "participant" {
		if err := validateParticipantAccount(req.Email, req.DNI, req.Address); err != nil {
			return nil, err
		}
		ctx = repositories.WithParticipantProfile(ctx, strings.TrimSpace(req.DNI), strings.TrimSpace(req.Address))
	}
	if req.RoleID == "role_teacher" && req.IsExternal {
		ctx = repositories.WithTeacherExternal(ctx, true)
	}

	// 3. Crear modelo. La cuenta que da de alta el administrador (o el
	// voluntario) nace con el correo verificado y aprobada para todos los
	// roles: la persona entra con la contraseña que le dieron y recibe el
	// mail "tu cuenta fue aprobada" (22/11/2026). Antes nacía sin verificar
	// y, si no tocaba el enlace del mail, no podía iniciar sesión.
	user := &models.User{
		UserID:          "usr-" + utils.GenerateRandomString(12), // helper pseudo-id o UUID
		FirstName:       strings.TrimSpace(req.FirstName),
		LastName:        strings.TrimSpace(req.LastName),
		Email:           strings.TrimSpace(req.Email),
		PasswordHash:    string(hashedPassword),
		Phone:           strings.TrimSpace(req.Phone),
		Status:          "active",
		RoleID:          req.RoleID,
		IsEmailVerified: true,
	}

	// 4. Guardar en Base de Datos (a través del repositorio con transacciones)
	reqRole, _ := ctx.Value(middlewares.RoleIDKey).(string)
	creatorID, _ := ctx.Value(middlewares.UserIDKey).(string)

	// Admin y dev crean cualquier rol; el voluntario, solo participantes
	// (HU-07). El resto no crea cuentas: antes cualquier usuario con
	// sesión podía crear un administrador.
	if !isAdminRequest(ctx) && (reqRole != "role_volunteer" || user.RoleID != "role_participant") {
		return nil, errors.New("forbidden: access denied to create accounts with this role")
	}

	if err := s.userRepo.CreateWithRoleProfile(ctx, user, user.RoleID, creatorID); err != nil {
		return nil, err
	}

	if s.producer != nil {
		_ = s.producer.UsuarioCreado(ctx, user.UserID, user.RoleID)
		// Mismo mail que cuando el admin aprueba un registro: avisa que
		// la cuenta está lista y trae el enlace para iniciar sesión.
		_ = s.producer.EmailSolicitado(ctx, "app-"+user.UserID, user.UserID, "cuenta_aprobada")
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

	oldStatus := user.Status
	oldRole := user.RoleID

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

	if s.producer != nil && (oldStatus != user.Status || oldRole != user.RoleID) {
		_ = s.producer.UsuarioActualizado(ctx, events.UsuarioActualizadoPayload{
			UserID: user.UserID,
			RoleID: user.RoleID,
			Status: user.Status,
		})
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

	if s.producer != nil && user.RoleID == "role_participant" {
		_ = s.producer.ParticipanteEliminado(ctx, user.UserID, user.UserID)
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

func (s *userService) ApproveUser(ctx context.Context, id string) (*dto.UserResponse, error) {
	return s.setApproval(ctx, id, "approved", "cuenta_aprobada")
}

func (s *userService) RejectUser(ctx context.Context, id string) (*dto.UserResponse, error) {
	return s.setApproval(ctx, id, "rejected", "cuenta_rechazada")
}

// setApproval cambia approval_status en la tabla del rol y pide el email
// correspondiente (cuenta_aprobada o cuenta_rechazada).
func (s *userService) setApproval(ctx context.Context, id, status, emailTipo string) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !canApprove(ctx, user.RoleID) {
		return nil, errors.New("forbidden: access denied to approve accounts with this role")
	}

	table := roleTableFromID(user.RoleID)
	if table == "" {
		return nil, errors.New("rol no compatible para aprobación")
	}

	approverID, _ := ctx.Value(middlewares.UserIDKey).(string)
	now := time.Now()
	updateData := map[string]any{
		"approval_status": status,
		"approval_date":   &now,
	}
	if approverID != "" {
		updateData["approved_by"] = approverID
	}

	if s.querier != nil {
		if err := s.querier.UpdateRoleData(ctx, table, user.UserID, updateData); err != nil {
			return nil, fmt.Errorf("error al actualizar el estado de aprobación: %w", err)
		}
	}

	user.RoleApprovalStatus = status

	if s.producer != nil {
		prefix := "app-"
		if status == "rejected" {
			prefix = "rej-"
		}
		_ = s.producer.EmailSolicitado(ctx, prefix+user.UserID, user.UserID, emailTipo)
	}

	return dto.ToUserResponse(user), nil
}

func (s *userService) GetCampaignEligibleUsers(ctx context.Context) ([]repositories.CampaignEligibleUser, error) {
	return s.userRepo.FindCampaignEligibleUsers(ctx)
}

func (s *userService) IsEmailVerified(ctx context.Context, identifier string) (bool, error) {
	trimmed := strings.TrimSpace(identifier)
	if trimmed == "" {
		return false, errors.New("identifier cannot be empty")
	}
	var user *models.User
	var err error
	if strings.Contains(trimmed, "@") {
		user, err = s.userRepo.FindByEmail(ctx, trimmed)
	} else {
		user, err = s.userRepo.FindByID(ctx, trimmed)
	}
	if err != nil || user == nil {
		return false, err
	}
	return user.IsEmailVerified, nil
}

func (s *userService) GetActiveUserIDsByRoles(ctx context.Context, roles []string) ([]string, error) {
	validRoles := make([]string, 0, len(roles))
	for _, r := range roles {
		trimmed := strings.TrimSpace(r)
		if trimmed == "role_admin" || trimmed == "role_volunteer" || trimmed == "role_company" || trimmed == "role_participant" || trimmed == "role_teacher" {
			validRoles = append(validRoles, trimmed)
		}
	}
	if len(validRoles) == 0 {
		return []string{}, nil
	}
	return s.userRepo.FindActiveUserIDsByRoles(ctx, validRoles)
}

func roleTableFromID(roleID string) string {
	switch roleID {
	case "role_teacher":
		return "teachers"
	case "role_volunteer":
		return "volunteers"
	case "role_participant":
		return "participants"
	case "role_company":
		return "companies"
	default:
		return ""
	}
}
