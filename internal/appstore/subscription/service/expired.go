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

func (s *appstoreSubscriptionService) handleExpired(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload

	// Validate required fields
	if payload.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in expired notification")
	}

	if notification.ResponseBodyV2DecodedPayload.Subtype == nil {
		return nil, fmt.Errorf("missing subtype in EXPIRED notification")
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

	// Determine reason based on subtype
	var reason string
	switch *notification.ResponseBodyV2DecodedPayload.Subtype {
	case dto.VOLUNTARY:
		reason = "User voluntarily canceled subscription"
	case dto.BILLING_RETRY:
		reason = "Subscription expired after billing retry period ended"
	case dto.PRICE_INCREASE_EXPIRED:
		reason = "Subscription expired due to price increase rejection"
	case dto.PRODUCT_NOT_FOR_SALE:
		reason = "Subscription expired because product was removed from sale"
	default:
		reason = "Subscription expired"
	}

	// Update subscription fields
	subscription.Status = models.SubscriptionStatusExpired
	subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload
	subscription.AutoRenewStatus = models.AutoRenewOff

	// Clear grace period if set
	subscription.GracePeriodExpiresDate = nil

	// Save subscription updates
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription updates: %w", err)
	}

	// Save subscription and create event atomically
	if err := subscription.AddEvent(tx, models.EventTypeExpiration, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add expiration event: %w", err)
	}

	logger.Log.Infof("Processed expiration for subscription %s (user: %s, original transaction ID: %s, "+
		"previous status: %s, reason: %s)",
		subscription.ID, subscription.UserID, payload.OriginalTransactionId, previousStatus, reason)

	// // Post-processing based on expiration reason
	// switch *notification.ResponseBodyV2DecodedPayload.Subtype {
	// case dto.VOLUNTARY:
	// 	if err := s.notificationService.SendVoluntaryExpirationNotification(
	// 		ctx,
	// 		subscription.UserID,
	// 	); err != nil {
	// 		logger.Log.Warnf("Failed to send voluntary expiration notification: %v", err)
	// 	}
	// case dto.PRICE_INCREASE:
	// 	if err := s.notificationService.SendPriceIncreaseRejectionNotification(
	// 		ctx,
	// 		subscription.UserID,
	// 	); err != nil {
	// 		logger.Log.Warnf("Failed to send price increase rejection notification: %v", err)
	// 	}
	// default:
	// 	if err := s.notificationService.SendExpirationNotification(
	// 		ctx,
	// 		subscription.UserID,
	// 		reason,
	// 	); err != nil {
	// 		logger.Log.Warnf("Failed to send expiration notification: %v", err)
	// 	}
	// }

	return subscription, nil
}
