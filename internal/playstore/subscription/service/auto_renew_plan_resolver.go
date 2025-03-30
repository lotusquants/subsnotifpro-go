package service

import (
	"context"
	"errors"
	"fmt"
	"subsnotifpro-go/internal/playstore/subscription/models"
	"time"

	"github.com/google/uuid"
	"google.golang.org/api/androidpublisher/v3"
	"gorm.io/gorm"
)

func (s *playstoreSubscriptionService) createAutoRenewingPlan(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	newLineItemModel *models.SubscriptionLineItem,
	newLineItemData *androidpublisher.SubscriptionPurchaseLineItem,
	recurringPrice models.Money,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {
	if newLineItemData.AutoRenewingPlan == nil {
		return nil, nil
	}

	expiryTime, err := time.Parse(time.RFC3339Nano, newLineItemData.ExpiryTime)
	if err != nil {
		return nil, fmt.Errorf("invalid expiry time format: %w", err)
	}

	autoPlan := &models.AutoRenewingPlan{
		SubscriptionID:   subscriptionID,
		LineItemID:       newLineItemModel.ID,
		ProductID:        newLineItemData.ProductId,
		ExpiryTime:       expiryTime,
		AutoRenewEnabled: newLineItemData.AutoRenewingPlan.AutoRenewEnabled,
		RecurringPrice:   recurringPrice,
	}

	// Handle price change details if present
	if newLineItemData.AutoRenewingPlan.PriceChangeDetails != nil {
		priceChangeDetails, err := s.createPriceChangeDetails(
			ctx, tx, newLineItemModel, newLineItemData, changeEventID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create price change details: %w", err)
		}
		autoPlan.PriceChangeDetailsID = &priceChangeDetails.ID
	}

	// Handle installment plan if present
	if newLineItemData.AutoRenewingPlan.InstallmentDetails != nil {
		installmentPlan, err := s.createInstallmentPlan(
			ctx, tx, subscriptionID, newLineItemModel, newLineItemData, changeEventID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create installment plan: %w", err)
		}
		autoPlan.InstallmentPlanID = &installmentPlan.ID
	}

	if err := tx.Create(autoPlan).Error; err != nil {
		return nil, fmt.Errorf("failed to create auto renewing plan: %w", err)
	}

	// Create history entry
	history := models.AutoRenewingPlanHistory{
		SubscriptionID:          subscriptionID,
		LineItemID:              newLineItemModel.ID,
		AutoRenewingPlanID:      autoPlan.ID,
		ChangeType:              models.AutoRenewingPlanChangeCreated,
		ProductID:               newLineItemData.ProductId,
		CurrentExpiryTime:       &expiryTime,
		CurrentAutoRenewEnabled: &newLineItemData.AutoRenewingPlan.AutoRenewEnabled,
		PriceChangeDetailsID:    autoPlan.PriceChangeDetailsID,
		InstallmentPlanID:       autoPlan.InstallmentPlanID,
		CurrentPrice:            &recurringPrice,
		ChangeEventID:           changeEventID,
	}

	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create auto renewing plan history: %w", err)
	}

	return &autoPlan.ID, nil
}

