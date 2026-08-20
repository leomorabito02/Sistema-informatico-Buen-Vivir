package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"auth/internal/dto"
	"auth/internal/models"
	"auth/internal/repositories"
	"auth/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.LoginResponse, error)
	CheckCredentials(ctx context.Context, req *dto.CheckCredentialsRequest) (*dto.CheckCredentialsResponse, error)
	CheckGoogleCredentials(ctx context.Context, req *dto.GoogleLoginRequest) (*dto.CheckCredentialsResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error)
	GetApprovalStatus(ctx context.Context, roleID, userID string) string
	UpdateMyProfile(ctx context.Context, roleID, userID string, data map[string]interface{}) error
	GetMyProfile(ctx context.Context, roleID, userID string) (map[string]interface{}, error)
	UpdateMyPassword(ctx context.Context, userID, oldPassword, newPassword string) error
	LinkGoogleAccount(ctx context.Context, userID string, req *dto.LinkGoogleRequest) error
	UnlinkGoogleAccount(ctx context.Context, userID string) error
}

type authService struct {
	transactor       repositories.Transactor
	querier          repositories.RoleProfileQuerier
	userRepo         repositories.UserRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	jwtService       JWTService
	firebaseService  FirebaseService
}

func NewAuthService(
	transactor repositories.Transactor,
	querier repositories.RoleProfileQuerier,
	userRepo repositories.UserRepository,
	rtRepo repositories.RefreshTokenRepository,
	jwtService JWTService,
	fbService FirebaseService,
) AuthService {
	return &authService{
		transactor:       transactor,
		querier:          querier,
		userRepo:         userRepo,
		refreshTokenRepo: rtRepo,
		jwtService:       jwtService,
		firebaseService:  fbService,
	}
}

func (s *authService) Register(ctx context.Context, req *dto.RegisterRequest) (resp *dto.LoginResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error crítico inesperado: %v", r)
		}
	}()

	if req.Email == "" || req.Password == "" || req.Role == "" {
		return nil, errors.New("el correo electrónico, la contraseña y el rol son obligatorios")
	}

	if err := utils.ValidatePasswordComplexity(req.Password); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && user != nil {
		return nil, errors.New("ya existe una cuenta registrada con este correo electrónico")
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

	err = s.transactor.RunInTx(ctx, func(tx repositories.Tx) error {
		if err := tx.Create(user); err != nil {
			return errors.New("error al guardar los datos del usuario")
		}
		if err := s.createRoleProfileTx(tx, user.UserID, roleID); err != nil {
			return errors.New("error al crear el perfil asociado al rol seleccionado")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.jwtService.GenerateTokenPair(user.UserID, user.RoleID, false)
	if err != nil {
		return nil, errors.New("error al generar la sesión de autenticación")
	}

	if err := s.saveRefreshToken(ctx, user.UserID, refreshToken); err != nil {
		return nil, errors.New("error al guardar el estado de la sesión")
	}

	return &dto.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

// CheckCredentials validates email and password and returns all accounts associated with the email.
// No token is issued at this step.
func (s *authService) CheckCredentials(ctx context.Context, req *dto.CheckCredentialsRequest) (resp *dto.CheckCredentialsResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error crítico inesperado: %v", r)
		}
	}()

	if req.Email == "" || req.Password == "" {
		return nil, errors.New("el correo electrónico y la contraseña son obligatorios")
	}

	users, err := s.userRepo.FindAllByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("no existe ninguna cuenta registrada con este correo electrónico")
	}

	// Validate password against the first account (all accounts with the same email share the same password hash)
	if !utils.CheckPasswordHash(req.Password, users[0].PasswordHash) {
		slog.Warn("security: failed login attempt", slog.String("event", "invalid_password"), slog.String("email_hash", utils.HashToken(req.Email)))
		return nil, errors.New("la contraseña ingresada es incorrecta")
	}

	roleNames := map[string]string{
		"role_admin":       "Responsable",
		"role_dev":         "Desarrollador",
		"role_participant": "Participante",
		"role_company":     "Empresa",
		"role_teacher":     "Docente",
		"role_volunteer":   "Voluntario",
	}

	accounts := make([]dto.AvailableAccount, 0, len(users))
	for _, u := range users {
		roleName := roleNames[u.RoleID]
		if roleName == "" {
			roleName = u.RoleID
		}
		accounts = append(accounts, dto.AvailableAccount{
			UserID:   u.UserID,
			RoleID:   u.RoleID,
			RoleName: roleName,
			Status:   u.Status,
		})
	}

	userIDs := make([]string, 0, len(accounts))
	for _, a := range accounts {
		userIDs = append(userIDs, a.UserID)
	}
	loginNonce, err := utils.GetNonceStore().Generate(userIDs, 2*time.Minute)
	if err != nil {
		return nil, errors.New("error generando sesión temporal, intente nuevamente")
	}

	return &dto.CheckCredentialsResponse{Accounts: accounts, LoginNonce: loginNonce}, nil
}

// CheckGoogleCredentials validates a Firebase ID token and returns available accounts linked to the Google account.
// Google login requires that an account with the email exists AND has a matching google_auth_id linked.
func (s *authService) CheckGoogleCredentials(ctx context.Context, req *dto.GoogleLoginRequest) (resp *dto.CheckCredentialsResponse, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error crítico inesperado: %v", r)
		}
	}()

	fbToken, err := s.firebaseService.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		return nil, errors.New("el token de autenticación de Google es inválido o expiró")
	}

	email, ok := fbToken.Claims["email"].(string)
	if !ok || email == "" {
		return nil, errors.New("no se pudo obtener la dirección de correo electrónico desde Google")
	}

	googleAuthID := strings.TrimSpace(req.GoogleID)
	if googleAuthID == "" {
		googleAuthID = fbToken.UID
	}

	// 1. Buscar usuarios por google_auth_id directamente en la base de datos
	matchingUsers, err := s.userRepo.FindAllByGoogleID(ctx, googleAuthID)
	if err != nil || len(matchingUsers) == 0 {
		matchingUsers, _ = s.userRepo.FindAllByGoogleID(ctx, fbToken.UID)
	}

	// 2. Si no se encuentra por google_auth_id, comprobar por correo electrónico para reportar error preciso
	if len(matchingUsers) == 0 {
		users, _ := s.userRepo.FindAllByEmail(ctx, email)
		if len(users) == 0 {
			return nil, errors.New("no existe ninguna cuenta registrada con este correo electrónico. Debe registrarse primero")
		}
		slog.Warn("security: google login denied — google_auth_id not linked", slog.String("event", "google_auth_not_linked"), slog.String("email_hash", utils.HashToken(email)))
		return nil, errors.New("la cuenta de Google no está vinculada a este usuario. Debe vincularla desde la configuración de su cuenta")
	}

	roleNames := map[string]string{
		"role_admin":       "Responsable",
		"role_dev":         "Desarrollador",
		"role_participant": "Participante",
		"role_company":     "Empresa",
		"role_teacher":     "Docente",
		"role_volunteer":   "Voluntario",
	}

	accounts := make([]dto.AvailableAccount, 0, len(matchingUsers))
	for _, u := range matchingUsers {
		roleName := roleNames[u.RoleID]
		if roleName == "" {
			roleName = u.RoleID
		}
		accounts = append(accounts, dto.AvailableAccount{
			UserID:   u.UserID,
			RoleID:   u.RoleID,
			RoleName: roleName,
			Status:   u.Status,
		})
	}

	googleUserIDs := make([]string, 0, len(matchingUsers))
	for _, u := range matchingUsers {
		googleUserIDs = append(googleUserIDs, u.UserID)
	}
	googleNonce, nonceErr := utils.GetNonceStore().Generate(googleUserIDs, 2*time.Minute)
	if nonceErr != nil {
		return nil, errors.New("error generando sesión temporal, intente nuevamente")
	}

	return &dto.CheckCredentialsResponse{Accounts: accounts, LoginNonce: googleNonce}, nil
}

