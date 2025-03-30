package service

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *playstoreSubscriptionService) ResolveSubscriptionState(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	existing *models.SubscriptionPurchaseV2,
	newStateEnum models.SubscriptionState,
	changeEventID uuid.UUID,
) (uuid.UUID, error) {
	newState := models.SubscriptionState(newStateEnum)

	// 1️⃣ Ensure the new state model exists (insert if not found)
	newStateModelID, err := s.repo.GetOrCreateSubscriptionStateModel(ctx, tx, newState)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get/create state model: %w", err)
	}

	// 2️⃣ If no existing subscription, no transition history needed
	if existing == nil {
		return newStateModelID, nil
	}

	// 3️⃣ Compare old vs new state
	if existing.SubscriptionStateModelID == newStateModelID {
		return newStateModelID, nil // ✅ No change
	}

	// 4️⃣ Record the transition
	history := models.SubscriptionStateTransitionHistory{
		SubscriptionID:  existing.ID,
		PreviousStateID: existing.SubscriptionStateModelID,
		CurrentStateID:  newStateModelID,
		Reason:          "State changed by Google RTDN",
		ChangeEventID:   changeEventID,
	}

	if err := s.repo.InsertSubscriptionStateTransition(ctx, tx, &history); err != nil {
		return uuid.Nil, fmt.Errorf("failed to insert state transition: %w", err)
	}

	// ✅ Done
	return newStateModelID, nil
}

func (s *playstoreSubscriptionService) ResolveAcknowledgementState(
	ctx context.Context,
	tx *gorm.DB,
	existingSubID *uuid.UUID,
	existing *models.SubscriptionPurchaseV2,
	incomingState string,
	changeEventID uuid.UUID,
	subscriptionProductId string,
) (uuid.UUID, error) {
	newState := models.AcknowledgementState(incomingState)

	// 1️⃣ Get or create state model
	newStateID, err := s.repo.GetOrCreateAcknowledgementStateModel(ctx, tx, newState)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get/create acknowledgement state model: %w", err)
	}

	// 2️⃣ If existing subscription exists and state changed → record transition
	if existing != nil && existing.AcknowledgementStateModelID != newStateID {
		history := models.AcknowledgementStateTransitionHistory{
			SubscriptionID:  existing.ID,
			PreviousStateID: existing.AcknowledgementStateModelID,
			CurrentStateID:  newStateID,
			ChangeEventID:   changeEventID,
			Reason:          "Acknowledgement state changed by RTDN",
		}
		if err := s.repo.InsertAcknowledgementStateTransition(ctx, tx, &history); err != nil {
			return uuid.Nil, fmt.Errorf("failed to insert acknowledgement transition: %w", err)
		}
	}

	// 3️⃣ If acknowledgement is pending or unspecified → trigger API call
	if newState == models.AcknowledgementStatePending || newState == models.AcknowledgementStateUnspecified {
		if existing == nil {
			return newStateID, nil // We need an existing subscription to acknowledge
		}

		// 🔁 Acknowledge via Play API
		err := s.playstoreApiService.AcknowledgeSubscription(ctx, existing.PackageName, subscriptionProductId, existing.PurchaseToken)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to acknowledge subscription: %w", err)
		}

		// 🔁 Fetch updated data
		refreshed, err := s.playstoreApiService.GetUserSubscriptionPurchase(ctx, existing.PurchaseToken, existing.PackageName)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to fetch subscription after ack: %w", err)
		}

		// ✅ Resolve updated state
		ackState := models.AcknowledgementState(refreshed.AcknowledgementState)
		ackStateID, err := s.repo.GetOrCreateAcknowledgementStateModel(ctx, tx, ackState)
		if err != nil {
			return uuid.Nil, fmt.Errorf("failed to get/create ack state after ack: %w", err)
		}

		// ➕ Record transition (acknowledged state)
		history := models.AcknowledgementStateTransitionHistory{
			SubscriptionID:  existing.ID,
			PreviousStateID: newStateID,
			CurrentStateID:  ackStateID,
			ChangeEventID:   changeEventID,
			Reason:          "State changed after successful acknowledgement",
		}
		if err := s.repo.InsertAcknowledgementStateTransition(ctx, tx, &history); err != nil {
			return uuid.Nil, fmt.Errorf("failed to record post-ack state: %w", err)
		}

		return ackStateID, nil
	}

	// ✅ Done
	return newStateID, nil
}
