package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"subsnotifpro-go/internal/appstore/subscription/models"
	"subsnotifpro-go/internal/appstore/webhooks/dto"
	"subsnotifpro-go/internal/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *appstoreSubscriptionService) handleMigration(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload

	// Validate required fields
	if payload.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in migration notification")
	}

	// Find existing subscription
	subscription, err := s.repo.FindByOriginalTransactionID(ctx, tx, payload.OriginalTransactionId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {

			// Create new subscription for migration

			eventType := models.EventTypeInitialPurchase
			migratedSubscription, err := s.createMigratedSubscription(ctx, tx, payload, notification, userID)

			if err != nil {
				return nil, fmt.Errorf("failed to create migrated subscription: %w", err)
			}

			// Add event with the subscription changes
			reason := fmt.Sprintf("Subscription %s", strings.ToLower(string(eventType)))
			if err := migratedSubscription.AddEvent(tx, eventType, notification, reason); err != nil {
				return nil, fmt.Errorf("failed to add subscription event: %w", err)
			}
			return nil, nil
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

	// Update subscription fields for migration
	previousProductID := subscription.ProductID
	subscription.ProductID = payload.ProductId
	subscription.CurrentTransactionID = payload.TransactionId
	subscription.PurchaseDate = payload.PurchaseDate
	subscription.ExpiresDate = payload.ExpiresDate
	subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload

	// Save subscription and create event atomically
	reason := fmt.Sprintf("Subscription migrated from %s to %s", previousProductID, payload.ProductId)
	if err := subscription.AddEvent(tx, models.EventTypeMigration, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add migration event: %w", err)
	}

	logger.Log.Infof("Processed migration for subscription %s (user: %s, original transaction ID: %s, product: %s -> %s)",
		subscription.ID, subscription.UserID, payload.OriginalTransactionId, previousProductID, payload.ProductId)

	// // Post-processing
	// if err := s.migrationService.HandleProductMigration(
	//     ctx,
	//     subscription.UserID,
	//     previousProductID,
	//     payload.ProductId,
	// ); err != nil {
	//     logger.Log.Errorf("Failed to handle product migration: %v", err)
	//     return fmt.Errorf("failed to handle product migration: %w", err)
	// }

	return subscription, nil
}

func (s *appstoreSubscriptionService) createMigratedSubscription(
	ctx context.Context,
	tx *gorm.DB,
	payload dto.JWSTransactionDecodedPayload,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	// Create new subscription with migrated data
	subscription := &models.AppStoreSubscription{
		OriginalTransactionID: payload.OriginalTransactionId,
		CurrentTransactionID:  payload.TransactionId,
		ProductID:             payload.ProductId,
		SubscriptionGroupID:   payload.SubscriptionGroupIdentifier,
		Status:                models.SubscriptionStatusActive,
		Environment:           models.Environment(payload.Environment),
		PurchaseDate:          payload.PurchaseDate,
		ExpiresDate:           payload.ExpiresDate,
		OriginalPurchaseDate:  payload.OriginalPurchaseDate,
		Currency:              payload.Currency,
		Price:                 payload.Price,
		CountryCode:           payload.Storefront,
		InAppOwnershipType:    models.InAppOwnershipType(payload.InAppOwnershipType),
		LatestRawData:         &notification.ResponseBodyV2DecodedPayload,
	}

	if userID != nil {
		subscription.UserID = *userID
	}

	// Create subscription in database
	if err := tx.Create(subscription).Error; err != nil {
		return nil, fmt.Errorf("failed to create migrated subscription: %w", err)
	}

	// Add migration event
	reason := fmt.Sprintf("New subscription created from migration (product: %s)", payload.ProductId)
	if err := subscription.AddEvent(tx, models.EventTypeMigration, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add migration event: %w", err)
	}

	logger.Log.Infof("Created new migrated subscription (original transaction ID: %s, product: %s, user: %v)",
		payload.OriginalTransactionId, payload.ProductId, userID)

	return subscription, nil
}
