package service

import (
	"context"
	"errors"
	"fmt"
	"subsnotifpro-go/internal/playstore/api/dto"
	"subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *playstoreSubscriptionService) createPriceChangeDetails(
	ctx context.Context,
	tx *gorm.DB,
	lineItem *models.SubscriptionLineItem,
	planData *dto.LineItem,
	changeEventID uuid.UUID,
) (*models.SubscriptionItemPriceChangeDetails, error) {
	if planData.AutoRenewingPlan == nil || planData.AutoRenewingPlan.PriceChangeDetails == nil {
		return nil, nil
	}

	details := planData.AutoRenewingPlan.PriceChangeDetails
	expectedChangeTime := details.ExpectedNewPriceChargeTime

	// Convert API enums to our model enums
	priceChangeMode := models.PriceChangeMode(details.PriceChangeMode)
	priceChangeState := models.PriceChangeState(details.PriceChangeState)

	newPrice := models.Money{
		CurrencyCode: details.NewPrice.CurrencyCode,
		Units:        details.NewPrice.Units,
		Nanos:        details.NewPrice.Nanos,
	}

	priceChange := &models.SubscriptionItemPriceChangeDetails{
		SubscriptionID:             lineItem.SubscriptionID,
		LineItemID:                 lineItem.ID,
		AutoRenewingPlanID:         lineItem.AutoRenewingPlan.ID,
		NewPrice:                   newPrice,
		PriceChangeMode:            priceChangeMode,
		PriceChangeState:           priceChangeState,
		ExpectedNewPriceChangeTime: &expectedChangeTime,
	}

	if err := tx.Create(priceChange).Error; err != nil {
		return nil, fmt.Errorf("failed to create price change details: %w", err)
	}

	// Create history entry
	history := models.SubscriptionItemPriceChangeDetailsHistory{
		PriceChangeDetailsID:  priceChange.ID,
		SubscriptionID:        lineItem.SubscriptionID,
		LineItemID:            lineItem.ID,
		AutoRenewingPlanID:    lineItem.AutoRenewingPlan.ID,
		NewPrice:              newPrice,
		NewPriceChangeMode:    priceChangeMode,
		NewPriceChangeState:   priceChangeState,
		NewExpectedChangeTime: &expectedChangeTime,
		ChangeType:            "ADDED",
		ChangeEventID:         changeEventID,
	}

	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create price change history: %w", err)
	}

	return priceChange, nil
}

func (s *playstoreSubscriptionService) updatePriceChangeDetails(
	ctx context.Context,
	tx *gorm.DB,
	existingLineItem *models.SubscriptionLineItem,
	planData *dto.LineItem,
	changeEventID uuid.UUID,
) (*models.SubscriptionItemPriceChangeDetails, error) {
	if planData.AutoRenewingPlan == nil || planData.AutoRenewingPlan.PriceChangeDetails == nil {
		return nil, nil
	}

	if existingLineItem.AutoRenewingPlan == nil || existingLineItem.AutoRenewingPlan.PriceChangeDetails == nil {
		return nil, nil
	}

	existing := existingLineItem.AutoRenewingPlan.PriceChangeDetails

	details := planData.AutoRenewingPlan.PriceChangeDetails
	expectedChangeTime := details.ExpectedNewPriceChargeTime
	newPrice := models.Money{
		CurrencyCode: details.NewPrice.CurrencyCode,
		Units:        details.NewPrice.Units,
		Nanos:        details.NewPrice.Nanos,
	}

	// Convert API enums to our model enums
	newPriceChangeMode := models.PriceChangeMode(details.PriceChangeMode)
	newPriceChangeState := models.PriceChangeState(details.PriceChangeState)

	// Update the existing record
	updates := map[string]interface{}{
		"new_price":                      newPrice,
		"price_change_mode":              newPriceChangeMode,
		"price_change_state":             newPriceChangeState,
		"expected_new_price_change_time": expectedChangeTime,
		"is_current":                     true,
	}

	if err := tx.Model(&existing).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update price change details: %w", err)
	}

	// Create history entry
	history := models.SubscriptionItemPriceChangeDetailsHistory{
		PriceChangeDetailsID:       existing.ID,
		SubscriptionID:             existingLineItem.SubscriptionID,
		LineItemID:                 existingLineItem.ID,
		AutoRenewingPlanID:         existingLineItem.AutoRenewingPlan.ID,
		PreviousPrice:              &existing.NewPrice,
		NewPrice:                   newPrice,
		PreviousPriceChangeMode:    &existing.PriceChangeMode,
		NewPriceChangeMode:         newPriceChangeMode,
		PreviousPriceChangeState:   &existing.PriceChangeState,
		NewPriceChangeState:        newPriceChangeState,
		PreviousExpectedChangeTime: existing.ExpectedNewPriceChangeTime,
		NewExpectedChangeTime:      &expectedChangeTime,
		ChangeType:                 "UPDATED",
		ChangeEventID:              changeEventID,
	}

	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create price change history: %w", err)
	}

	return existing, nil
}

func (s *playstoreSubscriptionService) expirePriceChangeDetails(
	ctx context.Context,
	tx *gorm.DB,
	expiredLineItem models.SubscriptionLineItem,
	changeEventID uuid.UUID,
) error {
	// Verify we have price change details to expire
	if expiredLineItem.AutoRenewingPlan == nil || expiredLineItem.AutoRenewingPlan.PriceChangeDetailsID == nil {
		return nil // Nothing to expire
	}

	// Get the full price change details record
	var details models.SubscriptionItemPriceChangeDetails
	if err := tx.First(&details, "id = ?", *expiredLineItem.AutoRenewingPlan.PriceChangeDetailsID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // Already expired or doesn't exist
		}
		return fmt.Errorf("failed to find price change details: %w", err)
	}

	// Create history before soft deleting
	history := models.SubscriptionItemPriceChangeDetailsHistory{
		PriceChangeDetailsID:       details.ID,
		SubscriptionID:             details.SubscriptionID,
		LineItemID:                 details.LineItemID,
		AutoRenewingPlanID:         details.AutoRenewingPlanID,
		PreviousPrice:              &details.NewPrice,
		PreviousPriceChangeMode:    &details.PriceChangeMode,
		PreviousPriceChangeState:   &details.PriceChangeState,
		PreviousExpectedChangeTime: details.ExpectedNewPriceChangeTime,
		ChangeType:                 "REMOVED",
		ChangeEventID:              changeEventID,
	}

	if err := tx.Create(&history).Error; err != nil {
		return fmt.Errorf("failed to create price change details history: %w", err)
	}

	// Soft delete the price change details
	if err := tx.Delete(&details).Error; err != nil {
		return fmt.Errorf("failed to soft delete price change details: %w", err)
	}

	return nil
}
