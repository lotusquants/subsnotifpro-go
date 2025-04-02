package service

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *playstoreSubscriptionService) RecordSubscriptionStateChange(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	existing *models.SubscriptionPurchaseV2,
	newState models.SubscriptionState,
	changeEventID uuid.UUID,
) error {
	// Validate the new state first
	if !newState.IsValid() {
		return fmt.Errorf("invalid subscription state: %s", newState)
	}

	var (
		previousState *models.SubscriptionState
		reason        string
	)

	// Determine transition context
	if existing == nil {
		// New subscription case
		reason = "New subscription created"
		// Previous state is nil (implicitly)
	} else {
		// Existing subscription case
		currentState := existing.SubscriptionState
		if currentState == newState {
			// No state change, nothing to record
			return nil
		}
		previousState = &currentState
		reason = "Subscription state updated"
	}

	// Record the transition
	history := models.SubscriptionStateTransitionHistory{
		SubscriptionID: subscriptionID,
		PreviousState:  previousState,
		CurrentState:   newState,
		Reason:         reason,
		ChangeEventID:  changeEventID,
	}

	if err := s.repo.InsertSubscriptionStateTransition(ctx, tx, &history); err != nil {
		return fmt.Errorf("failed to insert state transition: %w", err)
	}

	return nil
}
