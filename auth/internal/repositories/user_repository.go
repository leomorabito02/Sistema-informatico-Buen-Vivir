package repositories

import (
	"context"
	"errors"
	"fmt"

	"auth/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindAllByEmail(ctx context.Context, email string) ([]models.User, error)
	FindAllByGoogleID(ctx context.Context, googleID string) ([]models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindAll(ctx context.Context) ([]models.User, error)
	Create(ctx context.Context, user *models.User) error
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
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
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

func (r *userRepository) countApprovedUsersByRole(ctx context.Context, roleID, profileTable string) (int64, error) {
	var count int64
	joinClause := fmt.Sprintf("LEFT JOIN %s ON %s.user_id = users.user_id", profileTable, profileTable)
	whereApproval := fmt.Sprintf("users.role_id = ? AND users.status = ? AND (%s.approval_status = ? OR %s.approval_status IS NULL)", profileTable, profileTable)

	err := r.db.WithContext(ctx).
		Table("users").
		Joins(joinClause).
		Where(whereApproval, roleID, "active", "approved").
		Where("users.deleted_at IS NULL").
		Count(&count).Error
	return count, err
}

func (r *userRepository) CountActiveParticipants(ctx context.Context) (int64, error) {
	return r.countApprovedUsersByRole(ctx, "participant", "participants")
}

func (r *userRepository) CountVolunteers(ctx context.Context) (int64, error) {
	return r.countApprovedUsersByRole(ctx, "volunteer", "volunteers")
}
