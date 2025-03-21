package repository

import (
	"errors"
	"fmt"
	"subsnotifpro-go/internal/google_playstore/settings/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PlaystoreSettingsRepository interface {
	SaveServiceAccount(tx *gorm.DB, appID, filePath, fileName string) (*models.GooglePlayServiceAccount, error)
	GetLatestServiceAccount(tx *gorm.DB, appID string) (models.GooglePlayServiceAccount, error)
	UpdateServiceAccountValidationStatus(tx *gorm.DB, appID string, isValid bool) error
	SoftDeleteAllServiceAccounts(tx *gorm.DB, appID string) error
	GetSettings(tx *gorm.DB, appID string) (models.GooglePlaySettings, error)
	UpsertSettings(tx *gorm.DB, settings *models.GooglePlaySettings) error
}

type playstoreSettingsRepository struct{}

func NewPlaystoreSettingsRepository() PlaystoreSettingsRepository {
	return &playstoreSettingsRepository{}
}

// ✅ Save a new service account and associate it with the given app
func (r *playstoreSettingsRepository) SaveServiceAccount(tx *gorm.DB, appID, filePath, fileName string) (*models.GooglePlayServiceAccount, error) {
	account := &models.GooglePlayServiceAccount{
		FileName: fileName,
		FilePath: filePath,
	}

	if err := tx.Create(account).Error; err != nil {
		return nil, fmt.Errorf("failed to save service account: %w", err)
	}

	// ✅ Upsert settings for this app with new service account
	settings := &models.GooglePlaySettings{
		AppID:                  appID,
		LatestServiceAccountID: &account.ID,
	}
	if err := r.UpsertSettings(tx, settings); err != nil {
		return nil, fmt.Errorf("failed to upsert settings for app %s: %w", appID, err)
	}

	return account, nil
}

// ✅ Get latest service account for a specific app
func (r *playstoreSettingsRepository) GetLatestServiceAccount(tx *gorm.DB, appID string) (models.GooglePlayServiceAccount, error) {
	var settings models.GooglePlaySettings
	if err := tx.Where("app_id = ?", appID).First(&settings).Error; err != nil {
		return models.GooglePlayServiceAccount{}, fmt.Errorf("settings not found for app: %w", err)
	}

	if settings.LatestServiceAccountID == nil {
		return models.GooglePlayServiceAccount{}, errors.New("no service account set for app")
	}

	var account models.GooglePlayServiceAccount
	if err := tx.Where("id = ?", *settings.LatestServiceAccountID).First(&account).Error; err != nil {
		return models.GooglePlayServiceAccount{}, fmt.Errorf("failed to fetch service account: %w", err)
	}

	return account, nil
}

// ✅ Update validation status of the latest service account for a specific app
func (r *playstoreSettingsRepository) UpdateServiceAccountValidationStatus(tx *gorm.DB, appID string, isValid bool) error {
	account, err := r.GetLatestServiceAccount(tx, appID)
	if err != nil {
		return err
	}

	account.Validated = isValid
	now := time.Now()
	account.LastChecked = &now

	if err := tx.Save(&account).Error; err != nil {
		return fmt.Errorf("failed to update service account validation: %w", err)
	}
	return nil
}

// ✅ Soft delete all service accounts for an app (you may want to filter by AppID in real-world case)
func (r *playstoreSettingsRepository) SoftDeleteAllServiceAccounts(tx *gorm.DB, appID string) error {
	// This deletes all — if you want to scope by AppID you'll need a relation
	return tx.Model(&models.GooglePlayServiceAccount{}).Where("deleted_at IS NULL").Delete(nil).Error
}

// ✅ Get settings by AppID
func (r *playstoreSettingsRepository) GetSettings(tx *gorm.DB, appID string) (models.GooglePlaySettings, error) {
	var settings models.GooglePlaySettings
	if err := tx.Where("app_id = ?", appID).First(&settings).Error; err != nil {
		return models.GooglePlaySettings{}, fmt.Errorf("settings not found for app: %w", err)
	}
	return settings, nil
}

// ✅ Upsert settings based on app_id (conflict resolution)
func (r *playstoreSettingsRepository) UpsertSettings(tx *gorm.DB, settings *models.GooglePlaySettings) error {
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "app_id"}}, // Unique key to detect conflict
		DoUpdates: clause.AssignmentColumns([]string{"latest_service_account_id"}),
	}).Create(settings).Error
}
