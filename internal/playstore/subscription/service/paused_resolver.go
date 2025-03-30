package service

import (
	"context"
	"fmt"
	"log"
	"time"

	rtdnModels "subsnotifpro-go/internal/playstore/rtdn/models"
	"subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"google.golang.org/api/androidpublisher/v3"
	"gorm.io/gorm"
)

func (s *playstoreSubscriptionService) ResolvePausedContext(
	ctx context.Context,
	tx *gorm.DB,
	existingSubID *uuid.UUID,
	existing *models.SubscriptionPurchaseV2,
	subData *androidpublisher.SubscriptionPurchaseV2,
	changeEventID uuid.UUID,
	notificationType rtdnModels.SubscriptionNotificationType,
) (*uuid.UUID, error) {
	subscriptionID := *existingSubID
	pausedContext := existing.SubscriptionPausedContext

	switch notificationType {
	case rtdnModels.SubscriptionPauseScheduleChanged:
		return s.handlePauseScheduleChanged(ctx, tx, pausedContext, subscriptionID, subData, changeEventID, existing.SubscriptionStateModel.State)

	case rtdnModels.SubscriptionPaused:
		return s.handlePaused(ctx, tx, pausedContext, subscriptionID, subData, changeEventID, existing.SubscriptionStateModel.State)

	case rtdnModels.SubscriptionRenewed:
		return s.handleResumed(ctx, tx, pausedContext, subscriptionID, changeEventID, existing.SubscriptionStateModel.State)

	case rtdnModels.SubscriptionExpired, rtdnModels.SubscriptionRevoked:
		return s.handleExpired(ctx, tx, pausedContext, subscriptionID, changeEventID, existing.SubscriptionStateModel.State)
	}

	return nil, nil
}

func (s *playstoreSubscriptionService) handlePauseScheduleChanged(
	ctx context.Context,
	tx *gorm.DB,
	existing *models.SubscriptionPausedContext,
	subscriptionID uuid.UUID,
	subData *androidpublisher.SubscriptionPurchaseV2,
	changeEventID uuid.UUID,
	subState models.SubscriptionState,
) (*uuid.UUID, error) {

	autoPauseTimeStr := subData.LineItems[0].ExpiryTime
	autoPauseTime, err := time.Parse(time.RFC3339Nano, autoPauseTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid autoResumeTime format: %w", err)
	}

	autoResumeTimeStr := subData.PausedStateContext.AutoResumeTime
	autoResumeTime, err := time.Parse(time.RFC3339Nano, autoResumeTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid autoResumeTime format: %w", err)
	}

	if existing != nil {
		existing.ScheduledAt = time.Now()
		existing.AutoPauseTime = autoPauseTime
		existing.AutoResumeTime = autoResumeTime
		existing.Status = models.PauseContextStatusScheduled
		existing.PauseReason = models.PauseReasonUser

		if err := s.repo.UpdatePausedContext(ctx, tx, existing); err != nil {
			return nil, err
		}
		history := buildPausedContextHistory(existing, subscriptionID, changeEventID, models.PauseChangeScheduleEdit, subState)
		err := s.repo.InsertPausedContextHistory(ctx, tx, history)
		if err != nil {
			log.Printf("Failed to insert paused context history: %v", err)
			return nil, err
		}
		return &existing.ID, nil
	}

	newCtx := &models.SubscriptionPausedContext{
		ID:             uuid.New(),
		SubscriptionID: subscriptionID,
		ScheduledAt:    time.Now(),
		AutoPauseTime:  autoPauseTime,
		AutoResumeTime: autoResumeTime,
		Status:         models.PauseContextStatusScheduled,
		PauseReason:    models.PauseReasonUser,
	}
	if err := s.repo.InsertPausedContext(ctx, tx, newCtx); err != nil {
		return nil, err
	}
	history := buildPausedContextHistory(newCtx, subscriptionID, changeEventID, models.PauseChangeScheduleCreate, subState)

	return &newCtx.ID, s.repo.InsertPausedContextHistory(ctx, tx, history)
}

