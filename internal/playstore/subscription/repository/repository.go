// ------------------------------
// 📦 Repository: google playstore subscription repository
// ------------------------------
package repository

import (
	"context"
	"errors"
	"fmt"

	"subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PlaystoreSubscriptionRepository interface {
	GetSubscriptionByPurchaseToken(ctx context.Context, tx *gorm.DB, token string) (*models.SubscriptionPurchaseV2, error)
	InsertSubscription(ctx context.Context, tx *gorm.DB, sub *models.SubscriptionPurchaseV2) error
	UpdateSubscriptionFields(ctx context.Context, tx *gorm.DB, subscriptionID uuid.UUID, updateFields map[string]interface{}) error

	InsertSubscriptionStateTransition(ctx context.Context, tx *gorm.DB, history *models.SubscriptionStateTransitionHistory) error
	InsertOrderIDTransition(ctx context.Context, tx *gorm.DB, history *models.SubscriptionOrderIdTransitionHistory) error

	InsertAcknowledgementStateTransition(ctx context.Context, tx *gorm.DB, history *models.AcknowledgementStateTransitionHistory) error

	UpdatePausedContext(ctx context.Context, tx *gorm.DB, context *models.SubscriptionPausedContext) error
	InsertPausedContext(ctx context.Context, tx *gorm.DB, context *models.SubscriptionPausedContext) error
	InsertPausedContextHistory(ctx context.Context, tx *gorm.DB, history *models.SubscriptionPausedContextHistory) error

	InsertCancellationContext(ctx context.Context, tx *gorm.DB, cancelCtx *models.SubscriptionCancellationContext) error
	UpdateCancellationContext(ctx context.Context, tx *gorm.DB, cancelCtx *models.SubscriptionCancellationContext) error
	InsertCancellationHistory(ctx context.Context, tx *gorm.DB, history *models.SubscriptionCancellationContextHistory) error

	InsertOfferDetails(ctx context.Context, tx *gorm.DB, detail *models.OfferDetails) error
	InsertOfferDetailsHistory(ctx context.Context, tx *gorm.DB, detail *models.OfferDetailsHistory) error
	UpdateOfferDetails(ctx context.Context, tx *gorm.DB, offer *models.OfferDetails) error

	InsertAutoRenewingPlan(ctx context.Context, tx *gorm.DB, plan *models.AutoRenewingPlan) error
	DeleteAutoRenewingPlansByLineItem(ctx context.Context, tx *gorm.DB, lineItemID uuid.UUID) error
	InsertAutoRenewingPlanHistory(ctx context.Context, tx *gorm.DB, history *models.AutoRenewingPlanHistory) error
	InsertPrepaidPlan(ctx context.Context, tx *gorm.DB, plan *models.PrepaidPlan) error
	UpdatePrepaidPlan(ctx context.Context, tx *gorm.DB, plan *models.PrepaidPlan) error
	InsertPrepaidPlanHistory(ctx context.Context, tx *gorm.DB, history *models.PrepaidPlanHistory) error

	CreateSignupPromotion(ctx context.Context, tx *gorm.DB, promotion *models.SignupPromotion) error
	UpdateSignupPromotion(ctx context.Context, tx *gorm.DB, promotion *models.SignupPromotion) error
	CreateSignupPromotionHistory(ctx context.Context, tx *gorm.DB, promotion *models.SignupPromotionHistory) error

	CreateDeferredReplacement(ctx context.Context, tx *gorm.DB, replacement *models.DeferredItemReplacement) error
	UpdateDeferredReplacement(ctx context.Context, tx *gorm.DB, replacement *models.DeferredItemReplacement) error
	CreateDeferredReplacementHistory(ctx context.Context, tx *gorm.DB, replacement *models.DeferredItemReplacementHistory) error

	CreateBulkSubscriptionLineItemHistory(ctx context.Context, tx *gorm.DB, entries []models.SubscriptionLineItemHistory) error
	CreateSubscriptionEvent(ctx context.Context, tx *gorm.DB, event *models.SubscriptionEvent) error
}

type playstoreSubscriptionRepository struct{}

func NewPlaystoreSubscriptionRepository() PlaystoreSubscriptionRepository {
	return &playstoreSubscriptionRepository{}
}

