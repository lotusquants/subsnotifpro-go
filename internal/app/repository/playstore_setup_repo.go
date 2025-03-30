package repository

import (
	"context"

	"subsnotifpro-go/internal/app/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PlaystoreSetupRepository interface {
	WithTx(tx *gorm.DB) PlaystoreSetupRepository

	CreateOrUpdate(ctx context.Context, setup *models.PlaystoreSetup) error
	GetByAppID(ctx context.Context, appID string) (*models.PlaystoreSetup, error)
	MarkServiceAccountVerified(ctx context.Context, appID string, verified bool) error
	MarkPubSubCreated(ctx context.Context, appID string, created bool) error
	UpdateBucketID(ctx context.Context, appID string, bucketID string) error
}

type playstoreSetupRepo struct {
	db *gorm.DB
}

func NewPlaystoreSetupRepository(db *gorm.DB) PlaystoreSetupRepository {
	return &playstoreSetupRepo{db: db}
}

// WithTx allows injecting a transaction
func (r *playstoreSetupRepo) WithTx(tx *gorm.DB) PlaystoreSetupRepository {
	return &playstoreSetupRepo{db: tx}
}

func (r *playstoreSetupRepo) CreateOrUpdate(ctx context.Context, setup *models.PlaystoreSetup) error {
	return r.db.WithContext(ctx).Clauses(
		clause.OnConflict{
			UpdateAll: true,
		},
	).Create(setup).Error
}

func (r *playstoreSetupRepo) GetByAppID(ctx context.Context, appID string) (*models.PlaystoreSetup, error) {
	var setup models.PlaystoreSetup
	err := r.db.WithContext(ctx).First(&setup, "app_id = ?", appID).Error
	return &setup, err
}

func (r *playstoreSetupRepo) MarkServiceAccountVerified(ctx context.Context, appID string, verified bool) error {
	return r.db.WithContext(ctx).
		Model(&models.PlaystoreSetup{}).
		Where("app_id = ?", appID).
		Update("service_account_verified", verified).Error
}

func (r *playstoreSetupRepo) MarkPubSubCreated(ctx context.Context, appID string, created bool) error {
	return r.db.WithContext(ctx).
		Model(&models.PlaystoreSetup{}).
		Where("app_id = ?", appID).
		Update("pub_sub_created", created).Error
}

func (r *playstoreSetupRepo) UpdateBucketID(ctx context.Context, appID string, bucketID string) error {
	return r.db.WithContext(ctx).
		Model(&models.PlaystoreSetup{}).
		Where("app_id = ?", appID).
		Update("bucket_id", bucketID).Error
}
