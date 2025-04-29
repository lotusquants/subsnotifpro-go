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

func (s *appstoreSubscriptionService) handleOfferRedeemed(
	ctx context.Context,
	tx *gorm.DB,
	notification *dto.AppStoreNotification,
	userID *uuid.UUID,
) (*models.AppStoreSubscription, error) {
	payload := notification.ResponseBodyV2DecodedPayload.Data.SignedTransactionInfo.JWSTransactionDecodedPayload

	// Validate required fields
	if payload.OriginalTransactionId == "" {
		return nil, fmt.Errorf("missing original transaction ID in offer redeemed notification")
	}

	if notification.ResponseBodyV2DecodedPayload.Subtype == nil {
		return nil, fmt.Errorf("missing subtype in OFFER_REDEEMED notification")
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

	// Determine offer type
	var offerType string
	switch *notification.ResponseBodyV2DecodedPayload.Subtype {
	case dto.INITIAL_BUY:
		offerType = "initial purchase"
	case dto.RESUBSCRIBE:
		offerType = "resubscribe"
	case dto.UPGRADE:
		offerType = "upgrade"
	case dto.DOWNGRADE:
		offerType = "downgrade"
	default:
		offerType = "unknown"
	}

	// Update subscription fields with offer details
	subscription.OfferIdentifier = &payload.OfferIdentifier
	subscription.OfferType = &payload.OfferType
	if payload.OfferPeriod != "" {
		subscription.OfferDuration = &payload.OfferPeriod
	}
	subscription.LatestRawData = &notification.ResponseBodyV2DecodedPayload

	// Save subscription updates
	if err := s.repo.Update(ctx, tx, subscription); err != nil {
		return nil, fmt.Errorf("failed to save subscription updates: %w", err)
	}

	// Save subscription and create event atomically
	reason := fmt.Sprintf("Offer redeemed (%s), ID: %s", offerType, payload.OfferIdentifier)
	if err := subscription.AddEvent(tx, models.EventTypeOfferCodeRedeem, notification, reason); err != nil {
		return nil, fmt.Errorf("failed to add offer redeemed event: %w", err)
	}

	logger.Log.Infof("Processed offer redemption for subscription %s (user: %s, original transaction ID: %s, offer: %s, type: %s)",
		subscription.ID, subscription.UserID, payload.OriginalTransactionId, payload.OfferIdentifier, offerType)

	// // Post-processing
	// if err := s.offerService.TrackOfferRedemption(
	//     ctx,
	//     subscription.UserID,
	//     payload.OfferIdentifier,
	//     offerType,
	// ); err != nil {
	//     logger.Log.Warnf("Failed to track offer redemption: %v", err)
	// }

	return subscription, nil
}
