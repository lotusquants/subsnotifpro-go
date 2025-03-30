package repository

import (
	"context"
	"fmt"

	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/playstore/rtdn/models"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RTDNRepository defines the interface for RTDN repository
type RTDNRepository interface {
	SaveWebhookEvent(ctx context.Context, tx *gorm.DB, event *models.GooglePlayWebhookEvent) error
	GetPendingEvents(ctx context.Context, tx *gorm.DB, limit int) ([]models.GooglePlayWebhookEvent, error)
	UpdateWebhookStatus(ctx context.Context, tx *gorm.DB, eventID string, status string) error
	IncrementRetryCount(ctx context.Context, tx *gorm.DB, eventID string) error
	MoveToDeadLetterQueue(ctx context.Context, tx *gorm.DB, eventID string) error

	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}

// ✅ Struct with Injected Database Instance
type rtdnRepository struct {
	db *gorm.DB
}

// ✅ Constructor Function to Inject DB
func NewRTDNRepository(db *gorm.DB) RTDNRepository {
	return &rtdnRepository{db: db}
}

func (r *rtdnRepository) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(fn)
}

// -------------------------
// 🚀 Save Webhook Event (Transaction)
// -------------------------
func (r *rtdnRepository) SaveWebhookEvent(ctx context.Context, tx *gorm.DB, event *models.GooglePlayWebhookEvent) error {
	if tx == nil {
		tx = r.db.WithContext(ctx)
	} else {
		tx = tx.WithContext(ctx)
	}
	return tx.Create(event).Error
}

// -------------------------
// 🚀 Get Pending Events
// -------------------------
func (r *rtdnRepository) GetPendingEvents(ctx context.Context, tx *gorm.DB, limit int) ([]models.GooglePlayWebhookEvent, error) {
	if tx == nil {
		tx = r.db.WithContext(ctx)
	} else {
		tx = tx.WithContext(ctx)
	}

	var events []models.GooglePlayWebhookEvent
	err := tx.
		Where("status = ?", "pending").
		Order("retry_count DESC, created_at ASC").
		Limit(limit).
		Find(&events).Error

	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending webhook events: %w", err)
	}
	return events, nil
}

// -------------------------
// 🚀 Update Webhook Status
// -------------------------
func (r *rtdnRepository) UpdateWebhookStatus(ctx context.Context, tx *gorm.DB, eventID string, status string) error {
	if tx == nil {
		tx = r.db.WithContext(ctx)
	} else {
		tx = tx.WithContext(ctx)
	}

	logger.Log.WithFields(logrus.Fields{
		"event_id": eventID,
		"status":   status,
	}).Info("✅ Updating webhook status")

	result := tx.Model(&models.GooglePlayWebhookEvent{}).
		Where("id = ?", eventID).
		Update("status", status)

	if result.Error != nil {
		return fmt.Errorf("failed to update webhook status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("event not found: %s", eventID)
	}

	return nil
}

// -------------------------
// 🚀 Increment Retry Count
// -------------------------
func (r *rtdnRepository) IncrementRetryCount(ctx context.Context, tx *gorm.DB, eventID string) error {
	if tx == nil {
		tx = r.db.WithContext(ctx)
	} else {
		tx = tx.WithContext(ctx)
	}

	result := tx.Model(&models.GooglePlayWebhookEvent{}).
		Where("id = ?", eventID).
		Update("retry_count", gorm.Expr("retry_count + ?", 1))

	if result.Error != nil {
		return fmt.Errorf("failed to increment retry count: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("event not found: %s", eventID)
	}
	return nil
}

// -------------------------
// 🚀 Move to Dead Letter Queue (DLQ)
// -------------------------

func (r *rtdnRepository) MoveToDeadLetterQueue(ctx context.Context, tx *gorm.DB, eventID string) error {
	if tx == nil {
		tx = r.db.WithContext(ctx)
	} else {
		tx = tx.WithContext(ctx)
	}

	result := tx.Model(&models.GooglePlayWebhookEvent{}).
		Where("id = ?", eventID).
		Update("status", "dead_letter")

	if result.Error != nil {
		return fmt.Errorf("failed to move event to DLQ: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("event not found: %s", eventID)
	}
	return nil
}
