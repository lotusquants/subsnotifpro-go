package service

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/appstore/subscription/models"
	"subsnotifpro-go/internal/appstore/webhooks/dto"
	"subsnotifpro-go/internal/pkg/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *appstoreSubscriptionService) handleOneTimeCharge(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload

	// Validate required fields
	if payload.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in one-time charge notification")
	}
	if payload.ProductId == "" {
		return nil, fmt.Errorf("missing product ID in one-time charge notification")
	}

	// Check if this is a Family Sharing purchase
	isFamilyShared := payload.InAppOwnershipType == string(models.OwnershipFamilyShared)

	// For one-time charges, we typically create a new record each time
	oneTimePurchase := &models.AppStoreSubscription{
		OriginalTransactionID: payload.OriginalTransactionId,
		CurrentTransactionID:  payload.TransactionId,
		ProductID:             payload.ProductId,
		Status:                models.SubscriptionStatusActive, // Treat as active for the product's duration
		Environment:           models.Environment(payload.Environment),
		PurchaseDate:          payload.PurchaseDate,
		ExpiresDate:           payload.ExpiresDate, // For non-consumables, this might be far future
		OriginalPurchaseDate:  payload.OriginalPurchaseDate,
		Currency:              payload.Currency,
		Price:                 payload.Price,
		CountryCode:           payload.Storefront,
		InAppOwnershipType:    models.InAppOwnershipType(payload.InAppOwnershipType),
		LatestRawData:         &notification.ResponseBodyV2DecodedPayload,
	}

	if userID != nil {
		oneTimePurchase.UserID = *userID
	}

	// Create record in database
	if err := tx.Create(oneTimePurchase).Error; err != nil {
		return nil, fmt.Errorf("failed to create one-time charge record: %w", err)
	}

	// Determine event type based on context
	eventType := models.EventTypeOneTimeCharge
	reason := "One-time charge processed"
	if isFamilyShared {
		reason = "Family shared one-time charge processed"
	}

	// Add event
	if err := oneTimePurchase.AddEvent(tx, eventType, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add one-time charge event: %w", err)
	}

	logger.Log.Infof("Processed one-time charge (transaction ID: %s, product: %s, user: %v, family shared: %t)",
		payload.TransactionId, payload.ProductId, userID, isFamilyShared)

	// // Post-processing
	// if err := s.entitlementService.GrantOneTimePurchase(
	//     ctx,
	//     oneTimePurchase.UserID,
	//     payload.ProductId,
	//     payload.ExpiresDate,
	//     isFamilyShared,
	// ); err != nil {
	//     logger.Log.Errorf("Failed to grant one-time purchase entitlements: %v", err)
	//     return fmt.Errorf("failed to grant entitlements: %w", err)
	// }

	return nil, nil
}
