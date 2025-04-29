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

func (s *appstoreSubscriptionService) handleMetadataUpdate(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload

	// Validate required fields
	if payload.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in metadata update notification")
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

	// Store previous metadata for comparison
	// previousMetadata := subscription.LatestRawData

	// Update subscription fields
	subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload

	// Save subscription and create event atomically

	// Save subscription updates
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription updates: %w", err)
	}
	reason := "Subscription metadata updated by App Store"
	if err := subscription.AddEvent(tx, models.EventTypeMetadataUpdate, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add metadata update event: %w", err)
	}

	logger.Log.Infof("Processed metadata update for subscription %s (user: %s, original transaction ID: %s)",
		subscription.ID, subscription.UserID, payload.OriginalTransactionId)

	// // Post-processing - analyze what changed if needed
	// if err := s.metadataService.ProcessMetadataChanges(
	//     ctx,
	//     previousMetadata,
	//     notification.ResponseBodyV2DecodedPayload,
	// ); err != nil {
	//     logger.Log.Warnf("Failed to process metadata changes: %v", err)
	// }

	return subscription, nil
}
