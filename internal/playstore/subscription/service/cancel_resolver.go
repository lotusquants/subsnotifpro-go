package service

import (
	"context"
	"fmt"
	"time"

	rtdnModels "subsnotifpro-go/internal/playstore/rtdn/models"
	"subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"google.golang.org/api/androidpublisher/v3"
	"gorm.io/gorm"
)

func (s *playstoreSubscriptionService) ResolveCancellationContext(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	existing *models.SubscriptionPurchaseV2,
	subData *androidpublisher.SubscriptionPurchaseV2,
	changeEventID uuid.UUID,
	notificationType rtdnModels.SubscriptionNotificationType,
) (*uuid.UUID, error) {

	existingCtx := existing.SubscriptionCancellationContext

	switch notificationType {
	case rtdnModels.SubscriptionCanceled:
		return s.handleCancelled(ctx, tx, subscriptionID, existingCtx, subData, changeEventID)

	case rtdnModels.SubscriptionRestarted:
		return s.handleResubscribed(ctx, tx, subscriptionID, existingCtx, changeEventID)

	case rtdnModels.SubscriptionExpired, rtdnModels.SubscriptionRevoked:
		// Terminal state, no action for cancellation context
		return nil, nil

	default:
		return nil, nil
	}
}

func (s *playstoreSubscriptionService) handleCancelled(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	existing *models.SubscriptionCancellationContext,
	subData *androidpublisher.SubscriptionPurchaseV2,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {

	cancelTimeStr := subData.CanceledStateContext.UserInitiatedCancellation.CancelTime
	cancelTime, err := time.Parse(time.RFC3339Nano, cancelTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid autoResumeTime format: %w", err)
	}

	cancelReason := models.CancellationReasonUserInitiated
	var cancelSurvey *models.CancelSurveyReason
	var userInput *string

	if survey := subData.CanceledStateContext.UserInitiatedCancellation.CancelSurveyResult; survey != nil {
		s := models.CancelSurveyReason(survey.Reason)
		cancelSurvey = &s
		userInput = &survey.ReasonUserInput
	}

	if existing != nil {
		existing.CancelTime = cancelTime
		existing.CancellationReason = cancelReason
		existing.CancelSurveyReason = cancelSurvey
		existing.CancelSurveyUserInput = userInput

		if err := s.repo.UpdateCancellationContext(ctx, tx, existing); err != nil {
			return nil, err
		}
		history := buildCancellationHistory(existing, subscriptionID, changeEventID)
		return &existing.ID, s.repo.InsertCancellationHistory(ctx, tx, history)
	}

	// First-time cancellation
	newCtx := &models.SubscriptionCancellationContext{
		ID:                    uuid.New(),
		SubscriptionID:        subscriptionID,
		CancelTime:            cancelTime,
		CancellationReason:    cancelReason,
		CancelSurveyReason:    cancelSurvey,
		CancelSurveyUserInput: userInput,
	}
	if err := s.repo.InsertCancellationContext(ctx, tx, newCtx); err != nil {
		return nil, err
	}
	history := buildCancellationHistory(newCtx, subscriptionID, changeEventID)
	return &newCtx.ID, s.repo.InsertCancellationHistory(ctx, tx, history)
}

func (s *playstoreSubscriptionService) handleResubscribed(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	existing *models.SubscriptionCancellationContext,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {
	if existing == nil {
		return nil, nil // No prior cancellation context to update
	}
	now := time.Now()
	existing.ResubscribeTime = &now
	if err := s.repo.UpdateCancellationContext(ctx, tx, existing); err != nil {
		return nil, err
	}
	history := buildCancellationHistory(existing, subscriptionID, changeEventID)
	return &existing.ID, s.repo.InsertCancellationHistory(ctx, tx, history)
}

func buildCancellationHistory(
	src *models.SubscriptionCancellationContext,
	subscriptionID uuid.UUID,
	eventID uuid.UUID,
) *models.SubscriptionCancellationContextHistory {
	return &models.SubscriptionCancellationContextHistory{
		ID:                    uuid.New(),
		CancellationContextID: src.ID,
		SubscriptionID:        subscriptionID,
		CancelTime:            src.CancelTime,
		CancellationReason:    src.CancellationReason,
		CancelSurveyReason:    src.CancelSurveyReason,
		CancelSurveyUserInput: src.CancelSurveyUserInput,
		ResubscribedTime:      src.ResubscribeTime,
		ChangeEventID:         eventID,
		ChangedAt:             time.Now(),
	}
}