func (s *playstoreSubscriptionService) handlePaused(
	ctx context.Context,
	tx *gorm.DB,
	existing *models.SubscriptionPausedContext,
	subscriptionID uuid.UUID,
	subData *androidpublisher.SubscriptionPurchaseV2,
	changeEventID uuid.UUID,
	subState models.SubscriptionState,
) (*uuid.UUID, error) {
	now := time.Now()

	if existing == nil {

		autoResumeTimeStr := subData.PausedStateContext.AutoResumeTime
		autoResumeTime, err := time.Parse(time.RFC3339Nano, autoResumeTimeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid autoResumeTime format: %w", err)
		}

		// Fallback: Create new context directly if one doesn't exist
		newCtx := &models.SubscriptionPausedContext{
			ID:             uuid.New(),
			SubscriptionID: subscriptionID,
			ScheduledAt:    now,
			AutoPauseTime:  now, // fallback default
			PausedAt:       ptr(now),
			AutoResumeTime: autoResumeTime,
			Status:         models.PauseContextStatusPaused,
			PauseReason:    models.PauseReasonSystem, // fallback to system
		}
		if err := s.repo.InsertPausedContext(ctx, tx, newCtx); err != nil {
			return nil, err
		}
		history := buildPausedContextHistory(newCtx, subscriptionID, changeEventID, models.PauseChangePausedAuto, subState)
		return &newCtx.ID, s.repo.InsertPausedContextHistory(ctx, tx, history)
	}

	// If context exists, update as usual
	existing.PausedAt = ptr(now)
	existing.Status = models.PauseContextStatusPaused
	existing.PauseReason = models.PauseReasonUser

	if err := s.repo.UpdatePausedContext(ctx, tx, existing); err != nil {
		return nil, err
	}
	history := buildPausedContextHistory(existing, subscriptionID, changeEventID, models.PauseChangePausedAuto, subState)
	return &existing.ID, s.repo.InsertPausedContextHistory(ctx, tx, history)
}

func (s *playstoreSubscriptionService) handleResumed(
	ctx context.Context,
	tx *gorm.DB,
	existing *models.SubscriptionPausedContext,
	subscriptionID uuid.UUID,
	changeEventID uuid.UUID,
	subState models.SubscriptionState,
) (*uuid.UUID, error) {
	if existing == nil {
		return nil, nil
	}
	existing.ResumedAt = ptr(time.Now())
	existing.Status = models.PauseContextStatusResumed
	reason := models.ResumeReasonAuto
	existing.ResumeReason = &reason
	if err := s.repo.UpdatePausedContext(ctx, tx, existing); err != nil {
		return nil, err
	}
	history := buildPausedContextHistory(existing, subscriptionID, changeEventID, models.PauseChangeResumedAuto, subState)
	return &existing.ID, s.repo.InsertPausedContextHistory(ctx, tx, history)
}

func (s *playstoreSubscriptionService) handleExpired(
	ctx context.Context,
	tx *gorm.DB,
	existing *models.SubscriptionPausedContext,
	subscriptionID uuid.UUID,
	changeEventID uuid.UUID,
	subState models.SubscriptionState,
) (*uuid.UUID, error) {
	if existing == nil {
		return nil, nil
	}
	expired := time.Now()
	existing.ExpiredAt = &expired
	existing.Status = models.PauseContextStatusExpired
	if err := s.repo.UpdatePausedContext(ctx, tx, existing); err != nil {
		return nil, err
	}
	history := buildPausedContextHistory(existing, subscriptionID, changeEventID, models.PauseChangeExpired, subState)
	return &existing.ID, s.repo.InsertPausedContextHistory(ctx, tx, history)
}

func buildPausedContextHistory(
	ctx *models.SubscriptionPausedContext,
	subscriptionID uuid.UUID,
	changeEventID uuid.UUID,
	reason models.PauseChangeReason,
	subState models.SubscriptionState,
) *models.SubscriptionPausedContextHistory {
	return &models.SubscriptionPausedContextHistory{
		ID:                uuid.New(),
		PauseContextID:    ctx.ID,
		SubscriptionID:    subscriptionID,
		Status:            ctx.Status,
		ScheduledAt:       ctx.ScheduledAt,
		CancelledAt:       ctx.CancelledAt,
		AutoPauseTime:     ctx.AutoPauseTime,
		PausedAt:          ctx.PausedAt,
		AutoResumeTime:    ctx.AutoResumeTime,
		ResumedAt:         ctx.ResumedAt,
		ExpiredAt:         ctx.ExpiredAt,
		PauseReason:       ctx.PauseReason,
		ResumeReason:      ctx.ResumeReason,
		PauseChangeReason: reason,
		ChangeEventID:     changeEventID,
		SubscriptionState: subState,
		ChangedAt:         time.Now(),
	}
}

func ptr[T any](v T) *T {
	return &v
}
