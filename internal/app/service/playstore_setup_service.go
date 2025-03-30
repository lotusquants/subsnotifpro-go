package service

import (
	"context"
	"fmt"

	"subsnotifpro-go/internal/app/models"
	"subsnotifpro-go/internal/app/repository"
	playstoreApiService "subsnotifpro-go/internal/playstore/api/service"
	"subsnotifpro-go/internal/playstore/settings/validator"

	"gorm.io/gorm"
)

type PlaystoreSetupServiceInterface interface {
	SaveSetup(ctx context.Context, setup *models.PlaystoreSetup) error
	GetByAppID(ctx context.Context, appID string) (*models.PlaystoreSetup, error)
	SetServiceAccountVerified(ctx context.Context, appID string, verified bool) error
	SetPubSubCreated(ctx context.Context, appID string, created bool) error
	UpdateBucketID(ctx context.Context, appID string, bucketID string) error
	ValidateServiceAccount(ctx context.Context, appID string) error
}

type playstoreSetupService struct {
	db                  *gorm.DB
	repo                repository.PlaystoreSetupRepository
	appSetupRepo        repository.AppSetupRepository
	playstoreApiService playstoreApiService.PlaystoreApiService
}

// ✅ Constructor returning interface
func NewPlaystoreSetupService(
	db *gorm.DB,
	repo repository.PlaystoreSetupRepository,
	appSetupRepo repository.AppSetupRepository,
	playstoreApiService playstoreApiService.PlaystoreApiService,
) PlaystoreSetupServiceInterface {
	return &playstoreSetupService{
		db:                  db,
		repo:                repo,
		appSetupRepo:        appSetupRepo,
		playstoreApiService: playstoreApiService,
	}
}

// Save or update the playstore setup (upsert)
func (s *playstoreSetupService) SaveSetup(ctx context.Context, setup *models.PlaystoreSetup) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		if err := txRepo.CreateOrUpdate(ctx, setup); err != nil {
			return fmt.Errorf("failed to save playstore setup: %w", err)
		}
		return nil
	})
}

// Get by App ID
func (s *playstoreSetupService) GetByAppID(ctx context.Context, appID string) (*models.PlaystoreSetup, error) {
	return s.repo.GetByAppID(ctx, appID)
}

// Mark service account verified
func (s *playstoreSetupService) SetServiceAccountVerified(ctx context.Context, appID string, verified bool) error {
	return s.repo.MarkServiceAccountVerified(ctx, appID, verified)
}

// Mark PubSub creation done
func (s *playstoreSetupService) SetPubSubCreated(ctx context.Context, appID string, created bool) error {
	return s.repo.MarkPubSubCreated(ctx, appID, created)
}

// Optional: Set GCP bucket ID
func (s *playstoreSetupService) UpdateBucketID(ctx context.Context, appID string, bucketID string) error {
	return s.repo.UpdateBucketID(ctx, appID, bucketID)
}

// Validate the uploaded service account + update flags
func (s *playstoreSetupService) ValidateServiceAccount(ctx context.Context, appID string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)
		txSetupRepo := s.appSetupRepo.WithTx(tx)

		setup, err := txRepo.GetByAppID(ctx, appID)
		if err != nil {
			return fmt.Errorf("setup not found for app %s: %w", appID, err)
		}

		if setup.PackageName == "" || setup.ServiceAccountPath == "" {
			return fmt.Errorf("missing package name or service account path")
		}

		if err := validator.ValidateServiceAccountJSONStructure(setup.ServiceAccountPath); err != nil {
			_ = txRepo.MarkServiceAccountVerified(ctx, appID, false)
			return fmt.Errorf("invalid JSON structure: %w", err)
		}

		canAccess, err := s.playstoreApiService.VerifyServiceAccountAccess(ctx, setup.ServiceAccountPath, setup.PackageName)
		if err != nil {
			_ = txRepo.MarkServiceAccountVerified(ctx, appID, false)
			return fmt.Errorf("API validation failed: %w", err)
		}

		if !canAccess {
			_ = txRepo.MarkServiceAccountVerified(ctx, appID, false)
			return fmt.Errorf("service account lacks permission for package: %s", setup.PackageName)
		}

		if err := txRepo.MarkServiceAccountVerified(ctx, appID, true); err != nil {
			return fmt.Errorf("failed to mark service account verified: %w", err)
		}
		if err := txSetupRepo.UpdateSetupStatus(ctx, appID, models.SetupIncomplete); err != nil {
			return fmt.Errorf("failed to update setup status: %w", err)
		}

		return nil
	})
}
