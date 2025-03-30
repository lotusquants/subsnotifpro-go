package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	rtdnModels "subsnotifpro-go/internal/playstore/rtdn/models"
	"subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"google.golang.org/api/androidpublisher/v3"
	"gorm.io/gorm"
)

func (s *playstoreSubscriptionService) ResolveLineItems(
	ctx context.Context,
	tx *gorm.DB,
	existingSubID *uuid.UUID,
	existing *models.SubscriptionPurchaseV2,
	subData *androidpublisher.SubscriptionPurchaseV2,
	changeEventID uuid.UUID,
	notificationType rtdnModels.SubscriptionNotificationType,
) ([]models.SubscriptionLineItem, error) {
	var resolvedItems []models.SubscriptionLineItem
	var lineItemHistories []models.SubscriptionLineItemHistory

	// Create maps for efficient lookup
	existingItemsByProduct := make(map[string]*models.SubscriptionLineItem)
	for i := range existing.LineItems {
		item := existing.LineItems[i]
		existingItemsByProduct[item.ProductID] = &item
	}

	// Process each line item from Play Store data
	for _, newItem := range subData.LineItems {
		productID := newItem.ProductId
		existingItem, exists := existingItemsByProduct[productID]

		if exists {
			// Update existing line item and its nested models
			updatedItem, itemHistory, err := s.updateLineItemAndNestedModels(
				ctx, tx, *existingSubID, *existing, existingItem, newItem, changeEventID, notificationType,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to update line item %s: %w", productID, err)
			}
			resolvedItems = append(resolvedItems, *updatedItem)
			if itemHistory != nil {
				lineItemHistories = append(lineItemHistories, *itemHistory)
			}
			delete(existingItemsByProduct, productID)
		} else {
			// Create new line item with nested models
			newLineItem, itemHistory, err := s.createLineItemWithNestedModels(
				ctx, tx, *existingSubID, *existing, newItem, changeEventID,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create line item %s: %w", productID, err)
			}
			resolvedItems = append(resolvedItems, *newLineItem)
			lineItemHistories = append(lineItemHistories, *itemHistory)
		}
	}

	// Handle expired line items and their nested models
	for _, expiredItem := range existingItemsByProduct {
		if err := s.handleExpiredLineItem(
			ctx, tx, expiredItem, changeEventID,
		); err != nil {
			return nil, fmt.Errorf("failed to handle expired line item %s: %w", expiredItem.ProductID, err)
		}
	}

	// Save all history entries
	if len(lineItemHistories) > 0 {
		if err := s.repo.CreateBulkSubscriptionLineItemHistory(ctx, tx, lineItemHistories); err != nil {
			return nil, fmt.Errorf("failed to save history entries: %w", err)
		}
	}

	return resolvedItems, nil
}

func (s *playstoreSubscriptionService) updateLineItemAndNestedModels(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	subscriptionData models.SubscriptionPurchaseV2,
	existingItem *models.SubscriptionLineItem,
	newItem *androidpublisher.SubscriptionPurchaseLineItem,
	changeEventID uuid.UUID,
	notificationType rtdnModels.SubscriptionNotificationType,
) (*models.SubscriptionLineItem, *models.SubscriptionLineItemHistory, error) {
	var history *models.SubscriptionLineItemHistory
	previousExpiry := existingItem.ExpiryTime
	previousStatus := existingItem.Status

	newExpiryTime, err := time.Parse(time.RFC3339Nano, newItem.ExpiryTime)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid expiry time format: %w", err)
	}

	// Check for changes that require history entry
	if !existingItem.ExpiryTime.Equal(newExpiryTime) || existingItem.Status != models.LineItemStatusActive {
		history = &models.SubscriptionLineItemHistory{
			LineItemID:         existingItem.ID,
			PlanType:           existingItem.PlanType,
			PreviousExpiryTime: &previousExpiry,
			CurrentExpiryTime:  newExpiryTime,
			PreviousStatus:     &previousStatus,
			CurrentStatus:      models.LineItemStatusActive,
			ChangedAt:          time.Now(),
			ChangeEventID:      changeEventID,
			Reason:             "Updated from Play Store",
		}
	}

	// Update line item fields
	existingItem.ExpiryTime = newExpiryTime
	existingItem.Status = models.LineItemStatusActive

	// Handle nested models
	if err := s.resolveNestedModelsForUpdate(
		ctx, tx, subscriptionID, subscriptionData, existingItem, newItem, changeEventID,
	); err != nil {
		return nil, nil, fmt.Errorf("failed to resolve nested models: %w", err)
	}

	if err := tx.Save(existingItem).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to save updated line item: %w", err)
	}

	return existingItem, history, nil
}

