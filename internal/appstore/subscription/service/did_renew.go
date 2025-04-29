package service

import (
	"context"
	"errors"
	"fmt"
	"subsnotifpro-go/internal/appstore/subscription/models"
	"subsnotifpro-go/internal/pkg/logger"
	"time"

	"subsnotifpro-go/internal/appstore/webhooks/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *appstoreSubscriptionService) handleDidRenew(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {

	logger.Log.Infof("Processing notification type %s app store subscription for transaction ID %s",
		notification.ResponseBodyV2DecodedPayload.NotificationType,
		notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload.TransactionId)

	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload
	renewalInfo := notification.ResponseBodyV2DecodedPayload.Data.SignedRenewalInfo.JWSRenewalInfoDecodedPayload

	// Validate required fields
	if payload.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in renewal notification")
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

	// Validate subscription status transition
	if !isValidRenewalTransition(subscription.Status) {
		logger.Log.Warnf("Unexpected status transition from %s to ACTIVE for subscription %s",
			subscription.Status, subscription.ID)
	}

	// Store previous state for comparison
	previousStatus := subscription.Status

	// Update subscription fields
	subscription.CurrentTransactionID = payload.TransactionId
	subscription.PurchaseDate = payload.PurchaseDate
	subscription.ExpiresDate = payload.ExpiresDate
	subscription.Status = models.SubscriptionStatusActive
	subscription.AutoRenewStatus = models.MapAutoRenewStatus(renewalInfo.AutoRenewStatus)
	subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload

	// Update pricing information if changed
	if renewalInfo.RenewalPrice != 0 {
		subscription.Price = renewalInfo.RenewalPrice
		subscription.Currency = renewalInfo.Currency
	}

	// Handle grace period expiration if applicable
	if !renewalInfo.GracePeriodExpiresDate.IsZero() {
		if renewalInfo.GracePeriodExpiresDate.After(time.Now()) {
			subscription.GracePeriodExpiresDate = &renewalInfo.GracePeriodExpiresDate
		} else {
			subscription.GracePeriodExpiresDate = nil
		}
	}

	// Update offer details if present
	if renewalInfo.OfferIdentifier != "" {
		subscription.OfferIdentifier = &renewalInfo.OfferIdentifier
		subscription.OfferType = &renewalInfo.OfferType
		if renewalInfo.OfferPeriod != "" {
			subscription.OfferDuration = &renewalInfo.OfferPeriod
		}
	}

	// Determine event type based on subtype
	eventType := models.EventTypeAutomaticRenew
	reason := "Subscription automatically renewed"
	if notification.ResponseBodyV2DecodedPayload.Subtype != nil &&
		*notification.ResponseBodyV2DecodedPayload.Subtype == dto.BILLING_RECOVERY {
		eventType = models.EventTypeBillingRecovery
		reason = "Subscription recovered after billing issue"
	}

	// Save subscription updates first
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription updates: %w", err)
	}

	// Save subscription and create event atomically
	if err := subscription.AddEvent(tx, eventType, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add renewal event: %w", err)
	}

	logger.Log.Infof("Processed %s for subscription %s (user: %s, original transaction ID: %s, previous status: %s)",
		eventType, subscription.ID, subscription.UserID, payload.OriginalTransactionId, previousStatus)

	// // Post-processing for specific renewal types
	// if eventType == models.EventTypeBillingRecovery {
	// 	if err := s.notificationService.SendBillingRecoveryNotification(ctx, subscription.UserID); err != nil {
	// 		logger.Log.Warnf("Failed to send billing recovery notification: %v", err)
	// 	}
	// }

	return subscription, nil
}

func isValidRenewalTransition(currentStatus models.SubscriptionStatus) bool {
	switch currentStatus {
	case models.SubscriptionStatusActive,
		models.SubscriptionStatusBillingRetry,
		models.SubscriptionStatusGracePeriod:
		return true
	default:
		return false
	}
}