func (r *playstoreSubscriptionRepository) GetSubscriptionByPurchaseToken(
	ctx context.Context,
	tx *gorm.DB,
	token string,
) (*models.SubscriptionPurchaseV2, error) {
	if token == "" {
		return nil, fmt.Errorf("purchase token is empty")
	}

	var sub models.SubscriptionPurchaseV2
	err := tx.WithContext(ctx).
		Preload("User").
		Preload("User.GoogleAccount"). // Load user + nested GoogleAccount
		Preload("SubscriptionCancellationContext").
		Preload("LineItems").
		Preload("LineItems.AutoRenewingPlan").
		Preload("LineItems.OfferDetails").
		Where("purchase_token = ?", token).
		First(&sub).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err // caller will check and handle
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription by token: %w", err)
	}

	return &sub, nil
}

func (r *playstoreSubscriptionRepository) InsertSubscription(
	ctx context.Context,
	tx *gorm.DB,
	sub *models.SubscriptionPurchaseV2,
) error {
	if sub == nil {
		return fmt.Errorf("subscription object is nil")
	}

	return tx.WithContext(ctx).Create(sub).Error
}

func (r *playstoreSubscriptionRepository) UpdateSubscriptionFields(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	updateFields map[string]interface{},
) error {

	if len(updateFields) == 0 {
		return nil // Nothing to update
	}

	err := tx.WithContext(ctx).
		Model(&models.SubscriptionPurchaseV2{}).
		Where("id = ?", subscriptionID).
		Updates(updateFields).Error

	if err != nil {
		return fmt.Errorf("failed to update subscription fields: %w", err)
	}

	return nil
}

func (r *playstoreSubscriptionRepository) InsertSubscriptionStateTransition(
	ctx context.Context,
	tx *gorm.DB,
	history *models.SubscriptionStateTransitionHistory,
) error {
	return tx.WithContext(ctx).Create(history).Error
}

func (r *playstoreSubscriptionRepository) InsertOrderIDTransition(
	ctx context.Context,
	tx *gorm.DB,
	history *models.SubscriptionOrderIdTransitionHistory,
) error {
	if history == nil {
		return errors.New("nil order ID transition history provided")
	}

	if err := tx.WithContext(ctx).Create(history).Error; err != nil {
		return fmt.Errorf("failed to insert order ID transition history: %w", err)
	}

	return nil
}

func (r *playstoreSubscriptionRepository) InsertAcknowledgementStateTransition(
	ctx context.Context,
	tx *gorm.DB,
	history *models.AcknowledgementStateTransitionHistory,
) error {
	if err := tx.WithContext(ctx).Create(history).Error; err != nil {
		return fmt.Errorf("failed to insert acknowledgement state transition: %w", err)
	}
	return nil
}

func (r *playstoreSubscriptionRepository) UpdatePausedContext(
	ctx context.Context,
	tx *gorm.DB,
	context *models.SubscriptionPausedContext,
) error {
	return tx.WithContext(ctx).Save(context).Error
}

func (r *playstoreSubscriptionRepository) InsertPausedContext(
	ctx context.Context,
	tx *gorm.DB,
	context *models.SubscriptionPausedContext,
) error {
	return tx.WithContext(ctx).Create(context).Error
}

func (r *playstoreSubscriptionRepository) InsertPausedContextHistory(
	ctx context.Context,
	tx *gorm.DB,
	history *models.SubscriptionPausedContextHistory,
) error {
	return tx.WithContext(ctx).Create(history).Error
}

func (r *playstoreSubscriptionRepository) InsertCancellationContext(
	ctx context.Context,
	tx *gorm.DB,
	cancelCtx *models.SubscriptionCancellationContext,
) error {
	return tx.WithContext(ctx).Create(cancelCtx).Error
}

func (r *playstoreSubscriptionRepository) UpdateCancellationContext(
	ctx context.Context,
	tx *gorm.DB,
	cancelCtx *models.SubscriptionCancellationContext,
) error {
	return tx.WithContext(ctx).Save(cancelCtx).Error
}

func (r *playstoreSubscriptionRepository) InsertCancellationHistory(
	ctx context.Context,
	tx *gorm.DB,
	history *models.SubscriptionCancellationContextHistory,
) error {
	return tx.WithContext(ctx).Create(history).Error
}

func (r *playstoreSubscriptionRepository) InsertOfferDetails(
	ctx context.Context, tx *gorm.DB, detail *models.OfferDetails,
) error {
	return tx.WithContext(ctx).Create(detail).Error
}

func (r *playstoreSubscriptionRepository) InsertOfferDetailsHistory(
	ctx context.Context, tx *gorm.DB, history *models.OfferDetailsHistory,
) error {
	return tx.WithContext(ctx).Create(history).Error
}

