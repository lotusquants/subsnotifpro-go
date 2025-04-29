package service

import (
	"context"
	"errors"
	"fmt"
	"subsnotifpro-go/internal/appstore/subscription/models"
	"subsnotifpro-go/internal/appstore/webhooks/dto"
	"subsnotifpro-go/internal/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *appstoreSubscriptionService) handleGracePeriodExpired(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload

	// Validate required fields
	if payload.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in grace period expired notification")
	}

	// Find existing subscription
	subscription, err := s.repo.FindByOriginalTransactionID(ctx, tx, payload.OriginalTransactionId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("subscription not found for original transaction ID: %s", payload.OriginalTransactionId)
		}
		return nil, fmt.Errorf("failed to find subscription: %w", err)
	}

	// Verify user ownership if userID is provided
	if userID != nil {
		if subscription.UserID != uuid.Nil && subscription.UserID != *userID {
			return nil, fmt.Errorf("user ID mismatch: subscription belongs to %s but request is for %s",
				subscription.UserID, *userID)
		}
	}

	// Validate current status
	if subscription.Status != models.SubscriptionStatusGracePeriod {
		logger.Log.Warnf("Grace period expired for subscription %s with unexpected status: %s",
			subscription.ID, subscription.Status)
	}

	// Update subscription fields
	previousStatus := subscription.Status
	subscription.Status = models.SubscriptionStatusBillingRetry
	subscription.GracePeriodExpiresDate = nil
	subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload

	// Save subscription updates
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription updates: %w", err)
	}

	// Save subscription and create event atomically
	reason := "Grace period expired, entered billing retry"
	if err := subscription.AddEvent(tx, models.EventTypeGracePeriodExit, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add grace period expired event: %w", err)
	}

	logger.Log.Infof("Processed grace period expiration for subscription %s (user: %s, original transaction ID: %s, previous status: %s)",
		subscription.ID, subscription.UserID, payload.OriginalTransactionId, previousStatus)

	// // Post-processing
	// if err := s.notificationService.SendGracePeriodExpiredNotification(
	//     ctx,
	//     subscription.UserID,
	//     subscription.ExpiresDate,
	// ); err != nil {
	//     logger.Log.Warnf("Failed to send grace period expired notification: %v", err)
	// }

	return subscription, nil
}
