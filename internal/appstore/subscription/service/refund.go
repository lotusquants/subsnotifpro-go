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

func (s *appstoreSubscriptionService) handleRefund(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload

	// Validate required fields
	if payload.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in refund notification")
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

	// Update subscription fields
	subscription.Status = models.SubscriptionStatusRevoked
	subscription.RevocationDate = payload.RevocationDate
	subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload
	subscription.AutoRenewStatus = models.AutoRenewOff

	// Save subscription updates
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription updates: %w", err)
	}

	// Save subscription and create event atomically
	reason := "Subscription refunded by Apple"
	if err := subscription.AddEvent(tx, models.EventTypeRefund, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add refund event: %w", err)
	}

	logger.Log.Infof("Processed refund for subscription %s (user: %s, original transaction ID: %s, "+
		"previous status: %s, revocation date: %v)",
		subscription.ID, subscription.UserID, payload.OriginalTransactionId,
		previousStatus, payload.RevocationDate)

	// // Post-processing
	// if err := s.notificationService.SendRefundNotification(
	// 	ctx,
	// 	subscription.UserID,
	// 	payload.RevocationDate,
	// ); err != nil {
	// 	logger.Log.Warnf("Failed to send refund notification: %v", err)
	// }

	// // Additional business logic for refund handling
	// if err := s.accessService.RevokeAccess(
	// 	ctx,
	// 	subscription.UserID,
	// 	subscription.ProductID,
	// ); err != nil {
	// 	logger.Log.Errorf("Failed to revoke access after refund: %v", err)
	// 	return fmt.Errorf("failed to revoke access: %w", err)
	// }

	return subscription, nil
}