func (s *playstoreSubscriptionService) createLineItemWithNestedModels(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	subscriptionData models.SubscriptionPurchaseV2,
	newLineItemData *androidpublisher.SubscriptionPurchaseLineItem,
	changeEventID uuid.UUID,
) (*models.SubscriptionLineItem, *models.SubscriptionLineItemHistory, error) {
	newExpiryTime, err := time.Parse(time.RFC3339Nano, newLineItemData.ExpiryTime)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid expiry time format: %w", err)
	}

	newLineItemModel := &models.SubscriptionLineItem{
		SubscriptionID: subscriptionID,
		ProductID:      newLineItemData.ProductId,
		ExpiryTime:     newExpiryTime,
		Status:         models.LineItemStatusActive,
		ItemType:       determineLineItemType(newLineItemData),
	}

	// Handle nested models
	if err := s.resolveNestedModelsForCreate(
		ctx, tx, subscriptionID, subscriptionData, newLineItemModel, newLineItemData, changeEventID,
	); err != nil {
		return nil, nil, fmt.Errorf("failed to resolve nested models: %w", err)
	}

	// Create the line item
	if err := tx.Create(newLineItemModel).Error; err != nil {
		return nil, nil, fmt.Errorf("failed to create line item: %w", err)
	}

	// Create history entry
	history := &models.SubscriptionLineItemHistory{
		LineItemID:        newLineItemModel.ID,
		PlanType:          newLineItemModel.PlanType,
		CurrentExpiryTime: newLineItemModel.ExpiryTime,
		CurrentStatus:     newLineItemModel.Status,
		ChangedAt:         time.Now(),
		ChangeEventID:     changeEventID,
		Reason:            "New line item from Play Store",
	}

	return newLineItemModel, history, nil
}

func (s *playstoreSubscriptionService) handleExpiredLineItem(
	ctx context.Context,
	tx *gorm.DB,
	expiredItem *models.SubscriptionLineItem,
	changeEventID uuid.UUID,
) error {
	// Handle nested models expiration first
	if err := s.expireNestedModels(
		ctx, tx, expiredItem, changeEventID,
	); err != nil {
		return fmt.Errorf("failed to expire nested models: %w", err)
	}

	// Mark line item as expired
	activeStatus := models.LineItemStatusActive
	expiredItem.Status = models.LineItemStatusExpired
	if err := tx.Save(expiredItem).Error; err != nil {
		return fmt.Errorf("failed to mark line item as expired: %w", err)
	}

	// Create history entry
	history := models.SubscriptionLineItemHistory{
		LineItemID:         expiredItem.ID,
		PlanType:           expiredItem.PlanType,
		PreviousExpiryTime: &expiredItem.ExpiryTime,
		CurrentExpiryTime:  expiredItem.ExpiryTime,
		PreviousStatus:     &activeStatus,
		CurrentStatus:      models.LineItemStatusExpired,
		ChangedAt:          time.Now(),
		ChangeEventID:      changeEventID,
		Reason:             "Expired (not in new data)",
	}

	if err := tx.Create(&history).Error; err != nil {
		return fmt.Errorf("failed to create line item history: %w", err)
	}

	return nil
}

