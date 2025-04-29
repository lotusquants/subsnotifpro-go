package repository

import (
	"subsnotifpro-go/internal/appstore/settings/models"

	"gorm.io/gorm"
)

type AppStoreSettingsRepository interface {
	GetSettings(bundleID string) (*models.AppStoreSettings, error)
	GetAllSettings() ([]models.AppStoreSettings, error)
	CreateOrUpdateSettings(settings *models.AppStoreSettings) (*models.AppStoreSettings, error)
	DeleteSettings(bundleID string) error
}

type appStoreSettingsRepository struct {
	db *gorm.DB
}

func NewAppStoreSettingsRepository(db *gorm.DB) AppStoreSettingsRepository {
	return &appStoreSettingsRepository{db: db}
}

func (r *appStoreSettingsRepository) GetSettings(bundleID string) (*models.AppStoreSettings, error) {
	var settings models.AppStoreSettings
	result := r.db.Where("bundle_id = ?", bundleID).First(&settings)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &settings, nil
}

func (r *appStoreSettingsRepository) GetAllSettings() ([]models.AppStoreSettings, error) {
	var settings []models.AppStoreSettings
	result := r.db.Find(&settings)
	if result.Error != nil {
		return nil, result.Error
	}
	return settings, nil
}

func (r *appStoreSettingsRepository) CreateOrUpdateSettings(settings *models.AppStoreSettings) (*models.AppStoreSettings, error) {
	var existing models.AppStoreSettings

	// Check if settings exist for this bundle ID
	err := r.db.Where("bundle_id = ?", settings.BundleID).First(&existing).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if err == gorm.ErrRecordNotFound {
		// Create new settings
		if err := r.db.Create(settings).Error; err != nil {
			return nil, err
		}
		return settings, nil
	}

	// Update existing settings
	settings.ID = existing.ID
	if err := r.db.Save(settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *appStoreSettingsRepository) DeleteSettings(bundleID string) error {
	result := r.db.Where("bundle_id = ?", bundleID).Delete(&models.AppStoreSettings{})
	return result.Error
}
