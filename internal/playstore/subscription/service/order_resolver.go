package service

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *playstoreSubscriptionService) ResolveOrderID(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	existing *models.SubscriptionPurchaseV2,
	newOrderID string,
	changeEventID uuid.UUID,
) (string, error) {
	// No existing subscription: no transition
	if existing == nil {
		return newOrderID, nil
	}

	// If no change, return as is
	if existing.LatestOrderId == newOrderID {
		return newOrderID, nil
	}

	// Log transition history
	history := models.SubscriptionOrderIdTransitionHistory{
		SubscriptionID:  subscriptionID,
		PreviousOrderID: existing.LatestOrderId,
		NewOrderID:      newOrderID,
		Reason:          "Order ID changed by Google RTDN",
		ChangeEventID:   changeEventID,
	}
	if err := s.repo.InsertOrderIDTransition(ctx, tx, &history); err != nil {
		return "", fmt.Errorf("failed to insert order ID transition: %w", err)
	}

	return newOrderID, nil
}
