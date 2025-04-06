package repository

import (
	"context"
	"fmt"
	"time"

	"subsnotifpro-go/internal/pkg/contextutil"
	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/playstore/rtdn/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RTDNRepository defines the interface for RTDN repository
type RTDNRepository interface {
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
	Create(ctx context.Context, event *models.GooglePlayWebhookEvent) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.WebhookEventStatus, errorMsg string) error
	IncrementRetryCount(ctx context.Context, eventID uuid.UUID) error
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
}

type rtdnRepository struct {
	db *gorm.DB
}

func NewRTDNRepository(db *gorm.DB) RTDNRepository {
	return &rtdnRepository{db: db}
}

func (r *rtdnRepository) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
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

func (r *rtdnRepository) Create(ctx context.Context, event *models.GooglePlayWebhookEvent) error {
	start := time.Now()

	tx, ok := contextutil.TxFromContext(ctx)
	if !ok {
		tx = r.db.WithContext(ctx)
	}

	if err := tx.Create(event).Error; err != nil {
		logger.Log.Errorf("failed to create event: %v", err)
		return fmt.Errorf("repository create failed: %w", err)
	}

	logger.Log.Infof("event created in %v", time.Since(start))
	return nil
}

func (r *rtdnRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.WebhookEventStatus, errorMsg string) error {
	start := time.Now()

	tx, ok := contextutil.TxFromContext(ctx)
	if !ok {
		tx = r.db.WithContext(ctx)
	}

	result := tx.Model(&models.GooglePlayWebhookEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status": status,
			"error":  errorMsg,
		})

	if result.Error != nil {
		logger.Log.Errorf("failed to update status: %v", result.Error)
		return fmt.Errorf("repository update status failed: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		logger.Log.Warnf("no rows affected when updating status for event %s", id)
		return fmt.Errorf("event not found")
	}

	logger.Log.Infof("status updated in %v", time.Since(start))
	return nil
}

func (r *rtdnRepository) IncrementRetryCount(ctx context.Context, eventID uuid.UUID) error {
	tx, ok := contextutil.TxFromContext(ctx)
	if !ok {
		tx = r.db.WithContext(ctx)
	}

	result := tx.Model(&models.GooglePlayWebhookEvent{}).
		Where("id = ?", eventID).
		Update("retry_count", gorm.Expr("retry_count + ?", 1))

	if result.Error != nil {
		return fmt.Errorf("failed to increment retry count: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("event not found when trying to increment retry count: %s", eventID)
	}
	return nil
}

// Implementation
func (r *rtdnRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	tx, ok := contextutil.TxFromContext(ctx)
	if !ok {
		tx = r.db.WithContext(ctx)
	}

	var count int64
	if err := tx.Model(&models.GooglePlayWebhookEvent{}).
		Where("id = ?", id).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return count > 0, nil
}