func (s *playstoreSubscriptionService) updateAutoRenewingPlan(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	existingLineItemModel *models.SubscriptionLineItem,
	newLineItemData *androidpublisher.SubscriptionPurchaseLineItem,
	recurringPrice models.Money,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {
	if newLineItemData.AutoRenewingPlan == nil {
		return nil, errors.New("no auto renewing plan data provided")
	}

	// Use the preloaded plan from the line item
	if existingLineItemModel.AutoRenewingPlan == nil {
		return nil, errors.New("no existing auto renewing plan found")
	}
	existingPlan := existingLineItemModel.AutoRenewingPlan

	// Parse new expiry time
	newExpiryTime, err := time.Parse(time.RFC3339Nano, newLineItemData.ExpiryTime)
	if err != nil {
		return nil, fmt.Errorf("invalid expiry time format: %w", err)
	}

	// Initialize history record with previous values
	history := models.AutoRenewingPlanHistory{
		SubscriptionID:           subscriptionID,
		LineItemID:               existingLineItemModel.ID,
		AutoRenewingPlanID:       existingLineItemModel.AutoRenewingPlan.ID,
		ProductID:                existingPlan.ProductID,
		CurrentExpiryTime:        &newExpiryTime,
		CurrentAutoRenewEnabled:  &newLineItemData.AutoRenewingPlan.AutoRenewEnabled,
		CurrentPrice:             &recurringPrice,
		PreviousExpiryTime:       &existingPlan.ExpiryTime,
		PreviousAutoRenewEnabled: &existingPlan.AutoRenewEnabled,
		PreviousPrice:            &existingPlan.RecurringPrice,
		PriceChangeDetailsID:     existingPlan.PriceChangeDetailsID,
		InstallmentPlanID:        existingPlan.InstallmentPlanID,
		ChangeEventID:            changeEventID,
	}

	changes := make(map[string]interface{})
	changeType := models.AutoRenewingPlanChangeRenewed

	// Check and track changes for each field
	if !existingPlan.ExpiryTime.Equal(newExpiryTime) {
		changes["expiry_time"] = newExpiryTime
		history.CurrentExpiryTime = &newExpiryTime
		changeType = models.AutoRenewingPlanChangeRenewed
	}

	if existingPlan.AutoRenewEnabled != newLineItemData.AutoRenewingPlan.AutoRenewEnabled {
		changes["auto_renew_enabled"] = newLineItemData.AutoRenewingPlan.AutoRenewEnabled
		history.CurrentAutoRenewEnabled = &newLineItemData.AutoRenewingPlan.AutoRenewEnabled
	}

	if existingPlan.RecurringPrice != recurringPrice {
		changes["recurring_price"] = recurringPrice
		history.CurrentPrice = &recurringPrice
		changeType = models.AutoRenewingPlanChangePriceChange
	}

	// Handle price change details
	if newLineItemData.AutoRenewingPlan.PriceChangeDetails != nil {
		var priceChangeDetails *models.SubscriptionItemPriceChangeDetails
		var err error

		if existingPlan.PriceChangeDetailsID != nil {
			priceChangeDetails, err = s.updatePriceChangeDetails(
				ctx, tx, existingLineItemModel, newLineItemData, changeEventID,
			)
		} else {
			priceChangeDetails, err = s.createPriceChangeDetails(
				ctx, tx, existingLineItemModel, newLineItemData, changeEventID,
			)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to handle price change details: %w", err)
		}

		if priceChangeDetails != nil && (existingPlan.PriceChangeDetailsID == nil ||
			*existingPlan.PriceChangeDetailsID != priceChangeDetails.ID) {
			changes["price_change_details_id"] = &priceChangeDetails.ID
		}
	}

	// Handle installment plan updates
	if newLineItemData.AutoRenewingPlan.InstallmentDetails != nil {
		var installmentPlan *models.InstallmentPlan
		var err error

		if existingPlan.InstallmentPlanID != nil {
			installmentPlan, err = s.updateInstallmentPlan(
				ctx, tx,
				subscriptionID, existingLineItemModel, newLineItemData, changeEventID,
			)
		} else {
			installmentPlan, err = s.createInstallmentPlan(
				ctx, tx, subscriptionID, existingLineItemModel, newLineItemData, changeEventID,
			)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to handle installment plan: %w", err)
		}

		if installmentPlan != nil && (existingPlan.InstallmentPlanID == nil ||
			*existingPlan.InstallmentPlanID != installmentPlan.ID) {
			changes["installment_plan_id"] = &installmentPlan.ID
		}
	}

	// Only proceed if there are changes
	if len(changes) == 0 {
		return &existingPlan.ID, nil
	}

	// Update the plan
	if err := tx.Model(&existingPlan).Updates(changes).Error; err != nil {
		return nil, fmt.Errorf("failed to update auto renewing plan: %w", err)
	}

	// Set the final change type and create history
	history.ChangeType = changeType
	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create auto renewing plan history: %w", err)
	}

	return &existingPlan.ID, nil
}

func (s *playstoreSubscriptionService) expireAutoRenewingPlan(
	ctx context.Context,
	tx *gorm.DB,
	expiredLineItem models.SubscriptionLineItem,
	changeEventID uuid.UUID,
) error {

	// Verify we have the auto renewing plan
	if expiredLineItem.AutoRenewingPlan == nil {
		return nil // Nothing to expire
	}

	plan := expiredLineItem.AutoRenewingPlan

	// Create comprehensive history before soft deleting
	history := models.AutoRenewingPlanHistory{
		SubscriptionID:           plan.SubscriptionID,
		LineItemID:               plan.LineItemID,
		AutoRenewingPlanID:       plan.ID,
		ChangeType:               models.AutoRenewingPlanChangeExpired,
		ProductID:                plan.ProductID,
		PreviousExpiryTime:       &plan.ExpiryTime,
		PreviousAutoRenewEnabled: &plan.AutoRenewEnabled,
		PreviousPrice:            &plan.RecurringPrice,
		PriceChangeDetailsID:     plan.PriceChangeDetailsID,
		InstallmentPlanID:        plan.InstallmentPlanID,
		ChangeEventID:            changeEventID,
	}

	if err := tx.Create(&history).Error; err != nil {
		return fmt.Errorf("failed to create auto renewing plan history: %w", err)
	}

	// First expire any nested models
	if plan.PriceChangeDetailsID != nil {
		if err := s.expirePriceChangeDetails(ctx, tx, expiredLineItem, changeEventID); err != nil {
			return fmt.Errorf("failed to expire price change details: %w", err)
		}
	}

	if plan.InstallmentPlanID != nil {
		if err := s.expireInstallmentPlan(ctx, tx, expiredLineItem, changeEventID); err != nil {
			return fmt.Errorf("failed to expire installment plan: %w", err)
		}
	}

	// Finally soft delete the plan
	if err := tx.Delete(&plan).Error; err != nil {
		return fmt.Errorf("failed to soft delete auto renewing plan: %w", err)
	}

	return nil
}
