package repository

import (
	"subsnotifpro-go/internal/auth/models"

	"gorm.io/gorm"
)

type IAdminRepository interface {
	Create(tx *gorm.DB, admin *models.AdminUser) error
	FindByEmail(tx *gorm.DB, email string) (*models.AdminUser, error)
}

type adminRepository struct{}

func NewAdminRepository() IAdminRepository {
	return &adminRepository{}
}

func (r *adminRepository) Create(tx *gorm.DB, admin *models.AdminUser) error {
	return tx.Create(admin).Error
}

func (r *adminRepository) FindByEmail(tx *gorm.DB, email string) (*models.AdminUser, error) {
	var admin models.AdminUser
	if err := tx.Where("email = ?", email).First(&admin).Error; err != nil {
		return nil, err
	}
	return &admin, nil
}
