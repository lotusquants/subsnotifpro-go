package service

import (
	"context"
	"fmt"

	"subsnotifpro-go/internal/google_playstore/settings/repository"
	"subsnotifpro-go/internal/google_playstore/settings/validator"

	"gorm.io/gorm"
)

type PlaystoreSettingsService interface {
	UploadServiceAccount(ctx context.Context, appID, filePath, fileName string) error
	ValidateServiceAccount(ctx context.Context, appID string) error
	GetServiceAccountStatus(ctx context.Context, appID string) (map[string]interface{}, error)
}

type playstoreSettingsService struct {
	db   *gorm.DB
	repo repository.PlaystoreSettingsRepository
}

func NewPlaystoreSettingsService(db *gorm.DB, repo repository.PlaystoreSettingsRepository) PlaystoreSettingsService {
	return &playstoreSettingsService{db: db, repo: repo}
}

// UploadServiceAccount saves a new service account and updates app settings
func (s *playstoreSettingsService) UploadServiceAccount(ctx context.Context, appID, filePath, fileName string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// ✅ Soft delete existing accounts for the app
		if err := s.repo.SoftDeleteAllServiceAccounts(tx, appID); err != nil {
			return err
		}

		// ✅ Save the new service account and upsert into settings
		_, err := s.repo.SaveServiceAccount(tx, appID, filePath, fileName)
		return err
	})
}

// ValidateServiceAccount validates the stored service account for an app
func (s *playstoreSettingsService) ValidateServiceAccount(ctx context.Context, appID string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Get latest service account
		account, err := s.repo.GetLatestServiceAccount(tx, appID)
		if err != nil {
			return err
		}

		// Step 1: Validate JSON structure
		if err := validator.ValidateServiceAccountJSONStructure(account.FilePath); err != nil {
			_ = s.repo.UpdateServiceAccountValidationStatus(tx, appID, false)
			return fmt.Errorf("JSON invalid: %w", err)
		}

		// Step 2: [Optional] External validation can be added here

		// Step 3: Update validation status
		if err := s.repo.UpdateServiceAccountValidationStatus(tx, appID, true); err != nil {
			return err
		}
		return nil
	})
}

// GetServiceAccountStatus retrieves the latest service account validation metadata
func (s *playstoreSettingsService) GetServiceAccountStatus(ctx context.Context, appID string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		account, err := s.repo.GetLatestServiceAccount(tx, appID)
		if err != nil {
			return err
		}

		result = map[string]interface{}{
			"validated":    account.Validated,
			"last_checked": account.LastChecked,
			"file_name":    account.FileName,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}
