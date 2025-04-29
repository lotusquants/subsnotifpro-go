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

func (s *appstoreSubscriptionService) handleSubscribed(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload

	// Validate notification subtype
	if notification.ResponseBodyV2DecodedPayload.Subtype == nil {
		return nil, fmt.Errorf("missing subtype in SUBSCRIBED notification")
	}

	// Determine event type based on subtype
	var eventType models.SubscriptionEventType
	switch *notification.ResponseBodyV2DecodedPayload.Subtype {
	case dto.INITIAL_BUY:
		eventType = models.EventTypeInitialPurchase
	case dto.RESUBSCRIBE:
		eventType = models.EventTypeResubscribe
	default:
		return nil, fmt.Errorf("unhandled SUBSCRIBED subtype: %s", *notification.ResponseBodyV2DecodedPayload.Subtype)
	}

	// Check if subscription already exists
	existingSub, err := s.repo.FindByOriginalTransactionID(ctx, tx, payload.OriginalTransactionId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing subscription: %w", err)
	}

	var subscription *models.AppStoreSubscription
	if existingSub != nil {
		// Update existing subscription for resubscribe case
		subscription, err = s.updateExistingSubscription(ctx, tx, existingSub, payload, eventType, userID, notification)
		if err != nil {
			return nil, fmt.Errorf("failed to update existing subscription: %w", err)
		}
	} else {
		// Create new subscription for initial purchase
		subscription, err = s.createNewSubscription(ctx, tx, payload, eventType, userID, notification)
		if err != nil {
			return nil, fmt.Errorf("failed to create new subscription: %w", err)
		}
	}

	// Add event with the subscription changes
	reason := fmt.Sprintf("Subscription %s", strings.ToLower(string(eventType)))
	if err := subscription.AddEvent(tx, eventType, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add subscription event: %w", err)
	}

	logger.Log.Infof("Successfully processed %s for subscription %s (original transaction ID: %s, user: %v)",
		eventType, subscription.ID, payload.OriginalTransactionId, userID)

	return subscription, nil
}

func (s *appstoreSubscriptionService) createNewSubscription(
	ctx context.Context,
	tx *gorm.DB,
	payload dto.JWSTransactionDecodedPayload,
	eventType models.SubscriptionEventType,
	userID *uuid.UUID,
	notification *dto.AppStoreNotification,
) (*models.AppStoreSubscription, error) {
	// Create new subscription with initial state
	subscription := &models.AppStoreSubscription{
		OriginalTransactionID: payload.OriginalTransactionId,
		CurrentTransactionID:  payload.TransactionId,
		WebOrderLineItemID:    payload.WebOrderLineItemId,
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
		IsUpgraded:            payload.IsUpgraded,
		LatestRawData:         &notification.ResponseBodyV2DecodedPayload,
		BundleID:              payload.BundleId,
	}

	// Set auto-renew status if renewal info is available
	if notification.ResponseBodyV2DecodedPayload.Data.SignedRenewalInfo != nil {
		renewalInfo := notification.ResponseBodyV2DecodedPayload.Data.SignedRenewalInfo.JWSRenewalInfoDecodedPayload
		subscription.AutoRenewStatus = models.MapAutoRenewStatus(renewalInfo.AutoRenewStatus)
	}

	// Set user ID if available
	if userID != nil {
		subscription.UserID = *userID
	}

	// Handle offer details if present
	if payload.OfferIdentifier != "" {
		subscription.OfferIdentifier = &payload.OfferIdentifier
		subscription.OfferType = &payload.OfferType
		if payload.OfferPeriod != "" {
			subscription.OfferDuration = &payload.OfferPeriod
		}
	}

	// Set app account token if available
	if payload.AppAccountToken != nil {
		subscription.AppAccountToken = payload.AppAccountToken
	}

	// Create subscription in database
	if err := tx.Create(subscription).Error; err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	return subscription, nil
}

func (s *appstoreSubscriptionService) updateExistingSubscription(
	ctx context.Context,
	tx *gorm.DB,
	existingSub *models.AppStoreSubscription,
	payload dto.JWSTransactionDecodedPayload,
	eventType models.SubscriptionEventType,
	userID *uuid.UUID,
	notification *dto.AppStoreNotification,
) (*models.AppStoreSubscription, error) {
	// Update subscription fields
	existingSub.CurrentTransactionID = payload.TransactionId
	existingSub.PurchaseDate = payload.PurchaseDate
	existingSub.ExpiresDate = payload.ExpiresDate
	existingSub.IsUpgraded = payload.IsUpgraded
	existingSub.LatestRawData = &notification.ResponseBodyV2DecodedPayload

	// Update auto-renew status if renewal info is available
	if notification.ResponseBodyV2DecodedPayload.Data.SignedRenewalInfo != nil {
		renewalInfo := notification.ResponseBodyV2DecodedPayload.Data.SignedRenewalInfo.JWSRenewalInfoDecodedPayload
		existingSub.AutoRenewStatus = models.MapAutoRenewStatus(renewalInfo.AutoRenewStatus)
	}

	// Update status based on event type
	existingSub.Status = models.SubscriptionStatusActive

	// Update user ID if provided and not already set
	if userID != nil && existingSub.UserID == uuid.Nil {
		existingSub.UserID = *userID
	}

	// Update offer details if present
	if payload.OfferIdentifier != "" {
		existingSub.OfferIdentifier = &payload.OfferIdentifier
		existingSub.OfferType = &payload.OfferType
		if payload.OfferPeriod != "" {
			duration := payload.OfferPeriod
			existingSub.OfferDuration = &duration
		}
	}

	// Update app account token if available
	if payload.AppAccountToken != nil {
		existingSub.AppAccountToken = payload.AppAccountToken
	}

	// Save updated subscription
	if err := tx.Save(existingSub).Error; err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	return existingSub, nil
}