// Login is step 2: given a user_id, role_id and a valid login nonce, issue the JWT token.
func (s *authService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, error) {
	if req.UserID == "" || req.RoleID == "" {
		return nil, errors.New("el identificador de usuario y el rol son obligatorios")
	}

	// Validate and consume the single-use login nonce (A04: prevents Step 2 bypass)
	if !utils.GetNonceStore().Consume(req.LoginNonce, req.UserID) {
		slog.Warn("security: login step 2 rejected — invalid or expired nonce",
			slog.String("event", "invalid_login_nonce"),
			slog.String("user_id", req.UserID),
		)
		return nil, errors.New("la sesión de verificación expiró o es inválida. Por favor, iniciá sesión nuevamente")
	}

	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, errors.New("el usuario especificado no fue encontrado")
	}

	if user.RoleID != req.RoleID {
		slog.Warn("security: role mismatch on login step 2", slog.String("event", "role_mismatch"), slog.String("user_id", req.UserID), slog.String("requested_role", req.RoleID))
		return nil, errors.New("el rol seleccionado no corresponde a este usuario")
	}

	statusLower := strings.ToLower(strings.TrimSpace(user.Status))
	if statusLower != "active" {
		slog.Warn("security: login attempt on blocked account", slog.String("event", "account_blocked"), slog.String("user_id", user.UserID), slog.String("role_id", user.RoleID))
		return &dto.LoginResponse{
			User:       user,
			UserStatus: "blocked",
		}, nil
	}

	approvalStatus := s.GetApprovalStatus(ctx, user.RoleID, user.UserID)
	if user.RoleID != "role_admin" && user.RoleID != "role_dev" {
		if approvalStatus != "active" && approvalStatus != "accepted" && approvalStatus != "pending" {
			return nil, errors.New("el perfil aún no está aprobado")
		}
	}

	isExternal := false
	if user.RoleID == "role_teacher" {
		isExt, err := s.querier.GetIsExternal(ctx, user.UserID)
		if err == nil {
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

	return &dto.LoginResponse{
		Token:          accessToken,
		RefreshToken:   refreshToken,
		User:           user,
		ApprovalStatus: approvalStatus,
		UserStatus:     user.Status,
	}, nil
}

func (s *authService) saveRefreshToken(ctx context.Context, userID, token string) error {
	tokenHash := utils.HashToken(token)
	rt := &models.RefreshToken{
		RefreshTokenID: "rt-" + utils.GenerateRandomString(16),
		UserID:         userID,
		TokenHash:      tokenHash,
		ExpiresAt:      time.Now().Add(7 * 24 * time.Hour),
	}
	_ = s.refreshTokenRepo.RevokeAllForUser(ctx, userID)

	go func() {
		_ = s.refreshTokenRepo.DeleteExpired(context.Background())
	}()

	return s.refreshTokenRepo.Create(ctx, rt)
}

func (s *authService) createRoleProfileTx(tx repositories.Tx, userID, roleID string) error {
	status := "pending"
	switch roleID {
	case "role_company":
		return tx.Create(&models.Company{
			CompanyID:      "comp-" + utils.GenerateRandomString(10),
			UserID:         userID,
			ApprovalStatus: status,
		})
	case "role_participant":
		return tx.Create(&models.Participant{
			ParticipantID:  "part-" + utils.GenerateRandomString(10),
			UserID:         userID,
			ApprovalStatus: status,
		})
	case "role_teacher":
		return tx.Create(&models.Teacher{
			TeacherID:      "tech-" + utils.GenerateRandomString(10),
			UserID:         userID,
			ApprovalStatus: status,
		})
	case "role_volunteer":
		return tx.Create(&models.Volunteer{
			VolunteerID:    "vol-" + utils.GenerateRandomString(10),
			UserID:         userID,
			ApprovalStatus: status,
		})
	default:
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

	status, err := s.querier.GetApprovalStatus(ctx, tableName, userID)
	if err != nil || status == "" {
		return "accepted"
	}
	return status
}

func (s *authService) GetMyProfile(ctx context.Context, roleID, userID string) (map[string]interface{}, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("usuario no encontrado")
	}

	profile := map[string]interface{}{
		"user": map[string]interface{}{
			"user_id":        user.UserID,
			"first_name":     user.FirstName,
			"last_name":      user.LastName,
			"email":          user.Email,
			"phone":          user.Phone,
			"created_at":     user.CreatedAt,
			"role_id":        user.RoleID,
			"status":         user.Status,
			"google_auth_id": user.GoogleAuthID,
		},
	}

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
	}

	if tableName != "" {
		roleData, err := s.querier.GetRoleData(ctx, tableName, userID)
		if err == nil {
			profile["role_data"] = roleData
		}
	}

	return profile, nil
}

