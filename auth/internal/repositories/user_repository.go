package repositories

import (
	"context"
	"errors"

	"auth/internal/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindAllByEmail(ctx context.Context, email string) ([]models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindAll(ctx context.Context) ([]models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id string) error
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
