package service

import (
	"context"
	"fmt"

	"subsnotifpro-go/internal/app/models"
	"subsnotifpro-go/internal/app/repository"

	"gorm.io/gorm"
)

// Interface for dependency injection and unit testing
type AppServiceInterface interface {
	CreateAppWithSetup(ctx context.Context, app *models.App) error
	GetAppWithSetup(ctx context.Context, appID string) (*models.App, *models.AppSetup, error)
	UpdateSetupStatus(ctx context.Context, appID string, status models.SetupStatus) error
}

// Concrete implementation
type appService struct {
	db           *gorm.DB
	appRepo      repository.AppRepository
	appSetupRepo repository.AppSetupRepository
}

// Constructor returning interface
func NewAppService(
	db *gorm.DB,
	appRepo repository.AppRepository,
	appSetupRepo repository.AppSetupRepository,
) AppServiceInterface {
	return &appService{
		db:           db,
		appRepo:      appRepo,
		appSetupRepo: appSetupRepo,
	}
}

// Transactional app + app_setup creation
func (s *appService) CreateAppWithSetup(ctx context.Context, app *models.App) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txAppRepo := s.appRepo.WithTx(tx)
		txSetupRepo := s.appSetupRepo.WithTx(tx)

		if err := txAppRepo.CreateApp(ctx, app); err != nil {
			return fmt.Errorf("failed to create app: %w", err)
		}

		setup := &models.AppSetup{
			AppID:       app.ID,
			Platform:    app.Platform,
			SetupStatus: models.SetupPending,
		}

		if err := txSetupRepo.CreateAppSetup(ctx, setup); err != nil {
			return fmt.Errorf("failed to create app setup: %w", err)
		}

		return nil
	})
}

// Read-only: returns both app and setup
func (s *appService) GetAppWithSetup(ctx context.Context, appID string) (*models.App, *models.AppSetup, error) {
	app, err := s.appRepo.GetAppByID(ctx, appID)
	if err != nil {
		return nil, nil, fmt.Errorf("get app failed: %w", err)
	}

	setup, err := s.appSetupRepo.GetAppSetupByAppID(ctx, appID)
	if err != nil {
		return nil, nil, fmt.Errorf("get setup failed: %w", err)
	}

	return app, setup, nil
}

// Just update setup status
func (s *appService) UpdateSetupStatus(ctx context.Context, appID string, status models.SetupStatus) error {
	return s.appSetupRepo.UpdateSetupStatus(ctx, appID, status)
}
