package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	clientService "subsnotifpro-go/internal/google_playstore/client/service"
	"subsnotifpro-go/internal/google_playstore/settings/repository"
	"subsnotifpro-go/internal/google_playstore/settings/validator"
)

// PlaystoreSettingsService defines an interface for playstore settings service methods
type PlaystoreSettingsService interface {
	// UploadServiceAccount processes a new service account upload
	UploadServiceAccount(filePath string, fileName string) error

	// ValidateServiceAccount checks if the stored service account is valid
	ValidateServiceAccount() error

	// GetServiceAccountStatus retrieves the latest service account validation status
	GetServiceAccountStatus() (map[string]interface{}, error)

	// DeleteServiceAccount deletes the current service account
	DeleteServiceAccount() error

	// SetPackageName updates the package name
	SetPackageName(packageName string) error

	// GetPackageName retrieves the package name
	GetPackageName() (string, error)
}

// playstoreSettingsService Service implements the PlaystoreSettingsService interface
type playstoreSettingsService struct {
	repo          repository.PlaystoreSettingsRepository
	clientService clientService.PlaystoreClientService
	ctx           context.Context
}

// NewRTDNService creates a new instance of RTDNService
func NewPlaystoreSettingsService(ctx context.Context, repo repository.PlaystoreSettingsRepository, clientService clientService.PlaystoreClientService) PlaystoreSettingsService {
	service := &playstoreSettingsService{repo: repo, clientService: clientService, ctx: ctx}
	return service
}

// UploadServiceAccount processes a new service account upload
func (s *playstoreSettingsService) UploadServiceAccount(filePath string, fileName string) error {
	if err := s.repo.DeleteExistingServiceAccount(s.ctx); err != nil {
		log.Println("⚠️ Error deleting old service account:", err)
	}
	return s.repo.SaveServiceAccount(s.ctx, filePath, fileName)
}

// ValidateServiceAccount checks if the stored service account is valid
func (s *playstoreSettingsService) ValidateServiceAccount() error {
	packageName, err := s.GetPackageName()
	if err != nil || packageName == "" {
		return errors.New("package name not set. Please configure it first")
	}

	serviceAccount, err := s.repo.GetLatestServiceAccount(s.ctx)
	if err != nil {
		return errors.New("no service account found")
	}

	if err := validator.ValidateServiceAccountJSONStructure(serviceAccount.FilePath); err != nil {
		return err
	}

	// Use the centralized publisher service client (singleton)
	service, err := s.clientService.GetPublisherService()
	if err != nil {
		return fmt.Errorf("failed to initialize Google Play Publisher service: %w", err)
	}

	// Call Google Play API to check if we can list subscriptions (basic test for valid service account)
	_, err = service.Monetization.Subscriptions.List(packageName).Do()
	if err != nil {
		// Update status as invalid if API call fails
		s.repo.UpdateServiceAccountValidationStatus(s.ctx, false)
		return fmt.Errorf("service account validation failed: %w", err)
	}

	s.repo.UpdateServiceAccountValidationStatus(s.ctx, true)
	return nil
}

// GetServiceAccountStatus retrieves the latest service account validation status
func (s *playstoreSettingsService) GetServiceAccountStatus() (map[string]interface{}, error) {
	serviceAccount, err := s.repo.GetLatestServiceAccount(s.ctx)
	if err != nil {
		return nil, err
	}

	packageName, _ := s.GetPackageName()

	return map[string]interface{}{
		"validated":    serviceAccount.Validated,
		"last_checked": serviceAccount.LastChecked,
		"file_name":    serviceAccount.FileName,
		"package_name": packageName,
	}, nil
}

// DeleteServiceAccount deletes the current service account
func (s *playstoreSettingsService) DeleteServiceAccount() error {
	return s.repo.DeleteExistingServiceAccount(s.ctx)
}

// SetPackageName updates the package name
func (s *playstoreSettingsService) SetPackageName(packageName string) error {
	return s.repo.UpdatePackageName(s.ctx, packageName)
}

// GetPackageName retrieves the package name
func (s *playstoreSettingsService) GetPackageName() (string, error) {
	return s.repo.FetchPackageName(s.ctx)
}