func (s *authService) UpdateMyPassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("usuario no encontrado")
	}

	if user.PasswordHash == "" {
		return errors.New("la cuenta no tiene contraseña configurada (quizás fue creada con Google)")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("la contraseña actual es incorrecta")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("error al encriptar la nueva contraseña")
	}

	user.PasswordHash = string(hashedPassword)
	if err := s.userRepo.Update(ctx, user); err != nil {
		return errors.New("error al guardar la nueva contraseña")
	}

	return nil
}

func (s *authService) LinkGoogleAccount(ctx context.Context, userID string, req *dto.LinkGoogleRequest) error {
	fbToken, err := s.firebaseService.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		return errors.New("el token de autenticación de Google es inválido o expiró")
	}

	googleAuthID := strings.TrimSpace(req.GoogleID)
	if googleAuthID == "" {
		googleAuthID = fbToken.UID
	}

	currentUser, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("usuario no encontrado")
	}

	allUsers, _ := s.userRepo.FindAll(ctx)
	for _, u := range allUsers {
		if u.GoogleAuthID == googleAuthID && u.Email != currentUser.Email {
			return errors.New("esta cuenta de Google ya está vinculada a otro usuario")
		}
	}

	users, err := s.userRepo.FindAllByEmail(ctx, currentUser.Email)
	if err != nil {
		return errors.New("error al obtener las cuentas del usuario")
	}

	for i := range users {
		users[i].GoogleAuthID = googleAuthID
		if err := s.userRepo.Update(ctx, &users[i]); err != nil {
			return errors.New("error al vincular la cuenta de Google")
		}
	}

	return nil
}

func (s *authService) UnlinkGoogleAccount(ctx context.Context, userID string) error {
	currentUser, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return errors.New("usuario no encontrado")
	}

	users, err := s.userRepo.FindAllByEmail(ctx, currentUser.Email)
	if err != nil {
		return errors.New("error al obtener las cuentas del usuario")
	}

	hasPassword := false
	for _, u := range users {
		if strings.TrimSpace(u.PasswordHash) != "" {
			hasPassword = true
			break
		}
	}

	if !hasPassword {
		return errors.New("no podés desvincular la cuenta de Google porque no tenés una contraseña configurada")
	}

	for i := range users {
		users[i].GoogleAuthID = ""
		if err := s.userRepo.Update(ctx, &users[i]); err != nil {
			return errors.New("error al desvincular la cuenta de Google")
		}
	}

	return nil
}
