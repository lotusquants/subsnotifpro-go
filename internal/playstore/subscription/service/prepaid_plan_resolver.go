package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"google.golang.org/api/androidpublisher/v3"
	"gorm.io/gorm"
)

// PrepaidPlan specific methods
func (s *playstoreSubscriptionService) createPrepaidPlan(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	newLineItemModel *models.SubscriptionLineItem,
	newLineItemData *androidpublisher.SubscriptionPurchaseLineItem,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {
	if newLineItemData.PrepaidPlan == nil {
		return nil, nil
	}

	// Parse time values
	allowExtendAfterTime := parseTimeOrNil(newLineItemData.PrepaidPlan.AllowExtendAfterTime)
	expiryTime := parseTimeOrNil(newLineItemData.ExpiryTime)

	prepaidPlan := &models.PrepaidPlan{
		SubscriptionID:       subscriptionID,
		LineItemID:           newLineItemModel.ID,
		ProductID:            newLineItemData.ProductId,
		AllowExtendAfterTime: &allowExtendAfterTime,
		ExpiryTime:           expiryTime,
	}

	// Create the prepaid plan
	if err := tx.Create(prepaidPlan).Error; err != nil {
		return nil, fmt.Errorf("failed to create prepaid plan: %w", err)
	}

	// Create history entry
	history := models.PrepaidPlanHistory{
		SubscriptionID:              subscriptionID,
		LineItemID:                  newLineItemModel.ID,
		ChangeType:                  models.PrepaidPlanChangeCreated,
		ProductID:                   newLineItemData.ProductId,
		CurrentAllowExtendAfterTime: &allowExtendAfterTime,
		CurrentExpiryTime:           &expiryTime,
		ChangeEventID:               changeEventID,
	}

	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create prepaid plan history: %w", err)
	}

	return &prepaidPlan.ID, nil
}

func (s *playstoreSubscriptionService) updatePrepaidPlan(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	existingLineItemModel *models.SubscriptionLineItem,
	newLineItemData *androidpublisher.SubscriptionPurchaseLineItem,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {
	if newLineItemData.PrepaidPlan == nil {
		return nil, nil
	}

	if existingLineItemModel.PrepaidPlan == nil {
		return nil, errors.New("no existing prepaid renewing plan found")
	}

	existingPlan := existingLineItemModel.PrepaidPlan

	// Parse new time values
	var newAllowExtendAfterTime *time.Time
	if newLineItemData.PrepaidPlan.AllowExtendAfterTime != "" {
		parsedTime := parseTimeOrNil(newLineItemData.PrepaidPlan.AllowExtendAfterTime)
		newAllowExtendAfterTime = &parsedTime
	}
	newExpiryTime := parseTimeOrNil(newLineItemData.ExpiryTime)

	// Prepare history record
	history := models.PrepaidPlanHistory{
		SubscriptionID:               subscriptionID,
		LineItemID:                   existingLineItemModel.ID,
		ChangeType:                   models.PrepaidPlanChangeExtended,
		ProductID:                    existingPlan.ProductID,
		PreviousAllowExtendAfterTime: existingPlan.AllowExtendAfterTime,
		PreviousExpiryTime:           &existingPlan.ExpiryTime,
		ChangeEventID:                changeEventID,
	}

	updates := make(map[string]interface{})

	// Check and update AllowExtendAfterTime
	if newAllowExtendAfterTime != nil {
		if existingPlan.AllowExtendAfterTime == nil ||
			!existingPlan.AllowExtendAfterTime.Equal(*newAllowExtendAfterTime) {
			updates["allow_extend_after_time"] = newAllowExtendAfterTime
			history.CurrentAllowExtendAfterTime = newAllowExtendAfterTime
		}
	} else if existingPlan.AllowExtendAfterTime != nil {
		// Clear the allow extend time if it was set but now empty
		updates["allow_extend_after_time"] = nil
		history.CurrentAllowExtendAfterTime = nil
	}

	// Check and update ExpiryTime
	if !existingPlan.ExpiryTime.Equal(newExpiryTime) {
		updates["expiry_time"] = newExpiryTime
		history.CurrentExpiryTime = &newExpiryTime
	}

	// Only proceed if there are updates
	if len(updates) == 0 {
		return &existingPlan.ID, nil
	}

	// Update the plan
	if err := tx.Model(&existingPlan).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update prepaid plan: %w", err)
	}

	// Create history entry
	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create prepaid plan history: %w", err)
	}

	return &existingPlan.ID, nil
}

func (s *playstoreSubscriptionService) expirePrepaidPlan(
	ctx context.Context,
	tx *gorm.DB,
	expiredLineItem models.SubscriptionLineItem,
	changeEventID uuid.UUID,
) error {

	// Verify we have the auto renewing plan
	if expiredLineItem.PrepaidPlan == nil {
		return nil // Nothing to expire
	}

	plan := expiredLineItem.PrepaidPlan

	// Create history before soft deleting
	history := models.PrepaidPlanHistory{
		SubscriptionID:               plan.SubscriptionID,
		LineItemID:                   plan.LineItemID,
		ChangeType:                   "EXPIRED",
		ProductID:                    plan.ProductID,
		PreviousAllowExtendAfterTime: plan.AllowExtendAfterTime,
		PreviousExpiryTime:           &plan.ExpiryTime,

		ChangeEventID: changeEventID,
	}

	if err := tx.Create(&history).Error; err != nil {
		return fmt.Errorf("failed to create prepaid plan history: %w", err)
	}

	// Soft delete the plan
	if err := tx.Delete(&plan).Error; err != nil {
		return fmt.Errorf("failed to soft delete prepaid plan: %w", err)
	}

	return nil
}
