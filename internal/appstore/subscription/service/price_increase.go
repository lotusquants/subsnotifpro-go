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

func (s *appstoreSubscriptionService) handlePriceIncrease(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload

	// Validate required fields
	if payload.Data.SignedRenewalInfo == nil {
		return nil, fmt.Errorf("missing renewal info in price increase notification")
	}
	renewalInfo := payload.Data.SignedRenewalInfo.JWSRenewalInfoDecodedPayload

	if renewalInfo.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in price increase notification")
	}

	if payload.Subtype == nil {
		return nil, fmt.Errorf("missing subtype in PRICE_INCREASE notification")
	}

	// Find existing subscription
	subscription, err := s.repo.FindByOriginalTransactionID(ctx, tx, renewalInfo.OriginalTransactionId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("subscription not found for original transaction ID: %s", renewalInfo.OriginalTransactionId)
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

	// Store previous pricing for comparison
	previousPrice := subscription.Price
	previousCurrency := subscription.Currency

	// Update subscription fields
	subscription.Price = renewalInfo.RenewalPrice
	subscription.Currency = renewalInfo.Currency
	subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload

	// Determine event details based on subtype
	var reason string
	switch *payload.Subtype {
	case dto.PENDING:
		reason = "Price increase pending customer consent"
	case dto.ACCEPTED:
		reason = "Customer accepted price increase"
	default:
		reason = "Price increase notification"
	}

	// Save subscription updates
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription updates: %w", err)
	}

	// Save subscription and create event atomically
	if err := subscription.AddEvent(tx, models.EventTypePriceIncrease, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add price increase event: %w", err)
	}

	logger.Log.Infof("Processed price increase for subscription %s (user: %s, original transaction ID: %s, subtype: %s, %d %s -> %d %s)",
		subscription.ID, subscription.UserID, renewalInfo.OriginalTransactionId, *payload.Subtype,
		previousPrice, previousCurrency, renewalInfo.RenewalPrice, renewalInfo.Currency)

	// // Post-processing based on subtype
	// switch *payload.Subtype {
	// case dto.PENDING:
	//     if err := s.notificationService.SendPriceIncreasePendingNotification(
	//         ctx,
	//         subscription.UserID,
	//         renewalInfo.RenewalPrice,
	//         renewalInfo.Currency,
	//     ); err != nil {
	//         logger.Log.Warnf("Failed to send price increase pending notification: %v", err)
	//     }
	// case dto.ACCEPTED:
	//     if err := s.notificationService.SendPriceIncreaseAcceptedNotification(
	//         ctx,
	//         subscription.UserID,
	//         renewalInfo.RenewalPrice,
	//         renewalInfo.Currency,
	//     ); err != nil {
	//         logger.Log.Warnf("Failed to send price increase accepted notification: %v", err)
	//     }
	// }

	return subscription, nil
}
