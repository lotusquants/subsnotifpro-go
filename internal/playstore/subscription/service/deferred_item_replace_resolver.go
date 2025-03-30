package service

import (
	"context"
	"errors"
	"fmt"
	"subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"google.golang.org/api/androidpublisher/v3"
	"gorm.io/gorm"
)

// DeferredItemReplacement methods
func (s *playstoreSubscriptionService) createDeferredItemReplacement(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	newlineItemID uuid.UUID,
	newLineItemData *androidpublisher.SubscriptionPurchaseLineItem,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {
	if newLineItemData.DeferredItemReplacement == nil {
		return nil, nil
	}

	deferredItem := &models.DeferredItemReplacement{
		SubscriptionID:    subscriptionID,
		LineItemID:        newlineItemID,
		PreviousProductID: newLineItemData.ProductId,
		NewProductID:      newLineItemData.DeferredItemReplacement.ProductId,
	}

	if err := tx.Create(deferredItem).Error; err != nil {
		return nil, fmt.Errorf("failed to create deferred item replacement: %w", err)
	}

	history := models.DeferredItemReplacementHistory{
		SubscriptionID:    subscriptionID,
		LineItemID:        newlineItemID,
		ChangeType:        "CREATED",
		PreviousProductID: &deferredItem.PreviousProductID,
		NewProductID:      newLineItemData.DeferredItemReplacement.ProductId,
		ChangeEventID:     changeEventID,
	}

	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create deferred item replacement history: %w", err)
	}

	return &deferredItem.ID, nil
}

func (s *playstoreSubscriptionService) updateDeferredItemReplacement(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	existingLineItemID uuid.UUID,
	newLineItemData *androidpublisher.SubscriptionPurchaseLineItem,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {
	if newLineItemData.DeferredItemReplacement == nil {
		// If no replacement in new data, consider removing existing replacement
		return nil, nil
	}

	// Get existing replacement record
	var existing models.DeferredItemReplacement
	if err := tx.Where("line_item_id = ?", existingLineItemID).
		First(&existing).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No existing record found, treat as create
			return s.createDeferredItemReplacement(ctx, tx, subscriptionID, existingLineItemID, newLineItemData, changeEventID)
		}
		return nil, fmt.Errorf("failed to find existing deferred item replacement: %w", err)
	}

	// Check if replacement actually changed
	if existing.NewProductID == newLineItemData.DeferredItemReplacement.ProductId &&
		existing.PreviousProductID == newLineItemData.ProductId {
		// No changes needed
		return &existing.ID, nil
	}

	// Prepare updates
	updates := map[string]interface{}{
		"previous_product_id": newLineItemData.ProductId,
		"new_product_id":      newLineItemData.DeferredItemReplacement.ProductId,
	}

	// Create history before updating
	history := models.DeferredItemReplacementHistory{
		SubscriptionID:    subscriptionID,
		LineItemID:        existingLineItemID,
		PreviousProductID: &existing.NewProductID,
		NewProductID:      newLineItemData.DeferredItemReplacement.ProductId,
		ChangeType:        "UPDATED",
		ChangeEventID:     changeEventID,
	}

	// Update the record
	if err := tx.Model(&existing).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update deferred item replacement: %w", err)
	}

	// Create history entry
	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create deferred item replacement history: %w", err)
	}

	return &existing.ID, nil
}

// DeferredItemReplacement expiration
func (s *playstoreSubscriptionService) expireDeferredItemReplacement(
	ctx context.Context,
	tx *gorm.DB,
	expiredLineItem models.SubscriptionLineItem,
	changeEventID uuid.UUID,
) error {
	// Verify we have a deferred item replacement to expire
	if expiredLineItem.DeferredItemReplacement == nil {
		return nil // Nothing to expire
	}

	replacement := expiredLineItem.DeferredItemReplacement

	// Create comprehensive history before soft deleting
	history := models.DeferredItemReplacementHistory{
		SubscriptionID:    replacement.SubscriptionID,
		LineItemID:        replacement.LineItemID,
		PreviousProductID: &replacement.PreviousProductID,
		NewProductID:      replacement.NewProductID,
		ChangeType:        "REMOVED",
		ChangeEventID:     changeEventID,
	}

	if err := tx.Create(&history).Error; err != nil {
		return fmt.Errorf("failed to create deferred item replacement history: %w", err)
	}

	// Soft delete the replacement record
	if err := tx.Delete(replacement).Error; err != nil {
		return fmt.Errorf("failed to soft delete deferred item replacement: %w", err)
	}

	return nil
}