func (r *playstoreSubscriptionRepository) UpdateOfferDetails(
	ctx context.Context,
	tx *gorm.DB,
	offer *models.OfferDetails,
) error {
	return tx.WithContext(ctx).Model(&models.OfferDetails{}).
		Where("id = ?", offer.ID).
		Updates(map[string]interface{}{
			"base_plan_id": offer.BasePlanID,
			"offer_id":     offer.OfferID,
			"offer_tags":   offer.OfferTags,
		}).Error
}

func (r *playstoreSubscriptionRepository) InsertAutoRenewingPlan(ctx context.Context, tx *gorm.DB, plan *models.AutoRenewingPlan) error {
	return tx.WithContext(ctx).Create(plan).Error
}

// Implementation in repository
func (r *playstoreSubscriptionRepository) DeleteAutoRenewingPlansByLineItem(ctx context.Context, tx *gorm.DB, lineItemID uuid.UUID) error {
	return tx.WithContext(ctx).
		Where("line_item_id = ?", lineItemID).
		Delete(&models.AutoRenewingPlan{}).
		Error
}

func (r *playstoreSubscriptionRepository) InsertAutoRenewingPlanHistory(ctx context.Context, tx *gorm.DB, history *models.AutoRenewingPlanHistory) error {
	return tx.WithContext(ctx).Create(history).Error
}

// InsertPrepaidPlan creates a new prepaid plan entry
func (r *playstoreSubscriptionRepository) InsertPrepaidPlan(ctx context.Context, tx *gorm.DB, plan *models.PrepaidPlan) error {
	return tx.WithContext(ctx).Create(plan).Error
}

// UpdatePrepaidPlan updates the current prepaid plan (overwrite strategy)
func (r *playstoreSubscriptionRepository) UpdatePrepaidPlan(ctx context.Context, tx *gorm.DB, plan *models.PrepaidPlan) error {
	return tx.WithContext(ctx).Save(plan).Error
}

// InsertPrepaidPlanHistory appends a row to the prepaid plan history
func (r *playstoreSubscriptionRepository) InsertPrepaidPlanHistory(ctx context.Context, tx *gorm.DB, history *models.PrepaidPlanHistory) error {
	return tx.WithContext(ctx).Create(history).Error
}

func (r *playstoreSubscriptionRepository) CreateSignupPromotion(ctx context.Context, tx *gorm.DB, promotion *models.SignupPromotion) error {
	return tx.WithContext(ctx).Create(promotion).Error
}

func (r *playstoreSubscriptionRepository) UpdateSignupPromotion(ctx context.Context, tx *gorm.DB, promotion *models.SignupPromotion) error {
	return tx.WithContext(ctx).Save(promotion).Error
}
func (r *playstoreSubscriptionRepository) CreateSignupPromotionHistory(ctx context.Context, tx *gorm.DB, history *models.SignupPromotionHistory) error {
	return tx.WithContext(ctx).Create(history).Error
}

func (r *playstoreSubscriptionRepository) CreateDeferredReplacement(ctx context.Context, tx *gorm.DB, replacement *models.DeferredItemReplacement) error {
	return tx.WithContext(ctx).Create(replacement).Error
}

func (r *playstoreSubscriptionRepository) UpdateDeferredReplacement(ctx context.Context, tx *gorm.DB, replacement *models.DeferredItemReplacement) error {
	return tx.WithContext(ctx).Save(replacement).Error
}

func (r *playstoreSubscriptionRepository) CreateDeferredReplacementHistory(ctx context.Context, tx *gorm.DB, history *models.DeferredItemReplacementHistory) error {
	return tx.WithContext(ctx).Create(history).Error
}

func (r *playstoreSubscriptionRepository) CreateBulkSubscriptionLineItemHistory(ctx context.Context, tx *gorm.DB, entries []models.SubscriptionLineItemHistory) error {
	if len(entries) == 0 {
		return nil
	}
	return tx.WithContext(ctx).Create(&entries).Error
}

func (r *playstoreSubscriptionRepository) CreateSubscriptionEvent(ctx context.Context, tx *gorm.DB, event *models.SubscriptionEvent) error {
	if event == nil {
		return errors.New("nil event provided")
	}

	// Validate required fields
	if event.SubscriptionID == uuid.Nil {
		return errors.New("subscription ID cannot be empty")
	}
	if event.EventID == uuid.Nil {
		return errors.New("event ID cannot be empty")
	}
	if event.EventType == "" {
		return errors.New("event type cannot be empty")
	}

	// Create the event with context
	result := tx.WithContext(ctx).Create(event)
	if result.Error != nil {
		return fmt.Errorf("database error: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("no rows affected - event not created")
	}

	return nil
}
