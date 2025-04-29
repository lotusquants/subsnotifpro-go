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

func (s *appstoreSubscriptionService) handleDidChangeRenewalPref(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload
	renewalInfo := notification.ResponseBodyV2DecodedPayload.Data.SignedRenewalInfo.JWSRenewalInfoDecodedPayload

	// Validate required fields
	if payload.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in renewal preference change notification")
	}

	if notification.ResponseBodyV2DecodedPayload.Subtype == nil {
		return nil, fmt.Errorf("missing subtype in DID_CHANGE_RENEWAL_PREF notification")
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
	previousProductID := subscription.ProductID
	previousPrice := subscription.Price
	previousCurrency := subscription.Currency

	// Determine event type and update subscription based on subtype
	var eventType models.SubscriptionEventType
	var reason string

	switch *notification.ResponseBodyV2DecodedPayload.Subtype {
	case dto.UPGRADE:
		eventType = models.EventTypeUpgrade
		reason = "User upgraded subscription"
	case dto.DOWNGRADE:
		eventType = models.EventTypeDowngrade
		reason = "User downgraded subscription"
	default:
		return nil, fmt.Errorf("unhandled DID_CHANGE_RENEWAL_PREF subtype: %s",
			*notification.ResponseBodyV2DecodedPayload.Subtype)
	}

	// Update subscription fields
	subscription.ProductID = renewalInfo.ProductId
	subscription.CurrentTransactionID = payload.TransactionId
	subscription.PurchaseDate = payload.PurchaseDate
	subscription.ExpiresDate = payload.ExpiresDate
	subscription.AutoRenewStatus = models.MapAutoRenewStatus(renewalInfo.AutoRenewStatus)
	subscription.Price = renewalInfo.RenewalPrice
	subscription.Currency = renewalInfo.Currency
	subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload

	// Handle offer details if present
	if renewalInfo.OfferIdentifier != "" {
		subscription.OfferIdentifier = &renewalInfo.OfferIdentifier
		subscription.OfferType = &renewalInfo.OfferType
		if renewalInfo.OfferPeriod != "" {
			subscription.OfferDuration = &renewalInfo.OfferPeriod
		}
	}

	// Save subscription updates first
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription updates: %w", err)
	}

	// Save subscription and create event atomically
	if err := subscription.AddEvent(tx, eventType, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add renewal preference change event: %w", err)
	}

	logger.Log.Infof("Processed %s for subscription %s (user: %s, original transaction ID: %s, "+
		"previous product: %s, new product: %s, previous price: %d %s, new price: %d %s)",
		eventType, subscription.ID, subscription.UserID, payload.OriginalTransactionId,
		previousProductID, subscription.ProductID,
		previousPrice, previousCurrency, subscription.Price, subscription.Currency)

	// // Post-processing based on subscription change
	// switch eventType {
	// case models.EventTypeUpgrade:
	// 	if err := s.notificationService.SendUpgradeConfirmation(
	// 		ctx,
	// 		subscription.UserID,
	// 		previousProductID,
	// 		subscription.ProductID,
	// 	); err != nil {
	// 		logger.Log.Warnf("Failed to send upgrade confirmation: %v", err)
	// 	}
	// case models.EventTypeDowngrade:
	// 	if err := s.notificationService.SendDowngradeConfirmation(
	// 		ctx,
	// 		subscription.UserID,
	// 		previousProductID,
	// 		subscription.ProductID,
	// 		subscription.ExpiresDate,
	// 	); err != nil {
	// 		logger.Log.Warnf("Failed to send downgrade confirmation: %v", err)
	// 	}
	// }

	return subscription, nil
}
