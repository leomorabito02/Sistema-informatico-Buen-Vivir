package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"auth/internal/dto"
	"auth/internal/models"
	"auth/internal/repositories"
	"auth/internal/utils"

	"gorm.io/gorm"
)

type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.LoginResponse, error)
	LoginWithGoogle(ctx context.Context, req *dto.GoogleLoginRequest) (*dto.LoginResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	GetApprovalStatus(ctx context.Context, roleID, userID string) string
}

type authService struct {
	db               *gorm.DB
	userRepo         repositories.UserRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	jwtService       JWTService
	firebaseService  FirebaseService
}

func NewAuthService(
	db *gorm.DB,
	userRepo repositories.UserRepository,
	rtRepo repositories.RefreshTokenRepository,
	jwtService JWTService,
	fbService FirebaseService,
) AuthService {
	return &authService{
		db:               db,
		userRepo:         userRepo,
		refreshTokenRepo: rtRepo,
		jwtService:       jwtService,
		firebaseService:  fbService,
	}
}

func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.LoginResponse, error) {
	if req.Email == "" || req.Password == "" || req.Role == "" {
		return nil, errors.New("email, contraseña y rol son obligatorios")
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && user != nil {
		return nil, errors.New("el usuario ya existe")
	}

	roleID := req.Role
	if !strings.HasPrefix(roleID, "role_") {
		roleID = "role_" + roleID
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("error al procesar la contraseña")
	}

	user = &models.User{
		UserID:       "usr-" + utils.GenerateRandomString(10),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: hashedPassword,
		RoleID:       roleID,
		Status:       "active",
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return errors.New("error al crear el usuario")
		}
		if err := s.createRoleProfileTx(tx, user.UserID, roleID); err != nil {
			return errors.New("error al crear el perfil de rol")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(user.UserID, user.RoleID, false)
	if err != nil {
		return nil, errors.New("error generando token")
	}

	if err := s.saveRefreshToken(ctx, user.UserID, refreshToken); err != nil {
		return nil, errors.New("error al guardar sesión")
	}

	return &dto.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *authService) LoginWithGoogle(ctx context.Context, req *dto.GoogleLoginRequest) (resp *dto.LoginResponse, err error) {
	// try-catch idiomático en Go (panic recovery)
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error crítico inesperado: %v", r)
		}
	}()

	// 1. Verificar el token con Firebase
	fbToken, err := s.firebaseService.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		return nil, errors.New("token de google inválido")
	}

	email, ok := fbToken.Claims["email"].(string)
	if !ok || email == "" {
		return nil, errors.New("email no proporcionado por google")
	}

	// 2. Buscar usuario
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		// Si no existe, creamos un registro básico dentro de una transacción en el repositorio
		roleID := req.Role
		if !strings.HasPrefix(roleID, "role_") {
			roleID = "role_" + roleID
		}

		user = &models.User{
			UserID:       "gen-" + email, // En un caso real usarías un UUID
			Email:        email,
			GoogleAuthID: req.GoogleID,
			FirstName:    "Google",
			LastName:     "User",
			Status:       "active",
			RoleID:       roleID,
		}

		txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(user).Error; err != nil {
				return errors.New("error al crear el usuario mediante google")
			}
			if err := s.createRoleProfileTx(tx, user.UserID, roleID); err != nil {
				return errors.New("error al crear el perfil de rol mediante google")
			}
			return nil
		})
		if txErr != nil {
			return nil, txErr
		}
	}

	isExternal := false
	if user.RoleID == "role_teacher" {
		var isExt bool
		if err := s.db.WithContext(ctx).Table("teachers").Where("user_id = ?", user.UserID).Select("is_external").Scan(&isExt).Error; err == nil {
			isExternal = isExt
		}
	}

	// 3. Generar nuestros propios JWTs
	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(user.UserID, user.RoleID, isExternal)
	if err != nil {
		return nil, errors.New("error generando tokens internos")
	}

	if err := s.saveRefreshToken(ctx, user.UserID, refreshToken); err != nil {
		return nil, errors.New("error al guardar sesión")
	}

	approvalStatus := s.GetApprovalStatus(ctx, user.RoleID, user.UserID)

	// 4. Retornar respuesta
	return &dto.LoginResponse{
		Token:          accessToken, 
		RefreshToken:   refreshToken,
		User:           user,
		ApprovalStatus: approvalStatus,
	}, nil
}

