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
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
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
	if err := r.db.WithContext(ctx).Where("email = ?", email).Find(&users).Error; err != nil {
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
	if err := r.db.WithContext(ctx).First(&user, "user_id = ?", id).Error; err != nil {
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
		Select("users.*, COALESCE(p.approval_status, c.approval_status, t.approval_status, v.approval_status) AS role_approval_status").
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
			return tx.Create(&models.Participant{
				ParticipantID:  "part-" + utils.GenerateRandomString(10),
				UserID:         user.UserID,
				ApprovalStatus: status,
				ApprovedBy:     appByPtr,
			}).Error
		case "role_teacher":
			return tx.Create(&models.Teacher{
				TeacherID:      "tech-" + utils.GenerateRandomString(10),
				UserID:         user.UserID,
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

func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		// Realiza soft-delete automáticamente debido a DeletedAt en el modelo
		return tx.Where("user_id = ?", id).Delete(&models.User{}).Error
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
	whereClause := fmt.Sprintf("(users.role_id = ? OR users.role_id = ?) AND users.status = ? AND %s.approval_status = ?", profileTable)

	err := r.db.WithContext(ctx).
		Table("users").
		Joins(joinClause).
		Where(whereClause, roleID, altRoleID, "active", "approved").
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
