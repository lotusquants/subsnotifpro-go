package repository

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/appstore/subscription/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AppStoreEventRepository interface {
	GetBaseEvents(
		ctx context.Context,
		subscriptionID uuid.UUID,
		page, pageSize int,
	) ([]models.AppStoreSubscriptionEvent, int64, error)
}

type appStoreEventRepository struct {
	db *gorm.DB
}

func NewAppStoreEventRepository(db *gorm.DB) AppStoreEventRepository {
	return &appStoreEventRepository{db: db}
}

func (r *appStoreEventRepository) GetBaseEvents(
	ctx context.Context,
	subscriptionID uuid.UUID,
	page, pageSize int,
) ([]models.AppStoreSubscriptionEvent, int64, error) {
	var events []models.AppStoreSubscriptionEvent
	var total int64

	offset := (page - 1) * pageSize

	// Get paginated events with essential fields
	err := r.db.WithContext(ctx).
		Where("subscription_id = ?", subscriptionID).
		Order("event_date DESC").
		Offset(offset).
		Limit(pageSize).
		Select([]string{
			"id",
			"subscription_id",
			"type",
			"event_date",
			"created_at",
			"status",
			"product_id",
			"price",
			"currency",
			"offer_identifier",
			"expires_date",
		}).
		Find(&events).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get base events: %w", err)
	}

	// Get total count
	err = r.db.WithContext(ctx).
		Model(&models.AppStoreSubscriptionEvent{}).
		Where("subscription_id = ?", subscriptionID).
		Count(&total).Error

	return events, total, err
}
