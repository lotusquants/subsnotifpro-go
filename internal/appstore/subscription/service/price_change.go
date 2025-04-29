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

func (s *appstoreSubscriptionService) handlePriceChange(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	renewalInfo := notification.ResponseBodyV2DecodedPayload.Data.SignedRenewalInfo.JWSRenewalInfoDecodedPayload

	// Validate required fields
	if renewalInfo.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in price change notification")
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

	// Save subscription and create event atomically
	reason := fmt.Sprintf("Price changed from %d %s to %d %s",
		previousPrice, previousCurrency, renewalInfo.RenewalPrice, renewalInfo.Currency)

	// Save subscription updates
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription updates: %w", err)
	}

	if err := subscription.AddEvent(tx, models.EventTypePriceChange, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add price change event: %w", err)
	}

	logger.Log.Infof("Processed price change for subscription %s (user: %s, original transaction ID: %s, %d %s -> %d %s)",
		subscription.ID, subscription.UserID, renewalInfo.OriginalTransactionId,
		previousPrice, previousCurrency, renewalInfo.RenewalPrice, renewalInfo.Currency)

	// // Post-processing
	// if err := s.notificationService.SendPriceChangeNotification(
	//     ctx,
	//     subscription.UserID,
	//     previousPrice,
	//     previousCurrency,
	//     renewalInfo.RenewalPrice,
	//     renewalInfo.Currency,
	// ); err != nil {
	//     logger.Log.Warnf("Failed to send price change notification: %v", err)
	// }

	return subscription, nil
}
