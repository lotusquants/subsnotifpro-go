package repository

import (
	"context"
	"errors"
	"fmt"
	"subsnotifpro-go/internal/appstore/webhooks/models"
	"subsnotifpro-go/internal/pkg/contextutil"
	"subsnotifpro-go/internal/pkg/logger"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AppstoreNotificationsRepository interface {
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
	Save(ctx context.Context, notification *models.AppStoreNotification) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.WebhookEventStatus, message string) error

	GetByID(ctx context.Context, id uuid.UUID) (*models.AppStoreNotification, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	IncrementRetryCount(ctx context.Context, id uuid.UUID) error
}

type appstoreNotificationsRepository struct {
	db *gorm.DB
}

func NewAppstoreNotificationsRepository(db *gorm.DB) AppstoreNotificationsRepository {
	return &appstoreNotificationsRepository{db: db}
}

func (r *appstoreNotificationsRepository) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	tx := r.db.Begin().WithContext(ctx)
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	txCtx := contextutil.WithTx(ctx, tx)

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Log.Errorf("panic recovered in transaction: %v", r)
			panic(r)
		}
	}()

	if err := fn(txCtx); err != nil {
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			logger.Log.Errorf("rollback failed: %v", rollbackErr)
		}
		return fmt.Errorf("transaction failed: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}

	return nil
}

// Save stores a new AppStore notification in the database
func (r *appstoreNotificationsRepository) Save(ctx context.Context, notification *models.AppStoreNotification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}

	// Use GORM's Create method to insert the record
	result := r.db.WithContext(ctx).Create(notification)
	if result.Error != nil {
		return fmt.Errorf("failed to save notification: %w", result.Error)
	}

	return nil
}

// UpdateStatus updates the status of a notification
func (r *appstoreNotificationsRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.WebhookEventStatus, message string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if message != "" {
		updates["status_message"] = message
	}

	result := r.db.WithContext(ctx).
		Model(&models.AppStoreNotification{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to update status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found with id: %s", id)
	}

	return nil
}

// GetByID retrieves a notification by its ID
func (r *appstoreNotificationsRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AppStoreNotification, error) {
	var notification models.AppStoreNotification
	result := r.db.WithContext(ctx).
		// Main notification relationships
		Preload("ResponseBodyV2DecodedPayload").
		Preload("JWSDecodedHeader").

		// Nested relationships under ResponseBodyV2DecodedPayload
		Preload("ResponseBodyV2DecodedPayload.Data").
		Preload("ResponseBodyV2DecodedPayload.Summary").

		// Relationships under Data
		Preload("ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo").
		Preload("ResponseBodyV2DecodedPayload.Data.SignedRenewalInfo").

		// Nested relationships under SignedTransactionInfo
		Preload("ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.DecodedPayload").

		// Nested relationships under SignedRenewalInfo
		Preload("ResponseBodyV2DecodedPayload.Data.SignedRenewalInfo.DecodedPayload").
		Where("id = ?", id).
		First(&notification)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get notification by ID: %w", result.Error)
	}
	return &notification, nil
}

// Exists checks if a notification with the given ID exists
func (r *appstoreNotificationsRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&models.AppStoreNotification{}).
		Where("id = ?", id).
		Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("failed to check notification existence: %w", result.Error)
	}
	return count > 0, nil
}

// IncrementRetryCount increments the retry count for a notification
func (r *appstoreNotificationsRepository) IncrementRetryCount(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.AppStoreNotification{}).
		Where("id = ?", id).
		Update("retry_count", gorm.Expr("retry_count + 1"))

	if result.Error != nil {
		return fmt.Errorf("failed to increment retry count: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("notification not found with id: %s", id)
	}
	return nil
}
