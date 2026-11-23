package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"auth/internal/models"
	"auth/internal/utils"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindAllByEmail(ctx context.Context, email string) ([]models.User, error)
	FindAllByGoogleID(ctx context.Context, googleID string) ([]models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindAll(ctx context.Context) ([]models.User, error)
	Create(ctx context.Context, user *models.User) error
	CreateWithRoleProfile(ctx context.Context, user *models.User, roleID string, approvedBy string) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
	CountActiveParticipants(ctx context.Context) (int64, error)
	CountVolunteers(ctx context.Context) (int64, error)
	// CountActiveStaff cuenta las cuentas activas y aprobadas de empresas,
	// docentes y voluntarios (para el Dashboard).
	CountActiveStaff(ctx context.Context) (empresas, docentes, voluntarios int64, err error)
	// FindParticipantsHistory devuelve todos los participantes, también
	// los dados de baja (soft delete), con sus fechas de alta,
	// aprobación y baja. Para el Dashboard.
	FindParticipantsHistory(ctx context.Context) ([]models.ParticipantHistory, error)
	UpdateEmailVerified(ctx context.Context, userID string, verified bool) error
	UpdateEmailVerifiedByEmail(ctx context.Context, email string, verified bool) error
	UpdatePasswordHashByEmail(ctx context.Context, email string, passwordHash string) error
	UpdatePasswordAndVerifiedByEmail(ctx context.Context, email string, passwordHash string, verified bool) error
	FindCampaignEligibleUsers(ctx context.Context) ([]CampaignEligibleUser, error)
	FindActiveUserIDsByRoles(ctx context.Context, roles []string) ([]string, error)
}

type userRepository struct {
	db *gorm.DB
}

const (
	queryUserID = "user_id = ?"
	queryEmail  = "email = ?"
)

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where(queryEmail, email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usuario no encontrado")
		}
		return nil, err
	}
	return &user, nil
}

// FindAllByEmail returns all user accounts associated with the given email (one per role in the current schema)
func (r *userRepository) FindAllByEmail(ctx context.Context, email string) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Where(queryEmail, email).Find(&users).Error; err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errors.New("usuario no encontrado")
	}
	return users, nil
}

func (r *userRepository) FindAllByGoogleID(ctx context.Context, googleID string) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).Where("google_auth_id = ?", googleID).Find(&users).Error; err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, errors.New("no se encontró ninguna cuenta vinculada a este usuario de Google")
	}
	return users, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).First(&user, queryUserID, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("usuario no encontrado")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindAll(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.*, COALESCE(p.approval_status, c.approval_status, t.approval_status, v.approval_status) AS role_approval_status, p.image_url AS image_url").
		Joins("LEFT JOIN participants p ON users.user_id = p.user_id").
		Joins("LEFT JOIN companies c ON users.user_id = c.user_id").
		Joins("LEFT JOIN teachers t ON users.user_id = t.user_id").
		Joins("LEFT JOIN volunteers v ON users.user_id = v.user_id").
		Preload("Role").
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	// Implementamos transacciones de GORM (equivalente a un try-catch seguro para DB)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Recovery en caso de panic dentro de la transacción
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				// Aquí podrías loguear el panic
			}
		}()

		if err := tx.Create(user).Error; err != nil {
			return err // GORM automáticamente hace Rollback si retornamos un error
		}

		return nil // Commit automático
	})
}

func (r *userRepository) CreateWithRoleProfile(ctx context.Context, user *models.User, roleID string, approvedBy string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		if err := tx.Create(user).Error; err != nil {
			return err
		}

		status := "approved"
		var appByPtr *string
		if approvedBy != "" {
			appByPtr = &approvedBy
		}

		switch roleID {
		case "role_company":
			return tx.Create(&models.Company{
				CompanyID:      "comp-" + utils.GenerateRandomString(10),
				UserID:         user.UserID,
				Cuit:           "PENDING-" + utils.GenerateRandomString(8),
				ApprovalStatus: status,
				ApprovedBy:     appByPtr,
			}).Error
		case "role_participant":
			dni, address := participantProfileFrom(ctx)
			return tx.Create(&models.Participant{
				ParticipantID:  "part-" + utils.GenerateRandomString(10),
				UserID:         user.UserID,
				DNI:            dni,
				Address:        address,
				ApprovalStatus: status,
				ApprovedBy:     appByPtr,
			}).Error
		case "role_teacher":
			return tx.Create(&models.Teacher{
				TeacherID:      "tech-" + utils.GenerateRandomString(10),
				UserID:         user.UserID,
				IsExternal:     teacherExternalFrom(ctx),
				ApprovalStatus: status,
				ApprovedBy:     appByPtr,
			}).Error
		case "role_volunteer":
			return tx.Create(&models.Volunteer{
				VolunteerID:    "vol-" + utils.GenerateRandomString(10),
				UserID:         user.UserID,
				ApprovalStatus: status,
				ApprovedBy:     appByPtr,
			}).Error
		}

		return nil
	})
}

func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Save(user).Error
	})
}

func (r *userRepository) UpdateEmailVerified(ctx context.Context, userID string, verified bool) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where(queryUserID, userID).
		Update("is_email_verified", verified).Error
}

func (r *userRepository) UpdateEmailVerifiedByEmail(ctx context.Context, email string, verified bool) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where(queryEmail, email).
		Update("is_email_verified", verified).Error
}

func (r *userRepository) UpdatePasswordHashByEmail(ctx context.Context, email string, passwordHash string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where(queryEmail, email).
		Update("password_hash", passwordHash).Error
}

