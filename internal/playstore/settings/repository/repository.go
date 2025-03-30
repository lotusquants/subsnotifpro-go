package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"subsnotifpro-go/internal/playstore/settings/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrFileDelete     = errors.New("failed to delete file")
)

type PlaystoreSettingsRepository interface {
	// Save or update package settings
	SavePackageSettings(ctx context.Context, tx *gorm.DB, settings *models.GooglePlaySettings) error

	// Upload or replace service account for a package
	UploadServiceAccount(ctx context.Context, tx *gorm.DB, serviceAccount *models.GooglePlayServiceAccount) error

	// Get settings by package name
	GetSettingsByPackageName(ctx context.Context, tx *gorm.DB, packageName string) (*models.GooglePlaySettings, error)

	// Get the service account by package name
	GetServiceAccount(ctx context.Context, tx *gorm.DB, packageName string) (*models.GooglePlayServiceAccount, error)

	// Update service account validation status
	UpdateServiceAccountValidationStatus(ctx context.Context, tx *gorm.DB, serviceAccountID string, isValid bool) error

	// Delete package settings (soft delete)
	DeletePackageSettings(ctx context.Context, tx *gorm.DB, packageName string) error
}

type playstoreSettingsRepository struct {
	db *gorm.DB
}

func NewPlaystoreSettingsRepository(db *gorm.DB) PlaystoreSettingsRepository {
	return &playstoreSettingsRepository{db: db}
}

// -------------------------
// 🚀 Save or Update Package Settings
// -------------------------
func (r *playstoreSettingsRepository) SavePackageSettings(ctx context.Context, tx *gorm.DB, settings *models.GooglePlaySettings) error {
	// Upsert operation: Update if package name exists, else create
	if err := tx.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "package_name"}},                         // Conflict target
		DoUpdates: clause.AssignmentColumns([]string{"latest_service_account_id"}), // Columns to update
	}).Create(settings).Error; err != nil {
		return fmt.Errorf("failed to save or update package settings for %s: %w", settings.PackageName, err)
	}
	return nil
}

