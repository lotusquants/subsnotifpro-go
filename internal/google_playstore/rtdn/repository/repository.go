// internal/google_playstore/rtdn/repository/repository.go
package repository

import (
	"context"
	"errors"
	"log"
	"subsnotifpro-go/database"
	"subsnotifpro-go/internal/google_playstore/models"

	"gorm.io/gorm"
)

// SaveWebhookEvent saves webhook event to the database
func SaveWebhookEvent(event *models.GooglePlayWebhookEvent) error {
	tx := database.DB.Begin() // ✅ Start transaction

	if err := tx.Create(event).Error; err != nil {
		tx.Rollback() // ❌ Rollback on failure
		return err
	}

	return tx.Commit().Error // ✅ Commit transaction
}

// GetPendingEvents retrieves unprocessed webhook events with pagination
func GetPendingEvents(ctx context.Context, limit int) ([]models.GooglePlayWebhookEvent, error) {
	var events []models.GooglePlayWebhookEvent
	tx := database.DB.WithContext(ctx).Begin() // ✅ Pass ctx to ensure cancellation

	err := tx.
		Where("status = ?", "pending").
		Order("retry_count DESC, created_at ASC").
		Limit(limit).
		Find(&events).Error

	if err != nil {
		tx.Rollback() // ❌ Rollback on failure
		return nil, err
	}
	return events, tx.Commit().Error // ✅ Commit if successful
}

// UpdateWebhookStatus updates the processing status of a webhook event
func UpdateWebhookStatus(eventID string, status string) error {
	result := database.DB.Model(&models.GooglePlayWebhookEvent{}).Where("id = ?", eventID).Update("status", status)
	if result.RowsAffected == 0 {
		return errors.New("event not found")
	}
	return nil
}

// IncrementRetryCount increases retry count for failed events
func IncrementRetryCount(eventID string) error {
	return database.DB.Model(&models.GooglePlayWebhookEvent{}).Where("id = ?", eventID).Update("retry_count", gorm.Expr("retry_count + ?", 1)).Error
}

// MoveToDeadLetterQueue moves failed events to DLQ after max retries
func MoveToDeadLetterQueue(event models.GooglePlayWebhookEvent) error {
	event.Status = "dead_letter"
	return database.DB.Save(&event).Error
}

// 🔹 Subscription Event Handlers 🔹

// SaveSubscription saves a new subscription purchase event
func SaveSubscription(event models.GooglePlayWebhookEvent) error {
	log.Println("💾 [Placeholder] Saving new subscription:", event.SubscriptionNotification.SubscriptionID)
	return nil
}

// UpdateSubscriptionRenewal handles subscription renewal
func UpdateSubscriptionRenewal(event models.GooglePlayWebhookEvent) error {
	log.Println("🔄 [Placeholder] Updating subscription renewal:", event.SubscriptionNotification.SubscriptionID)
	return nil
}

// CancelSubscription marks a subscription as canceled
func CancelSubscription(event models.GooglePlayWebhookEvent) error {
	log.Println("🚫 [Placeholder] Canceling subscription:", event.SubscriptionNotification.SubscriptionID)
	return nil
}

// RecoverSubscription recovers a subscription from hold
func RecoverSubscription(event models.GooglePlayWebhookEvent) error {
	log.Println("🔄 [Placeholder] Recovering subscription:", event.SubscriptionNotification.SubscriptionID)
	return nil
}

// HandleSubscriptionOnHold processes subscription hold events
func HandleSubscriptionOnHold(event models.GooglePlayWebhookEvent) error {
	log.Println("⏳ [Placeholder] Handling subscription on hold:", event.SubscriptionNotification.SubscriptionID)
	return nil
}

// HandleSubscriptionInGracePeriod processes subscription grace period events
func HandleSubscriptionInGracePeriod(event models.GooglePlayWebhookEvent) error {
	log.Println("⚠️ [Placeholder] Handling subscription in grace period:", event.SubscriptionNotification.SubscriptionID)
	return nil
}

// HandleSubscriptionRestart processes subscription restart events
func HandleSubscriptionRestart(event models.GooglePlayWebhookEvent) error {
	log.Println("🔄 [Placeholder] Handling subscription restart:", event.SubscriptionNotification.SubscriptionID)
	return nil
}

// HandlePriceChangeConfirmation processes subscription price change confirmation
func HandlePriceChangeConfirmation(event models.GooglePlayWebhookEvent) error {
	log.Println("💰 [Placeholder] Handling price change confirmation:", event.SubscriptionNotification.SubscriptionID)
	return nil
}

// HandleSubscriptionDeferred processes subscription deferment events
func HandleSubscriptionDeferred(event models.GooglePlayWebhookEvent) error {
	log.Println("📅 [Placeholder] Handling subscription deferment:", event.SubscriptionNotification.SubscriptionID)
	return nil
}

// HandleSubscriptionPaused processes subscription pause events
func HandleSubscriptionPaused(event models.GooglePlayWebhookEvent) error {
	log.Println("⏸️ [Placeholder] Handling subscription pause:", event.SubscriptionNotification.SubscriptionID)
	return nil
}

// HandlePauseScheduleChanged processes subscription pause schedule changes
func HandlePauseScheduleChanged(event models.GooglePlayWebhookEvent) error {
	log.Println("🔄 [Placeholder] Handling pause schedule change:", event.SubscriptionNotification.SubscriptionID)
	return nil
}

// HandleSubscriptionRevoked processes subscription revocation events
func HandleSubscriptionRevoked(event models.GooglePlayWebhookEvent) error {
	log.Println("🚫 [Placeholder] Handling subscription revocation:", event.SubscriptionNotification.SubscriptionID)
	return nil
}

// HandleSubscriptionExpired processes subscription expiration events
func HandleSubscriptionExpired(event models.GooglePlayWebhookEvent) error {
	log.Println("⏳ [Placeholder] Handling subscription expiration:", event.SubscriptionNotification.SubscriptionID)
	return nil
}

// HandlePendingPurchaseCanceled processes pending purchase cancellation events
func HandlePendingPurchaseCanceled(event models.GooglePlayWebhookEvent) error {
	log.Println("❌ [Placeholder] Handling pending purchase cancellation:", event.SubscriptionNotification.SubscriptionID)
	return nil
}
