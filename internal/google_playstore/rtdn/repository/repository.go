package repository

import (
	"context"
	"fmt"

	"subsnotifpro-go/internal/google_playstore/rtdn/models"
	"subsnotifpro-go/internal/logger"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// RTDNRepository defines the interface for RTDN repository
type RTDNRepository interface {
	SaveWebhookEvent(ctx context.Context, event *models.GooglePlayWebhookEvent) error
	GetPendingEvents(ctx context.Context, limit int) ([]models.GooglePlayWebhookEvent, error)
	UpdateWebhookStatus(ctx context.Context, eventID string, status string) error
	IncrementRetryCount(ctx context.Context, eventID string) error
	MoveToDeadLetterQueue(ctx context.Context, eventID string) error
}

// ✅ Struct with Injected Database Instance
type rtdnRepository struct {
	db *gorm.DB
}

// ✅ Constructor Function to Inject DB
func NewRTDNRepository(db *gorm.DB) RTDNRepository {
	return &rtdnRepository{db: db}
}

// -------------------------
// 🚀 Save Webhook Event (Transaction)
// -------------------------
func (r *rtdnRepository) SaveWebhookEvent(ctx context.Context, event *models.GooglePlayWebhookEvent) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(event).Error; err != nil {
			return fmt.Errorf("failed to save webhook event: %w", err)
		}
		return nil
	})
}

// -------------------------
// 🚀 Get Pending Events
// -------------------------
func (r *rtdnRepository) GetPendingEvents(ctx context.Context, limit int) ([]models.GooglePlayWebhookEvent, error) {
	var events []models.GooglePlayWebhookEvent
	err := r.db.WithContext(ctx).
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
func (r *rtdnRepository) UpdateWebhookStatus(ctx context.Context, eventID string, status string) error {
	logger.Log.WithFields(logrus.Fields{
		"event_id": eventID,
		"status":   status,
	}).Info("✅ Updating webhook status")

	result := r.db.WithContext(ctx).
		Model(&models.GooglePlayWebhookEvent{}).
		Where("id = ?", eventID).
		Update("status", status)

	if result.Error != nil {
		logger.Log.Errorf("❌ Error updating webhook status: %v", result.Error)
		return fmt.Errorf("failed to update webhook status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		logger.Log.Warnf("⚠️ Event not found: %s", eventID)
		return fmt.Errorf("event not found: %s", eventID)
	}

	logger.Log.WithFields(logrus.Fields{
		"event_id": eventID,
		"status":   status,
	}).Info("✅ Webhook status updated successfully")

	return nil
}

// -------------------------
// 🚀 Increment Retry Count
// -------------------------
func (r *rtdnRepository) IncrementRetryCount(ctx context.Context, eventID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
	})
}

// -------------------------
// 🚀 Move to Dead Letter Queue (DLQ)
// -------------------------
func (r *rtdnRepository) MoveToDeadLetterQueue(ctx context.Context, eventID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
	})
}