// -------------------------
// 🚀 Upload or Replace Service Account
// -------------------------
func (r *playstoreSettingsRepository) UploadServiceAccount(ctx context.Context, tx *gorm.DB, serviceAccount *models.GooglePlayServiceAccount) error {
	// Step 1: Soft delete any existing service account for the package
	var existingServiceAccount models.GooglePlayServiceAccount
	err := tx.WithContext(ctx).
		Where("package_name = ?", serviceAccount.PackageName).
		Order("created_at desc").
		First(&existingServiceAccount).Error

	if err == nil {
		// Soft delete the previous account
		if err := tx.WithContext(ctx).Delete(&existingServiceAccount).Error; err != nil {
			return fmt.Errorf("failed to soft delete existing service account for package %s: %w", serviceAccount.PackageName, err)
		}

		// Remove old file
		if _, err := os.Stat(existingServiceAccount.FilePath); err == nil {
			if err := os.Remove(existingServiceAccount.FilePath); err != nil && !os.IsNotExist(err) {
				log.Printf("⚠️ Error deleting old service account file: %v", err)
			}
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to query existing service account for package %s: %w", serviceAccount.PackageName, err)
	}

	// Step 2: Save the new service account
	if err := tx.WithContext(ctx).Create(serviceAccount).Error; err != nil {
		return fmt.Errorf("failed to save new service account for package %s: %w", serviceAccount.PackageName, err)
	}

	// Step 3: Try to update GooglePlaySettings with new service account ID
	updateResult := tx.WithContext(ctx).
		Model(&models.GooglePlaySettings{}).
		Where("package_name = ? AND deleted_at IS NULL", serviceAccount.PackageName).
		Update("latest_service_account_id", serviceAccount.ID)

	if updateResult.Error != nil {
		return fmt.Errorf("failed to update latest_service_account_id for package %s: %w", serviceAccount.PackageName, updateResult.Error)
	}

	// Step 4: If settings not found (was deleted or missing), insert fresh settings
	// Step 4: If settings not found, try to restore soft-deleted record before insert
	if updateResult.RowsAffected == 0 {
		var deletedSetting models.GooglePlaySettings
		err := tx.WithContext(ctx).
			Unscoped(). // ✅ allow searching soft-deleted records
			Where("package_name = ?", serviceAccount.PackageName).
			First(&deletedSetting).Error

		if err == nil {
			// ✅ Restore the soft-deleted settings
			if err := tx.Model(&models.GooglePlaySettings{}).
				Unscoped().
				Where("id = ?", deletedSetting.ID).
				Update("deleted_at", nil).Error; err != nil {
				return fmt.Errorf("failed to restore soft-deleted settings for package %s: %w", serviceAccount.PackageName, err)
			}

			// ✅ Update the restored setting with new service account ID
			if err := tx.Model(&models.GooglePlaySettings{}).
				Where("id = ?", deletedSetting.ID).
				Update("latest_service_account_id", serviceAccount.ID).Error; err != nil {
				return fmt.Errorf("failed to update restored settings for package %s: %w", serviceAccount.PackageName, err)
			}
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			// No existing record at all — create new
			newSettings := &models.GooglePlaySettings{
				PackageName:            serviceAccount.PackageName,
				LatestServiceAccountID: &serviceAccount.ID,
			}
			if err := tx.WithContext(ctx).Create(newSettings).Error; err != nil {
				return fmt.Errorf("failed to create GooglePlaySettings for package %s: %w", serviceAccount.PackageName, err)
			}
		} else {
			return fmt.Errorf("error checking soft-deleted settings for package %s: %w", serviceAccount.PackageName, err)
		}
	}

	return nil
}

// -------------------------
// 🚀 Get Settings by Package Name
// -------------------------
func (r *playstoreSettingsRepository) GetSettingsByPackageName(ctx context.Context, tx *gorm.DB, packageName string) (*models.GooglePlaySettings, error) {
	var settings models.GooglePlaySettings
	err := tx.WithContext(ctx).Where("package_name = ?", packageName).First(&settings).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no settings found for package %s: %w", packageName, ErrRecordNotFound)
	}
	return &settings, err
}

// -------------------------
// 🚀 Get Service Account by Package Name
// -------------------------
func (r *playstoreSettingsRepository) GetServiceAccount(ctx context.Context, tx *gorm.DB, packageName string) (*models.GooglePlayServiceAccount, error) {
	var serviceAccount models.GooglePlayServiceAccount
	err := tx.WithContext(ctx).Where("package_name = ?", packageName).Order("created_at desc").First(&serviceAccount).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("no service account found for package %s: %w", packageName, ErrRecordNotFound)
	}
	return &serviceAccount, err
}

// -------------------------
// 🚀 Update Service Account Validation Status
// -------------------------
func (r *playstoreSettingsRepository) UpdateServiceAccountValidationStatus(ctx context.Context, tx *gorm.DB, serviceAccountID string, isValid bool) error {
	result := tx.WithContext(ctx).Model(&models.GooglePlayServiceAccount{}).Where("id = ?", serviceAccountID).Updates(map[string]interface{}{
		"validated":    isValid,
		"last_checked": time.Now(),
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update validation status for service account %s: %w", serviceAccountID, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no service account found with ID %s: %w", serviceAccountID, ErrRecordNotFound)
	}
	return nil
}

// -------------------------
// 🚀 Delete Package Settings (Soft Delete)
// -------------------------
func (r *playstoreSettingsRepository) DeletePackageSettings(ctx context.Context, tx *gorm.DB, packageName string) error {
	// Soft delete the settings
	result := tx.WithContext(ctx).Where("package_name = ?", packageName).Delete(&models.GooglePlaySettings{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete package settings for %s: %w", packageName, result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no settings found for package %s: %w", packageName, ErrRecordNotFound)
	}

	// Soft delete the associated service account
	if err := tx.WithContext(ctx).Where("package_name = ?", packageName).Delete(&models.GooglePlayServiceAccount{}).Error; err != nil {
		return fmt.Errorf("failed to delete service account for package %s: %w", packageName, err)
	}

	return nil
}
