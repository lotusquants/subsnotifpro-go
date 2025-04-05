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

func (s *playstoreSubscriptionService) createSignupPromotion(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	newLineItemID uuid.UUID,
	newLineItemData *dto.LineItem,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {
	if newLineItemData.SignupPromotion == nil {
		return nil, nil
	}

	// Determine promotion type and code from the input data
	var promoType models.SignupPromotionType
	var promoCode *string

	if newLineItemData.SignupPromotion.Type == dto.PromoTypeVanityCode {
		promoType = models.SignupPromotionTypeVanity
		promoCode = newLineItemData.SignupPromotion.Code
	} else if newLineItemData.SignupPromotion.Type == dto.PromoTypeOneTimeCode {
		promoType = models.SignupPromotionTypeOneTime
		promoCode = nil
	} else {
		return nil, errors.New("invalid signup promotion type")
	}

	// Create the signup promotion record
	signupPromo := &models.SignupPromotion{
		SubscriptionID: subscriptionID,
		LineItemID:     newLineItemID,
		PromotionType:  promoType,
		PromotionCode:  promoCode,
	}

	if err := tx.Create(signupPromo).Error; err != nil {
		return nil, fmt.Errorf("failed to create signup promotion: %w", err)
	}

	// Create the history record
	history := models.SignupPromotionHistory{
		SubscriptionID: subscriptionID,
		LineItemID:     newLineItemID,
		PromotionType:  promoType,
		PromotionCode:  promoCode,
		ChangeType:     "ADDED",
		ChangeEventID:  changeEventID,
	}

	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create signup promotion history: %w", err)
	}

	return &signupPromo.ID, nil
}

func (s *playstoreSubscriptionService) updateSignupPromo(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	existingLineItemID uuid.UUID,
	newLineItemData *dto.LineItem,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {
	if newLineItemData.SignupPromotion == nil {
		// If no promotion in new data, consider removing existing promotion
		return nil, nil
	}

	// Get existing promotion
	var existingPromo models.SignupPromotion
	if err := tx.Where("line_item_id = ?", existingLineItemID).
		First(&existingPromo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No existing promotion found, treat as create
			return s.createSignupPromotion(ctx, tx, subscriptionID, existingLineItemID, newLineItemData, changeEventID)
		}
		return nil, fmt.Errorf("failed to find existing signup promotion: %w", err)
	}

	// Determine new promotion type and code
	var newPromoType models.SignupPromotionType
	var newPromoCode *string

	if newLineItemData.SignupPromotion.Type == dto.PromoTypeVanityCode {
		newPromoType = models.SignupPromotionTypeVanity
		newPromoCode = newLineItemData.SignupPromotion.Code
	} else if newLineItemData.SignupPromotion.Type == dto.PromoTypeOneTimeCode {
		newPromoType = models.SignupPromotionTypeOneTime
		newPromoCode = nil
	} else {
		return nil, errors.New("invalid signup promotion type")
	}

	// Check if promotion actually changed
	if existingPromo.PromotionType == newPromoType {
		if newPromoType == models.SignupPromotionTypeVanity &&
			existingPromo.PromotionCode != nil &&
			newPromoCode != nil &&
			*existingPromo.PromotionCode == *newPromoCode {
			// No changes needed
			return &existingPromo.ID, nil
		}
		if newPromoType == models.SignupPromotionTypeOneTime {
			// One-time codes can't change
			return &existingPromo.ID, nil
		}
	}

	// Prepare updates
	updates := map[string]interface{}{
		"promotion_type": newPromoType,
		"promotion_code": newPromoCode,
	}

	// Create history before updating
	history := models.SignupPromotionHistory{
		SubscriptionID:        subscriptionID,
		LineItemID:            existingLineItemID,
		PreviousPromotionType: &existingPromo.PromotionType,
		PreviousPromotionCode: existingPromo.PromotionCode,
		PromotionType:         newPromoType,
		PromotionCode:         newPromoCode,
		ChangeType:            "UPDATED",
		ChangeEventID:         changeEventID,
	}

	// Update the promotion
	if err := tx.Model(&existingPromo).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update signup promotion: %w", err)
	}

	// Create history entry
	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create signup promotion history: %w", err)
	}

	return &existingPromo.ID, nil
}

// SignupPromo expiration
func (s *playstoreSubscriptionService) expireSignupPromotion(
	ctx context.Context,
	tx *gorm.DB,
	expiredLineItem models.SubscriptionLineItem,
	changeEventID uuid.UUID,
) error {
	// Verify we have a signup promotion to expire
	if expiredLineItem.SignupPromotion == nil {
		return nil // Nothing to expire
	}

	promo := expiredLineItem.SignupPromotion

	// Create comprehensive history before soft deleting
	history := models.SignupPromotionHistory{
		SubscriptionID:        promo.SubscriptionID,
		LineItemID:            promo.LineItemID,
		PreviousPromotionType: &promo.PromotionType,
		PreviousPromotionCode: promo.PromotionCode,
		ChangeType:            "REMOVED",
		ChangeEventID:         changeEventID,
	}

	if err := tx.Create(&history).Error; err != nil {
		return fmt.Errorf("failed to create signup promotion history: %w", err)
	}

	// Soft delete the promotion
	if err := tx.Delete(promo).Error; err != nil {
		return fmt.Errorf("failed to soft delete signup promotion: %w", err)
	}

	return nil
}
