package service

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *playstoreSubscriptionService) RecordAckStateTransitionForCreateBeforeAckApiCall(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	newState models.AcknowledgementState,
	changeEventID uuid.UUID,
	packageName string,
	subscriptionProductId string,
	purchaseToken string,
) error {
	history := models.AcknowledgementStateTransitionHistory{
		SubscriptionID: subscriptionID,
		PreviousState:  nil,
		CurrentState:   newState,
		ChangeEventID:  changeEventID,
		Reason:         "New Subscription before acknowledge api call",
	}
	if err := s.repo.InsertAcknowledgementStateTransition(ctx, tx, &history); err != nil {
		return fmt.Errorf("failed to insert acknowledgement transition: %w", err)
	}

	return nil
}

func (s *playstoreSubscriptionService) RecordAckStateTransitionForCreateAfterAckApiCall(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	previousState models.AcknowledgementState,
	refreshedState models.AcknowledgementState,
	changeEventID uuid.UUID,
	packageName string,
	subscriptionProductId string,
	purchaseToken string,
) error {

	// 2️⃣ If existing subscription exists and state changed → record transition

	history := models.AcknowledgementStateTransitionHistory{
		SubscriptionID: subscriptionID,
		PreviousState:  &previousState,
		CurrentState:   refreshedState,
		ChangeEventID:  changeEventID,
		Reason:         "New Subscription after acknowledge api call",
	}
	if err := s.repo.InsertAcknowledgementStateTransition(ctx, tx, &history); err != nil {
		return fmt.Errorf("failed to insert acknowledgement transition: %w", err)
	}

	return nil
}

func (s *playstoreSubscriptionService) AcknowledgeSubscription(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	currentState models.AcknowledgementState,
	changeEventID uuid.UUID,
	packageName string,
	subscriptionProductId string,
	purchaseToken string,
) (*models.AcknowledgementState, error) {
	// 1. Check if acknowledgement is needed
	if currentState != models.AcknowledgementStatePending &&
		currentState != models.AcknowledgementStateUnspecified {
		return nil, nil // No acknowledgement needed
	}

	// 3. Perform acknowledgement via Play API
	if err := s.playstoreApiService.AcknowledgeSubscription(
		ctx,
		packageName,
		subscriptionProductId,
		purchaseToken,
	); err != nil {
		return nil, fmt.Errorf("failed to acknowledge subscription: %w", err)
	}

	// 4. Fetch refreshed state from Google Play
	refreshedSub, err := s.playstoreApiService.GetUserSubscriptionPurchase(
		ctx,
		purchaseToken,
		packageName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch refreshed subscription: %w", err)
	}

	// 5. Parse the new acknowledgement state
	refreshedState := models.AcknowledgementState(refreshedSub.AcknowledgementState)
	if !refreshedState.IsValid() {
		return nil, fmt.Errorf("invalid acknowledgement state received: %s", refreshedSub.AcknowledgementState)
	}

	// 6. Record post-acknowledgement state transition
	if err := s.RecordAckStateTransitionForCreateAfterAckApiCall(
		ctx,
		tx,
		subscriptionID,
		currentState,   // previous state
		refreshedState, // new state
		changeEventID,
		packageName,
		subscriptionProductId,
		purchaseToken,
	); err != nil {
		return nil, fmt.Errorf("failed to record post-acknowledgement state: %w", err)
	}

	return &refreshedState, nil
}
