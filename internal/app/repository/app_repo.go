package repository

import (
	"context"

	"subsnotifpro-go/internal/app/models"

	"gorm.io/gorm"
)

type AppRepository interface {
	CreateApp(ctx context.Context, app *models.App) error
	GetAppByID(ctx context.Context, id string) (*models.App, error)

	// Transaction support
	WithTx(tx *gorm.DB) AppRepository
}

type appRepo struct {
	db *gorm.DB
}

func NewAppRepository(db *gorm.DB) AppRepository {
	return &appRepo{db: db}
}

// WithTx returns a new instance of AppRepository using the given transaction
func (r *appRepo) WithTx(tx *gorm.DB) AppRepository {
	return &appRepo{db: tx}
}

func (r *appRepo) CreateApp(ctx context.Context, app *models.App) error {
	return r.db.WithContext(ctx).Create(app).Error
}

func (r *appRepo) GetAppByID(ctx context.Context, id string) (*models.App, error) {
	var app models.App
	err := r.db.WithContext(ctx).First(&app, "id = ?", id).Error
	return &app, err
}
