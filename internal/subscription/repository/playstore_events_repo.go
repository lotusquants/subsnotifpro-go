package repository

import (
	"context"
	"fmt"

	"subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository methods needed (add these to your playstore repository)
type PlaystoreEventRepository interface {
	GetBaseEvents(ctx context.Context, subscriptionID uuid.UUID, page, pageSize int) ([]models.SubscriptionEvent, int64, error)
	GetLineItemHistories(ctx context.Context, eventIDs []uuid.UUID) ([]models.SubscriptionLineItemHistory, error)
	GetAutoRenewHistories(ctx context.Context, eventIDs []uuid.UUID) ([]models.AutoRenewingPlanHistory, error)
	GetOfferHistories(ctx context.Context, eventIDs []uuid.UUID) ([]models.OfferDetailsHistory, error)
}

type playstoreEventRepository struct {
	db *gorm.DB
}

func NewPlaystoreEventRepository(db *gorm.DB) PlaystoreEventRepository {
	return &playstoreEventRepository{db: db}
}

func (r *playstoreEventRepository) GetBaseEvents(
	ctx context.Context,
	subscriptionID uuid.UUID,
	page, pageSize int,
) ([]models.SubscriptionEvent, int64, error) {
	var events []models.SubscriptionEvent
	var total int64

	offset := (page - 1) * pageSize

	// Get paginated base events
	err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&events).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get base events: %w", err)
	}

	// Get total count
	err = r.db.WithContext(ctx).
		Model(&models.SubscriptionEvent{}).
		Where("subscription_id = ?", subscriptionID).
		Count(&total).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to count base events: %w", err)
	}

	return events, total, nil
}

func (r *playstoreEventRepository) GetLineItemHistories(
	ctx context.Context,
	eventIDs []uuid.UUID,
) ([]models.SubscriptionLineItemHistory, error) {
	var histories []models.SubscriptionLineItemHistory

	if len(eventIDs) == 0 {
		return histories, nil
	}

	err := r.db.WithContext(ctx).
		Where("change_event_id IN ?", eventIDs).
		Order("changed_at DESC").
		Find(&histories).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get line item histories: %w", err)
	}

	return histories, nil
}

func (r *playstoreEventRepository) GetAutoRenewHistories(
	ctx context.Context,
	eventIDs []uuid.UUID,
) ([]models.AutoRenewingPlanHistory, error) {
	var histories []models.AutoRenewingPlanHistory

	if len(eventIDs) == 0 {
		return histories, nil
	}

	err := r.db.WithContext(ctx).
		Where("change_event_id IN ?", eventIDs).
		Order("created_at DESC").
		Find(&histories).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get auto renew histories: %w", err)
	}

	return histories, nil
}

func (r *playstoreEventRepository) GetOfferHistories(
	ctx context.Context,
	eventIDs []uuid.UUID,
) ([]models.OfferDetailsHistory, error) {
	var histories []models.OfferDetailsHistory

	if len(eventIDs) == 0 {
		return histories, nil
	}

	// Select only fields needed for unified event model
	err := r.db.WithContext(ctx).
		Where("change_event_id IN ?", eventIDs).
		Order("changed_at DESC").
		Select([]string{
			"id",
			"offer_details_id",
			"subscription_id",
			"line_item_id",
			"base_plan_id",
			"offer_id",
			"current_base_plan_price",
			"current_phase_price",
			"change_type",
			"change_event_id",
			"changed_at",
		}).
		Find(&histories).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get offer histories: %w", err)
	}

	return histories, nil
}