func (r *userRepository) UpdatePasswordAndVerifiedByEmail(ctx context.Context, email string, passwordHash string, verified bool) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where(queryEmail, email).
		Updates(map[string]interface{}{
			"password_hash":     passwordHash,
			"is_email_verified": verified,
		}).Error
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		// Realiza soft-delete automáticamente debido a DeletedAt en el modelo
		return tx.Where(queryUserID, id).Delete(&models.User{}).Error
	})
}

func (r *userRepository) countAcceptedUsersByRole(ctx context.Context, roleID, profileTable string) (int64, error) {
	var count int64
	var altRoleID string
	if strings.HasPrefix(roleID, "role_") {
		altRoleID = strings.TrimPrefix(roleID, "role_")
	} else {
		altRoleID = "role_" + roleID
	}

	joinClause := fmt.Sprintf("JOIN %s ON %s.user_id = users.user_id", profileTable, profileTable)
	// "accepted" es el valor que se guardaba antes de unificar en
	// "approved" (ver services/approval_status.go): hay filas reales así.
	whereClause := fmt.Sprintf("(users.role_id = ? OR users.role_id = ?) AND users.status = ? AND %s.approval_status IN ?", profileTable)

	err := r.db.WithContext(ctx).
		Table("users").
		Joins(joinClause).
		Where(whereClause, roleID, altRoleID, "active", []string{"approved", "accepted"}).
		Where("users.deleted_at IS NULL").
		Count(&count).Error
	return count, err
}

func (r *userRepository) CountActiveParticipants(ctx context.Context) (int64, error) {
	return r.countAcceptedUsersByRole(ctx, "role_participant", "participants")
}

func (r *userRepository) CountVolunteers(ctx context.Context) (int64, error) {
	return r.countAcceptedUsersByRole(ctx, "role_volunteer", "volunteers")
}

func (r *userRepository) CountActiveStaff(ctx context.Context) (int64, int64, int64, error) {
	empresas, err := r.countAcceptedUsersByRole(ctx, "role_company", "companies")
	if err != nil {
		return 0, 0, 0, err
	}
	docentes, err := r.countAcceptedUsersByRole(ctx, "role_teacher", "teachers")
	if err != nil {
		return 0, 0, 0, err
	}
	voluntarios, err := r.CountVolunteers(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	return empresas, docentes, voluntarios, nil
}

func (r *userRepository) FindParticipantsHistory(ctx context.Context) ([]models.ParticipantHistory, error) {
	var rows []models.ParticipantHistory
	// Raw a propósito: hay que ver también a los usuarios borrados
	// (deleted_at no nulo), que GORM filtra en las consultas normales.
	err := r.db.WithContext(ctx).Raw(`
		SELECT u.user_id, u.status, u.created_at, u.updated_at, u.deleted_at,
		       p.approval_status, p.approval_date
		FROM users u
		JOIN participants p ON p.user_id = u.user_id
		WHERE u.role_id IN ('role_participant', 'participant')
		ORDER BY u.created_at`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

type CampaignEligibleUser struct {
	UserID          string `json:"user_id"`
	Email           string `json:"email"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	IsEmailVerified bool   `json:"is_email_verified"`
}

func (r *userRepository) FindCampaignEligibleUsers(ctx context.Context) ([]CampaignEligibleUser, error) {
	var results []CampaignEligibleUser
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.user_id, users.email, users.first_name, users.last_name, users.is_email_verified").
		Joins("LEFT JOIN teachers t ON users.user_id = t.user_id AND t.deleted_at IS NULL").
		Where("users.deleted_at IS NULL").
		Where("users.status = ?", "active").
		Where("users.role_id != ?", "role_company").
		Where("NOT (users.role_id = ? AND t.is_external = ?)", "role_teacher", true).
		Scan(&results).Error
	return results, err
}

func (r *userRepository) FindActiveUserIDsByRoles(ctx context.Context, roles []string) ([]string, error) {
	if len(roles) == 0 {
		return []string{}, nil
	}
	var userIDs []string
	query := r.db.WithContext(ctx).
		Table("users").
		Select("users.user_id").
		Joins("LEFT JOIN volunteers v ON users.user_id = v.user_id AND v.deleted_at IS NULL").
		Where("users.deleted_at IS NULL").
		Where("users.status = ?", "active").
		Where("users.role_id IN ?", roles).
		Where("(users.role_id != 'role_volunteer' OR v.approval_status = 'approved')")

	err := query.Pluck("users.user_id", &userIDs).Error
	if err != nil {
		return nil, err
	}
	return userIDs, nil
}


type participantProfileKey struct{}

type participantProfile struct{ dni, address string }

// WithParticipantProfile pasa el DNI y la dirección del participante a
// CreateWithRoleProfile, para guardarlos en la misma transacción que la
// cuenta (HU-07) sin cambiar la firma del repositorio.
func WithParticipantProfile(ctx context.Context, dni, address string) context.Context {
	return context.WithValue(ctx, participantProfileKey{}, participantProfile{dni: dni, address: address})
}

func participantProfileFrom(ctx context.Context) (string, string) {
	p, _ := ctx.Value(participantProfileKey{}).(participantProfile)
	return p.dni, p.address
}

type teacherExternalKey struct{}

// WithTeacherExternal marca que el docente que se crea es externo (sube
// material a la biblioteca pública). Usa la columna is_external que ya
// existe en teachers: no cambia la estructura de la base.
func WithTeacherExternal(ctx context.Context, external bool) context.Context {
	return context.WithValue(ctx, teacherExternalKey{}, external)
}

func teacherExternalFrom(ctx context.Context) bool {
	v, _ := ctx.Value(teacherExternalKey{}).(bool)
	return v
}
