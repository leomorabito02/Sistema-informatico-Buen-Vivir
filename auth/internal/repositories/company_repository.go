package repositories

import (
	"context"
	"errors"

	"auth/internal/models"
	"gorm.io/gorm"
)

type CompanyRepository interface {
	FindByID(ctx context.Context, id string) (*models.Company, error)
	FindAll(ctx context.Context) ([]models.Company, error)
	Create(ctx context.Context, company *models.Company) error
	Update(ctx context.Context, company *models.Company) error
	Delete(ctx context.Context, id string) error
}

type companyRepository struct {
	db *gorm.DB
}

func NewCompanyRepository(db *gorm.DB) CompanyRepository {
	return &companyRepository{db: db}
}

// FindByID acepta el id de la empresa o el de su usuario (igual que
// participantRepository.FindByID): Inserción Laboral guarda en cada oferta
// el id de usuario de la empresa (X-User-ID) y la consulta por gRPC con ese.
func (r *companyRepository) FindByID(ctx context.Context, id string) (*models.Company, error) {
	var company models.Company
	if err := r.db.WithContext(ctx).Where("company_id = ?", id).Or("user_id = ?", id).First(&company).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("empresa no encontrada")
		}
		return nil, err
	}
	return &company, nil
}

func (r *companyRepository) FindAll(ctx context.Context) ([]models.Company, error) {
	var companies []models.Company
	if err := r.db.WithContext(ctx).Find(&companies).Error; err != nil {
		return nil, err
	}
	return companies, nil
}

func (r *companyRepository) Create(ctx context.Context, company *models.Company) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Create(company).Error
	})
}

func (r *companyRepository) Update(ctx context.Context, company *models.Company) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Save(company).Error
	})
}

func (r *companyRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()
		return tx.Where("company_id = ?", id).Delete(&models.Company{}).Error
	})
}
