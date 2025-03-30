package repository

import (
	"context"

	"subsnotifpro-go/internal/app/models"

	"gorm.io/gorm"
)

type AppSetupRepository interface {
	WithTx(tx *gorm.DB) AppSetupRepository

	CreateAppSetup(ctx context.Context, setup *models.AppSetup) error
	GetAppSetupByAppID(ctx context.Context, appID string) (*models.AppSetup, error)
	UpdateSetupStatus(ctx context.Context, appID string, status models.SetupStatus) error
}

type appSetupRepo struct {
	db *gorm.DB
}

func NewAppSetupRepository(db *gorm.DB) AppSetupRepository {
	return &appSetupRepo{db: db}
}

// WithTx returns a new instance of AppSetupRepository with the given transaction
func (r *appSetupRepo) WithTx(tx *gorm.DB) AppSetupRepository {
	return &appSetupRepo{db: tx}
}

func (r *appSetupRepo) CreateAppSetup(ctx context.Context, setup *models.AppSetup) error {
	return r.db.WithContext(ctx).Create(setup).Error
}

func (r *appSetupRepo) GetAppSetupByAppID(ctx context.Context, appID string) (*models.AppSetup, error) {
	var setup models.AppSetup
	err := r.db.WithContext(ctx).First(&setup, "app_id = ?", appID).Error
	return &setup, err
}

func (r *appSetupRepo) UpdateSetupStatus(ctx context.Context, appID string, status models.SetupStatus) error {
	return r.db.WithContext(ctx).
		Model(&models.AppSetup{}).
		Where("app_id = ?", appID).
		Update("setup_status", status).Error
}