func (s *playstoreSubscriptionService) resolveNestedModelsForUpdate(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	subscriptionData models.SubscriptionPurchaseV2,
	existingLineItemModel *models.SubscriptionLineItem,
	newLineItemData *androidpublisher.SubscriptionPurchaseLineItem,
	changeEventID uuid.UUID,
) error {
	// Handle AutoRenewingPlan
	if newLineItemData.AutoRenewingPlan != nil {
		recurringPrice := models.Money{
			CurrencyCode: newLineItemData.AutoRenewingPlan.RecurringPrice.CurrencyCode,
			Units:        newLineItemData.AutoRenewingPlan.RecurringPrice.Units,
			Nanos:        newLineItemData.AutoRenewingPlan.RecurringPrice.Nanos,
		}

		if existingLineItemModel.AutoRenewingPlanID != nil {
			// Update existing auto renewing plan
			if _, err := s.updateAutoRenewingPlan(
				ctx, tx, subscriptionID, existingLineItemModel,
				newLineItemData, recurringPrice, changeEventID,
			); err != nil {
				return fmt.Errorf("failed to update auto renewing plan: %w", err)
			}
		} else {

			// Create new auto renewing plan
			autoPlanID, err := s.createAutoRenewingPlan(
				ctx, tx, subscriptionID, existingLineItemModel,
				newLineItemData, recurringPrice, changeEventID,
			)
			if err != nil {
				return fmt.Errorf("failed to create auto renewing plan: %w", err)
			}
			existingLineItemModel.AutoRenewingPlanID = autoPlanID
		}
		existingLineItemModel.PlanType = models.PlanTypeAutoRenewing
	}

	// Handle PrepaidPlan
	if newLineItemData.PrepaidPlan != nil {
		if existingLineItemModel.PrepaidPlanID != nil {
			// Update existing prepaid plan
			if _, err := s.updatePrepaidPlan(
				ctx, tx, subscriptionID, existingLineItemModel,
				newLineItemData, changeEventID,
			); err != nil {
				return fmt.Errorf("failed to update prepaid plan: %w", err)
			}
		} else {
			// Create new prepaid plan
			prepaidPlanID, err := s.createPrepaidPlan(
				ctx, tx, subscriptionID, existingLineItemModel,
				newLineItemData, changeEventID,
			)
			if err != nil {
				return fmt.Errorf("failed to create prepaid plan: %w", err)
			}
			existingLineItemModel.PrepaidPlanID = prepaidPlanID
		}
		existingLineItemModel.PlanType = models.PlanTypePrepaid
	}

	// Handle OfferDetails
	if newLineItemData.OfferDetails != nil {
		if existingLineItemModel.OfferDetailsID != nil {
			if _, err := s.updateOfferDetails(
				ctx, tx, subscriptionID, subscriptionData, existingLineItemModel,
				newLineItemData, changeEventID,
			); err != nil {
				return fmt.Errorf("failed to update offer details: %w", err)
			}
		} else {
			offerDetailsID, err := s.createOfferDetails(
				ctx, tx, subscriptionID, subscriptionData, existingLineItemModel,
				newLineItemData, changeEventID,
			)
			if err != nil {
				return fmt.Errorf("failed to create offer details: %w", err)
			}
			existingLineItemModel.OfferDetailsID = offerDetailsID
		}
	}

	// Handle SignupPromo
	if newLineItemData.SignupPromotion != nil {
		if existingLineItemModel.SignupPromotionID != nil {
			if _, err := s.updateSignupPromo(
				ctx, tx, subscriptionID, existingLineItemModel.ID, newLineItemData, changeEventID,
			); err != nil {
				return fmt.Errorf("failed to update signup promo: %w", err)
			}
		} else {
			signupPromoID, err := s.createSignupPromotion(
				ctx, tx, subscriptionID, existingLineItemModel.ID,
				newLineItemData, changeEventID,
			)
			if err != nil {
				return fmt.Errorf("failed to create signup promo: %w", err)
			}
			existingLineItemModel.SignupPromotionID = signupPromoID
		}
	}

	// Handle DeferredItemReplacement
	if newLineItemData.DeferredItemReplacement != nil {
		if existingLineItemModel.DeferredItemReplacementID != nil {
			if _, err := s.updateDeferredItemReplacement(
				ctx, tx, subscriptionID, existingLineItemModel.ID,
				newLineItemData, changeEventID,
			); err != nil {
				return fmt.Errorf("failed to update deferred item replacement: %w", err)
			}
		} else {
			deferredItemID, err := s.createDeferredItemReplacement(
				ctx, tx, subscriptionID, existingLineItemModel.ID,
				newLineItemData, changeEventID,
			)
			if err != nil {
				return fmt.Errorf("failed to create deferred item replacement: %w", err)
			}
			existingLineItemModel.DeferredItemReplacementID = deferredItemID
		}
	}

	return nil
}

