package repository

import (
	"subsnotifpro-go/internal/tenant/models"

	"gorm.io/gorm"
)

type IAppRepository interface {
	Create(tx *gorm.DB, app *models.App) error
	ListByTenantID(tx *gorm.DB, tenantID string) ([]models.App, error)
	GetByPackageName(tx *gorm.DB, packageName string) (*models.App, error) // 🆕
}

type appRepository struct{}

func NewAppRepository() IAppRepository {
	return &appRepository{}
}

func (r *appRepository) Create(tx *gorm.DB, app *models.App) error {
	return tx.Create(app).Error
}

func (r *appRepository) ListByTenantID(tx *gorm.DB, tenantID string) ([]models.App, error) {
	var apps []models.App
	if err := tx.Where("tenant_id = ?", tenantID).Find(&apps).Error; err != nil {
		return nil, err
	}
	return apps, nil
}

func (r *appRepository) GetByPackageName(tx *gorm.DB, packageName string) (*models.App, error) {
	var app models.App
	if err := tx.Where("package_name = ?", packageName).First(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}