func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	if req.Email == "" || req.Password == "" || req.Role == "" {
		return nil, errors.New("email, contraseña y rol son obligatorios")
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("credenciales inválidas")
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return nil, errors.New("credenciales inválidas")
	}

	roleID := req.Role
	if !strings.HasPrefix(roleID, "role_") {
		roleID = "role_" + roleID
	}

	if user.RoleID != roleID {
		return nil, errors.New("rol incorrecto para este usuario")
	}

	isExternal := false
	if user.RoleID == "role_teacher" {
		var isExt bool
		if err := s.db.WithContext(ctx).Table("teachers").Where("user_id = ?", user.UserID).Select("is_external").Scan(&isExt).Error; err == nil {
			isExternal = isExt
		}
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(user.UserID, user.RoleID, isExternal)
	if err != nil {
		return nil, errors.New("error generando token")
	}

	if err := s.saveRefreshToken(ctx, user.UserID, refreshToken); err != nil {
		return nil, errors.New("error al guardar sesión")
	}

	approvalStatus := s.GetApprovalStatus(ctx, user.RoleID, user.UserID)

	return &dto.LoginResponse{
		Token:          accessToken,
		RefreshToken:   refreshToken,
		User:           user,
		ApprovalStatus: approvalStatus,
	}, nil
}

func (s *authService) saveRefreshToken(ctx context.Context, userID, token string) error {
	tokenHash := utils.HashToken(token)
	rt := &models.RefreshToken{
		RefreshTokenID: "rt-" + utils.GenerateRandomString(16),
		UserID:         userID,
		TokenHash:      tokenHash,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour), // O leer de env
	}
	// Revocar tokens anteriores es una buena práctica (opcional, pero pedido implícitamente por rotación)
	_ = s.refreshTokenRepo.RevokeAllForUser(ctx, userID)

	// Limpieza en segundo plano de tokens expirados
	go func() {
		_ = s.refreshTokenRepo.DeleteExpired(context.Background())
	}()

	return s.refreshTokenRepo.Create(ctx, rt)
}

func (s *authService) createRoleProfileTx(tx *gorm.DB, userID, roleID string) error {
	status := "pending"
	switch roleID {
	case "role_company":
		return tx.Create(&models.Company{
			CompanyID:      "comp-" + utils.GenerateRandomString(10),
			UserID:         userID,
			ApprovalStatus: status,
		}).Error
	case "role_participant":
		return tx.Create(&models.Participant{
			ParticipantID:  "part-" + utils.GenerateRandomString(10),
			UserID:         userID,
			ApprovalStatus: status,
		}).Error
	case "role_teacher":
		return tx.Create(&models.Teacher{
			TeacherID:      "tech-" + utils.GenerateRandomString(10),
			UserID:         userID,
			ApprovalStatus: status,
		}).Error
	case "role_volunteer":
		return tx.Create(&models.Volunteer{
			VolunteerID:    "vol-" + utils.GenerateRandomString(10),
			UserID:         userID,
			ApprovalStatus: status,
		}).Error
	default:
		// Para role_admin, role_dev u otros roles que no requieren tabla extra
		return nil
	}
}

func (s *authService) GetApprovalStatus(ctx context.Context, roleID, userID string) string {
	var status string
	var tableName string

	switch roleID {
	case "role_teacher":
		tableName = "teachers"
	case "role_volunteer":
		tableName = "volunteers"
	case "role_participant":
		tableName = "participants"
	case "role_company":
		tableName = "companies"
	default:
		return "accepted"
	}

	if err := s.db.WithContext(ctx).Table(tableName).Where("user_id = ?", userID).Select("approval_status").Scan(&status).Error; err != nil || status == "" {
		return "accepted"
	}
	return status
}
