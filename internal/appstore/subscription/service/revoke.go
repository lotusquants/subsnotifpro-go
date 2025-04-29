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

func (s *appstoreSubscriptionService) handleRevoke(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload

	// Validate required fields
	if payload.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in revoke notification")
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
	previousOwnershipType := subscription.InAppOwnershipType

	// Update subscription fields
	subscription.InAppOwnershipType = models.OwnershipPurchased // No longer family shared
	subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload

	// Only mark as revoked if this was the only access method
	if subscription.InAppOwnershipType == models.OwnershipFamilyShared {
		subscription.Status = models.SubscriptionStatusRevoked
	}

	// Save subscription updates
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription updates: %w", err)
	}

	// Save subscription and create event atomically
	reason := "Family Sharing access revoked"
	if err := subscription.AddEvent(tx, models.EventTypeRevocation, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add revocation event: %w", err)
	}

	logger.Log.Infof("Processed revocation for subscription %s (user: %s, original transaction ID: %s, "+
		"previous status: %s, previous ownership: %s)",
		subscription.ID, subscription.UserID, payload.OriginalTransactionId,
		previousStatus, previousOwnershipType)

	// // Post-processing
	// if err := s.notificationService.SendFamilySharingRevokedNotification(
	// 	ctx,
	// 	subscription.UserID,
	// ); err != nil {
	// 	logger.Log.Warnf("Failed to send family sharing revoked notification: %v", err)
	// }

	// // Revoke access if this was the only access method
	// if previousOwnershipType == models.OwnershipFamilyShared {
	// 	if err := s.accessService.RevokeAccess(
	// 		ctx,
	// 		subscription.UserID,
	// 		subscription.ProductID,
	// 	); err != nil {
	// 		logger.Log.Errorf("Failed to revoke access after family sharing revocation: %v", err)
	// 		return fmt.Errorf("failed to revoke access: %w", err)
	// 	}
	// }

	return subscription, nil
}
