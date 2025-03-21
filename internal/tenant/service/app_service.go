package service

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/tenant/models"
	"subsnotifpro-go/internal/tenant/repository"

	"gorm.io/gorm"
)

type IAppService interface {
	CreateApp(ctx context.Context, app *models.App) (*models.App, error)
	ListApps(ctx context.Context, tenantID string) ([]models.App, error)
	GetAppByPackageName(ctx context.Context, packageName string) (*models.App, error)
}

type appService struct {
	db   *gorm.DB
	repo repository.IAppRepository
}

func NewAppService(db *gorm.DB, repo repository.IAppRepository) IAppService {
	return &appService{db: db, repo: repo}
}

func (s *appService) CreateApp(ctx context.Context, app *models.App) (*models.App, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return s.repo.Create(tx, app)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create app: %w", err)
	}
	return app, nil
}

func (s *appService) ListApps(ctx context.Context, tenantID string) ([]models.App, error) {
	var apps []models.App
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		apps, err = s.repo.ListByTenantID(tx, tenantID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}
	return apps, nil
}

func (s *appService) GetAppByPackageName(ctx context.Context, packageName string) (*models.App, error) {
	var app *models.App
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var err error
		app, err = s.repo.GetByPackageName(tx, packageName)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get app by package name: %w", err)
	}
	return app, nil
}
