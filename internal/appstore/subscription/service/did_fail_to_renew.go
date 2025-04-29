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

func (s *appstoreSubscriptionService) handleDidFailToRenew(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload

	// Validate required fields
	if payload.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in failed renewal notification")
	}

	if notification.ResponseBodyV2DecodedPayload.Subtype == nil {
		return nil, fmt.Errorf("missing subtype in DID_FAIL_TO_RENEW notification")
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

		// Update user ID if not set (for legacy subscriptions)
		if subscription.UserID == uuid.Nil {
			subscription.UserID = *userID
		}
	}

	// Store previous state for comparison
	previousStatus := subscription.Status

	// Determine event type and status based on subtype
	var eventType models.SubscriptionEventType
	var newStatus models.SubscriptionStatus
	var reason string

	switch *notification.ResponseBodyV2DecodedPayload.Subtype {
	case dto.GRACE_PERIOD:
		eventType = models.EventTypeGracePeriodEnter
		newStatus = models.SubscriptionStatusGracePeriod
		reason = "Subscription entered grace period after failed renewal"
	default:
		eventType = models.EventTypeBillingRetry
		newStatus = models.SubscriptionStatusBillingRetry
		reason = "Subscription failed to renew, entered billing retry"
	}

	// Update subscription fields
	subscription.Status = newStatus
	subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload

	// Update grace period expiration if applicable
	if *notification.ResponseBodyV2DecodedPayload.Subtype == dto.GRACE_PERIOD {
		if !payload.ExpiresDate.IsZero() {
			subscription.GracePeriodExpiresDate = &payload.ExpiresDate
		}
	}

	// Save subscription updates
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription updates: %w", err)
	}

	// Save subscription and create event atomically
	if err := subscription.AddEvent(tx, eventType, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add failed renewal event: %w", err)
	}

	logger.Log.Infof("Processed %s for subscription %s (user: %s, original transaction ID: %s, previous status: %s, new status: %s)",
		eventType, subscription.ID, subscription.UserID, payload.OriginalTransactionId, previousStatus, newStatus)

	// // Post-processing based on failure type
	// if *notification.ResponseBodyV2DecodedPayload.Subtype == dto.GRACE_PERIOD {
	// 	if err := s.notificationService.SendGracePeriodNotification(
	// 		ctx,
	// 		subscription.UserID,
	// 		subscription.GracePeriodExpiresDate,
	// 	); err != nil {
	// 		logger.Log.Warnf("Failed to send grace period notification: %v", err)
	// 	}
	// } else {
	// 	if err := s.notificationService.SendBillingRetryNotification(
	// 		ctx,
	// 		subscription.UserID,
	// 		subscription.ExpiresDate,
	// 	); err != nil {
	// 		logger.Log.Warnf("Failed to send billing retry notification: %v", err)
	// 	}
	// }

	return subscription, nil
}
