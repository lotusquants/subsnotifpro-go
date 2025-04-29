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

func (s *appstoreSubscriptionService) handleRenewalExtended(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload

	// Validate required fields
	if payload.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in renewal extended notification")
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

	// Store previous expiration date
	previousExpiresDate := subscription.ExpiresDate

	// Update subscription fields
	subscription.ExpiresDate = payload.ExpiresDate
	subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload

	// Save subscription updates
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription updates: %w", err)
	}

	// Save subscription and create event atomically
	reason := fmt.Sprintf("Renewal extended from %s to %s", previousExpiresDate, payload.ExpiresDate)
	if err := subscription.AddEvent(tx, models.EventTypeServiceExtension, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add renewal extended event: %w", err)
	}

	logger.Log.Infof("Processed renewal extension for subscription %s (user: %s, original transaction ID: %s, extended to %s)",
		subscription.ID, subscription.UserID, payload.OriginalTransactionId, payload.ExpiresDate)

	// // Post-processing
	// if err := s.notificationService.SendRenewalExtendedNotification(
	//     ctx,
	//     subscription.UserID,
	//     previousExpiresDate,
	//     payload.ExpiresDate,
	// ); err != nil {
	//     logger.Log.Warnf("Failed to send renewal extended notification: %v", err)
	// }

	return subscription, nil
}
