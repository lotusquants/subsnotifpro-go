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

func (s *appstoreSubscriptionService) handleDidChangeRenewalStatus(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload

	// Validate required fields
	if payload.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in renewal status change notification")
	}

	if notification.ResponseBodyV2DecodedPayload.Subtype == nil {
		return nil, fmt.Errorf("missing subtype in DID_CHANGE_RENEWAL_STATUS notification")
	}

	// Determine event type based on subtype
	var eventType models.SubscriptionEventType
	var newAutoRenewStatus models.AutoRenewStatus
	var reason string

	switch *notification.ResponseBodyV2DecodedPayload.Subtype {
	case dto.AUTO_RENEW_ENABLED:
		eventType = models.EventTypeAutoRenewEnable
		newAutoRenewStatus = models.AutoRenewOn
		reason = "User enabled auto-renewal"
	case dto.AUTO_RENEW_DISABLED:
		eventType = models.EventTypeAutoRenewDisable
		newAutoRenewStatus = models.AutoRenewOff
		reason = "User disabled auto-renewal"
	default:
		return nil, fmt.Errorf("unhandled DID_CHANGE_RENEWAL_STATUS subtype: %s",
			*notification.ResponseBodyV2DecodedPayload.Subtype)
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
	previousAutoRenewStatus := subscription.AutoRenewStatus

	// Update subscription fields
	subscription.AutoRenewStatus = newAutoRenewStatus
	subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload

	// Only update these fields if they're not zero values
	if !payload.PurchaseDate.IsZero() {
		subscription.PurchaseDate = payload.PurchaseDate
	}
	if !payload.ExpiresDate.IsZero() {
		subscription.ExpiresDate = payload.ExpiresDate
	}

	// Save subscription updates first
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription updates: %w", err)
	}

	// Save subscription and create event atomically
	if err := subscription.AddEvent(tx, eventType, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add renewal status change event: %w", err)
	}

	logger.Log.Infof("Processed %s for subscription %s (user: %s, original transaction ID: %s, previous auto-renew: %s)",
		eventType, subscription.ID, subscription.UserID, payload.OriginalTransactionId, previousAutoRenewStatus)

	// // Post-processing based on auto-renewal status change
	// switch eventType {
	// case models.EventTypeAutoRenewDisable:
	// 	// Notify user about expiration timeline
	// 	remainingDuration := time.Until(subscription.ExpiresDate)
	// 	logger.Log.Infof("Auto-renew disabled, subscription %s will expire in %v", subscription.ID, remainingDuration)

	// 	if err := s.notificationService.SendAutoRenewDisabledNotification(
	// 		ctx,
	// 		subscription.UserID,
	// 		subscription.ExpiresDate,
	// 	); err != nil {
	// 		logger.Log.Warnf("Failed to send auto-renew disabled notification: %v", err)
	// 	}

	// case models.EventTypeAutoRenewEnable:
	// 	if err := s.notificationService.SendAutoRenewEnabledNotification(
	// 		ctx,
	// 		subscription.UserID,
	// 	); err != nil {
	// 		logger.Log.Warnf("Failed to send auto-renew enabled notification: %v", err)
	// 	}
	// }

	return subscription, nil
}
