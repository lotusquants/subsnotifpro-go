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

func (s *appstoreSubscriptionService) handleExternalPurchaseToken(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	externalPurchaseToken := notification.ResponseBodyV2DecodedPayload.ExternalPurchaseToken
	payload := notification.ResponseBodyV2DecodedPayload.Data

	// Validate required fields
	if payload.BundleID == "" {
		return nil, fmt.Errorf("missing bundle ID in external purchase token notification")
	}

	// Try to find existing subscription by external purchase ID
	subscription, err := s.repo.FindByOriginalTransactionID(ctx, tx, *externalPurchaseToken)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing subscription: %w", err)
	}

	// If subscription exists, update it
	if subscription != nil {
		// Update token and dates
		subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload
		if err := tx.Save(subscription).Error; err != nil {
			return nil, fmt.Errorf("failed to update subscription with external purchase token: %w", err)
		}

		logger.Log.Infof("Updated external purchase token for existing subscription %s (external purchase ID: %s)",
			subscription.ID, *externalPurchaseToken)
	} else {
		// Create minimal subscription record for tracking
		newSubscription := &models.AppStoreSubscription{
			OriginalTransactionID: *externalPurchaseToken,

			BundleID:      payload.BundleID,
			ProductID:     payload.SignedTransactionInfo.JWSTransactionDecodedPayload.ProductId,
			Status:        models.SubscriptionStatusActive, // Assuming active until proven otherwise
			Environment:   models.Environment(payload.Environment),
			PurchaseDate:  payload.SignedTransactionInfo.JWSTransactionDecodedPayload.PurchaseDate,
			LatestRawData: &notification.ResponseBodyV2DecodedPayload,
		}

		if userID != nil {
			newSubscription.UserID = *userID
		}

		if err := tx.Create(newSubscription).Error; err != nil {
			return nil, fmt.Errorf("failed to create subscription for external purchase: %w", err)
		}

		logger.Log.Infof("Created new subscription record for external purchase (external purchase ID: %s)",
			*externalPurchaseToken)
	}

	// Create event for tracking
	eventType := models.EventTypeResubscribe

	reason := "Received external purchase token notification"
	if subscription != nil {

		// Save subscription updates
		if err := s.repo.Update(ctx, tx, subscription); err != nil {
			return nil, fmt.Errorf("failed to save subscription updates: %w", err)
		}

		if err := subscription.AddEvent(tx, eventType, notification, reason); err != nil {
			return nil, fmt.Errorf("failed to add external purchase token event: %w", err)
		}
	}

	// // Post-processing - typically you'd validate the token with Apple
	// if err := s.appStoreClient.ValidateExternalPurchaseToken(
	// 	ctx,
	// 	externalPurchaseToken.Token,
	// ); err != nil {
	// 	logger.Log.Errorf("Failed to validate external purchase token: %v", err)
	// 	return fmt.Errorf("failed to validate external purchase token: %w", err)
	// }

	return subscription, nil
}
