package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	settingsModels "subsnotifpro-go/internal/google_playstore/settings/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PlaystoreSettingsRepository defines the interface for Play Store settings repository
type PlaystoreSettingsRepository interface {
	SaveServiceAccount(ctx context.Context, filePath string, fileName string) error
	GetLatestServiceAccount(ctx context.Context) (settingsModels.GooglePlayServiceAccount, error)
	UpdateServiceAccountValidationStatus(ctx context.Context, isValid bool) error
	DeleteExistingServiceAccount(ctx context.Context) error
	UpdatePackageName(ctx context.Context, packageName string) error
	FetchPackageName(ctx context.Context) (string, error)
	GetSettings(ctx context.Context) (settingsModels.GooglePlaySettings, error)
	SaveSettings(ctx context.Context, settings settingsModels.GooglePlaySettings) error
	DeletePackageName(ctx context.Context) error
}

// ✅ Repository instance with injected DB (No Singleton!)
type playstoreSettingsRepository struct {
	db *gorm.DB
}

// ✅ Constructor function (Injects DB instance)
func NewPlaystoreSettingsRepository(db *gorm.DB) PlaystoreSettingsRepository {
	return &playstoreSettingsRepository{db: db}
}

// -------------------------
// 🚀 Save New Service Account (Transaction)
// -------------------------
func (r *playstoreSettingsRepository) SaveServiceAccount(ctx context.Context, filePath string, fileName string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		serviceAccount := settingsModels.GooglePlayServiceAccount{
			FileName: fileName,
			FilePath: filePath,
		}
		if err := tx.Create(&serviceAccount).Error; err != nil {
			return fmt.Errorf("failed to save service account %s: %w", fileName, err)
		}
		return nil
	})
}

// -------------------------
// 🚀 Get Latest Service Account
// -------------------------
func (r *playstoreSettingsRepository) GetLatestServiceAccount(ctx context.Context) (settingsModels.GooglePlayServiceAccount, error) {
	var serviceAccount settingsModels.GooglePlayServiceAccount
	err := r.db.WithContext(ctx).Order("created_at desc").First(&serviceAccount).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return settingsModels.GooglePlayServiceAccount{}, fmt.Errorf("no service account found: %w", err)
	}
	return serviceAccount, err
}

// -------------------------
// 🚀 Update Service Account Validation Status (Transaction)
// -------------------------
func (r *playstoreSettingsRepository) UpdateServiceAccountValidationStatus(ctx context.Context, isValid bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var serviceAccount settingsModels.GooglePlayServiceAccount
		if err := tx.Order("created_at desc").First(&serviceAccount).Error; err != nil {
			return fmt.Errorf("no service account found to update validation status: %w", err)
		}

		serviceAccount.Validated = isValid
		serviceAccount.LastChecked = time.Now()
		if err := tx.Save(&serviceAccount).Error; err != nil {
			return fmt.Errorf("failed to update validation status: %w", err)
		}
		return nil
	})
}

// -------------------------
// 🚀 Delete Existing Service Account (DB First, Then File)
// -------------------------
func (r *playstoreSettingsRepository) DeleteExistingServiceAccount(ctx context.Context) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		serviceAccount, err := r.GetLatestServiceAccount(ctx)
		if err != nil {
			return err // Already wrapped
		}

		// ✅ First delete from DB (prevents orphaned files)
		if err := tx.Delete(&serviceAccount).Error; err != nil {
			return fmt.Errorf("failed to delete service account record: %w", err)
		}

		// ✅ Then delete the file (Check if exists)
		if _, err := os.Stat(serviceAccount.FilePath); err == nil {
			if err := os.Remove(serviceAccount.FilePath); err != nil {
				log.Printf("⚠️ Error deleting service account file: %v", err)
			}
		}

		return nil
	})
}

// -------------------------
// 🚀 Update Package Name
// -------------------------
func (r *playstoreSettingsRepository) UpdatePackageName(ctx context.Context, packageName string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var settings settingsModels.GooglePlaySettings

		// Fetch existing settings record with row locking
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&settings).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No settings found, create a new record
			newSettings := settingsModels.GooglePlaySettings{
				PackageName:            packageName,
				LatestServiceAccountID: nil,
			}

			if err := tx.Create(&newSettings).Error; err != nil {
				return fmt.Errorf("❌ failed to create settings record: %w", err)
			}
			return nil
		}

		// Update the package name
		if err := tx.Model(&settings).Update("package_name", packageName).Error; err != nil {
			return fmt.Errorf("❌ failed to update package name: %w", err)
		}

		return nil
	})
}

// -------------------------
// 🚀 Fetch Package Name
// -------------------------
func (r *playstoreSettingsRepository) FetchPackageName(ctx context.Context) (string, error) {
	var settings settingsModels.GooglePlaySettings
	err := r.db.WithContext(ctx).First(&settings).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("package name not found: %w", err)
	}
	return settings.PackageName, nil
}

// -------------------------
// 🚀 Get Full Settings
// -------------------------
func (r *playstoreSettingsRepository) GetSettings(ctx context.Context) (settingsModels.GooglePlaySettings, error) {
	var settings settingsModels.GooglePlaySettings
	err := r.db.WithContext(ctx).First(&settings).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return settingsModels.GooglePlaySettings{}, fmt.Errorf("no settings found: %w", err)
	}
	return settings, nil
}

// -------------------------
// 🚀 Save Settings (Upsert)
// -------------------------
func (r *playstoreSettingsRepository) SaveSettings(ctx context.Context, settings settingsModels.GooglePlaySettings) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"package_name"}), // ✅ Only updates required columns
		}).Create(&settings).Error
	})
}

// -------------------------
// 🚀 Delete Package Name (Transaction)
// -------------------------
func (r *playstoreSettingsRepository) DeletePackageName(ctx context.Context) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		settings, err := r.GetSettings(ctx)
		if err != nil {
			return err // Already wrapped
		}

		// Clear the package name field
		return tx.Model(&settings).Update("package_name", "").Error
	})
}