func (s *playstoreSubscriptionService) resolveNestedModelsForCreate(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	subscriptionData models.SubscriptionPurchaseV2,
	newLineItemModel *models.SubscriptionLineItem,
	newLineItemData *androidpublisher.SubscriptionPurchaseLineItem,
	changeEventID uuid.UUID,
) error {
	// Handle AutoRenewingPlan
	if newLineItemData.AutoRenewingPlan != nil {
		recurringPrice := models.Money{
			CurrencyCode: newLineItemData.AutoRenewingPlan.RecurringPrice.CurrencyCode,
			Units:        newLineItemData.AutoRenewingPlan.RecurringPrice.Units,
			Nanos:        newLineItemData.AutoRenewingPlan.RecurringPrice.Nanos,
		}

		autoPlanID, err := s.createAutoRenewingPlan(
			ctx, tx, subscriptionID, newLineItemModel,
			newLineItemData, recurringPrice, changeEventID,
		)
		if err != nil {
			return fmt.Errorf("failed to create auto renewing plan: %w", err)
		}
		newLineItemModel.AutoRenewingPlanID = autoPlanID
		newLineItemModel.PlanType = models.PlanTypeAutoRenewing
	}

	// Handle PrepaidPlan
	if newLineItemData.PrepaidPlan != nil {
		prepaidPlanID, err := s.createPrepaidPlan(
			ctx, tx, subscriptionID, newLineItemModel,
			newLineItemData, changeEventID,
		)
		if err != nil {
			return fmt.Errorf("failed to create prepaid plan: %w", err)
		}
		newLineItemModel.PrepaidPlanID = prepaidPlanID
		newLineItemModel.PlanType = models.PlanTypePrepaid
	}

	// Handle OfferDetails
	if newLineItemData.OfferDetails != nil {
		offerDetailsID, err := s.createOfferDetails(
			ctx, tx, subscriptionID, subscriptionData, newLineItemModel,
			newLineItemData, changeEventID,
		)
		if err != nil {
			return fmt.Errorf("failed to create offer details: %w", err)
		}
		newLineItemModel.OfferDetailsID = offerDetailsID
	}

	// Handle SignupPromo
	if newLineItemData.SignupPromotion != nil {
		signupPromoID, err := s.createSignupPromotion(
			ctx, tx, subscriptionID, newLineItemModel.ID,
			newLineItemData, changeEventID,
		)
		if err != nil {
			return fmt.Errorf("failed to create signup promo: %w", err)
		}
		newLineItemModel.SignupPromotionID = signupPromoID
	}

	// Handle DeferredItemReplacement
	if newLineItemData.DeferredItemReplacement != nil {
		deferredItemID, err := s.createDeferredItemReplacement(
			ctx, tx, subscriptionID, newLineItemModel.ID,
			newLineItemData, changeEventID,
		)
		if err != nil {
			return fmt.Errorf("failed to create deferred item replacement: %w", err)
		}
		newLineItemModel.DeferredItemReplacementID = deferredItemID
	}

	return nil
}

func (s *playstoreSubscriptionService) expireNestedModels(
	ctx context.Context,
	tx *gorm.DB,
	expiredLineItem *models.SubscriptionLineItem,
	changeEventID uuid.UUID,
) error {
	// Handle AutoRenewingPlan expiration
	if expiredLineItem.AutoRenewingPlanID != nil {
		if err := s.expireAutoRenewingPlan(
			ctx, tx, *expiredLineItem, changeEventID,
		); err != nil {
			return fmt.Errorf("failed to expire auto renewing plan: %w", err)
		}
	}

	// Handle PrepaidPlan expiration
	if expiredLineItem.PrepaidPlanID != nil {
		if err := s.expirePrepaidPlan(
			ctx, tx, *expiredLineItem, changeEventID,
		); err != nil {
			return fmt.Errorf("failed to expire prepaid plan: %w", err)
		}
	}

	// Handle OfferDetails expiration
	if expiredLineItem.OfferDetailsID != nil {
		if err := s.expireOfferDetails(
			ctx, tx, *expiredLineItem, changeEventID,
		); err != nil {
			return fmt.Errorf("failed to expire offer details: %w", err)
		}
	}

	// Handle SignupPromo expiration
	if expiredLineItem.SignupPromotionID != nil {
		if err := s.expireSignupPromotion(
			ctx, tx, *expiredLineItem, changeEventID,
		); err != nil {
			return fmt.Errorf("failed to expire signup promo: %w", err)
		}
	}

	// Handle DeferredItemReplacement expiration
	if expiredLineItem.DeferredItemReplacementID != nil {
		if err := s.expireDeferredItemReplacement(
			ctx, tx, *expiredLineItem, changeEventID,
		); err != nil {
			return fmt.Errorf("failed to expire deferred item replacement: %w", err)
		}
	}

	return nil
}

func determineLineItemType(item *androidpublisher.SubscriptionPurchaseLineItem) models.LineItemType {
	if strings.HasSuffix(item.ProductId, ".base") {
		return models.LineItemTypeBase
	}
	return models.LineItemTypeAddOn
}
