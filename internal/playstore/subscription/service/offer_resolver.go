package service

import (
	"context"
	"errors"
	"fmt"
	"subsnotifpro-go/internal/playstore/api/dto"
	"subsnotifpro-go/internal/playstore/subscription/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// createOfferDetails creates offer details record with complete pricing information
func (s *playstoreSubscriptionService) createOfferDetails(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	existingSubscription *models.SubscriptionPurchaseV2,
	newLineItemModel *models.SubscriptionLineItem,
	newLineItemData *dto.LineItem,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {
	if newLineItemData.OfferDetails == nil {
		return nil, nil
	}

	offer := newLineItemData.OfferDetails
	packageName := existingSubscription.PackageName
	productID := newLineItemData.ProductID
	basePlanID := offer.BasePlanID
	regionCode := existingSubscription.RegionCode
	startTime := existingSubscription.StartTime

	// 1. Get base plan price (required for both base and discounted offers)
	basePlanPrice, err := s.GetRegionalBasePlanPrice(
		ctx,
		packageName,
		productID,
		basePlanID,
		regionCode,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get base plan price: %w", err)
	}

	// Initialize variables
	var (
		offerID      *string
		currentPhase *int
		currentPrice models.Money
	)

	// Handle offer tags (direct slice pointer)
	offerTags := &offer.OfferTags
	if len(*offerTags) == 0 {
		offerTags = nil // Set to nil if empty
	}

	// For discounted offers, get phase information
	if offer.OfferID != "" {
		offerID = &offer.OfferID

		// Get current phase index
		phaseIndex, err := s.GetCurrentOfferPhaseIndex(
			ctx,
			packageName,
			productID,
			basePlanID,
			offer.OfferID,
			startTime,
			time.Now(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get current offer phase: %w", err)
		}
		currentPhase = phaseIndex

		if phaseIndex != nil {
			// Get phase-specific price
			phasePrice, err := s.GetRegionalOfferPhasePrice(
				ctx,
				packageName,
				productID,
				offer.BasePlanID,
				offer.OfferID,
				*phaseIndex,
				regionCode,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to get offer phase price: %w", err)
			}
			currentPrice = *phasePrice
		} else {
			currentPrice = *basePlanPrice
		}
	} else {
		// For base plans without offers, use base price
		currentPrice = *basePlanPrice
	}

	// Create the offer details record
	offerDetails := &models.OfferDetails{
		SubscriptionID: subscriptionID,
		LineItemID:     newLineItemModel.ID,
		// PackageName:            packageName,
		// ProductID:              productID,
		BasePlanID:             basePlanID,
		OfferID:                offerID,
		OfferTags:              offerTags,
		BasePlanPrice:          *basePlanPrice,
		CurrentOfferPhaseIndex: currentPhase,
		CurrentPhasePrice:      currentPrice,
	}

	if err := tx.Create(offerDetails).Error; err != nil {
		return nil, fmt.Errorf("failed to create offer details: %w", err)
	}

	// Create history entry
	history := models.OfferDetailsHistory{
		OfferDetailsID: offerDetails.ID,
		SubscriptionID: subscriptionID,
		LineItemID:     newLineItemModel.ID,
		// PackageName:            packageName,
		// ProductID:              productID,
		BasePlanID:             basePlanID,
		OfferID:                offerID,
		OfferTags:              offerTags,
		CurrentOfferPhaseIndex: currentPhase,
		CurrentBasePlanPrice:   *basePlanPrice,
		CurrentPhasePrice:      currentPrice,
		ChangeEventID:          changeEventID,
		ChangeType:             "ADDED",
	}

	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create offer details history: %w", err)
	}

	return &offerDetails.ID, nil
}

func (s *playstoreSubscriptionService) updateOfferDetails(
	ctx context.Context,
	tx *gorm.DB,
	subscriptionID uuid.UUID,
	existingSubscription *models.SubscriptionPurchaseV2,
	existingLineItemModel *models.SubscriptionLineItem,
	newLineItemData *dto.LineItem,
	changeEventID uuid.UUID,
) (*uuid.UUID, error) {
	if newLineItemData.OfferDetails == nil {
		return nil, nil
	}

	if existingLineItemModel.OfferDetails == nil {
		return nil, errors.New("no existing offer details found")
	}

	existing := existingLineItemModel.OfferDetails

	offer := newLineItemData.OfferDetails

	packageName := existingSubscription.PackageName
	productID := newLineItemData.ProductID
	regionCode := existingSubscription.RegionCode
	startTime := existingSubscription.StartTime

	// 1. Get base plan price (required for both base and discounted offers)
	basePlanPrice, err := s.GetRegionalBasePlanPrice(
		ctx,
		packageName,
		productID,
		offer.BasePlanID,
		regionCode,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get base plan price: %w", err)
	}

	// Initialize variables
	var (
		offerID      *string
		currentPhase *int
		currentPrice models.Money
	)

	// Handle offer tags (direct slice pointer)
	offerTags := &offer.OfferTags
	if len(*offerTags) == 0 {
		offerTags = nil // Set to nil if empty
	}

	// For discounted offers, get phase information
	if offer.OfferID != "" {
		offerID = &offer.OfferID

		// Get current phase index
		phaseIndex, err := s.GetCurrentOfferPhaseIndex(
			ctx,
			packageName,
			productID,
			offer.BasePlanID,
			offer.OfferID,
			startTime,
			time.Now(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to get current offer phase: %w", err)
		}
		currentPhase = phaseIndex

		if phaseIndex != nil {
			// Get phase-specific price
			phasePrice, err := s.GetRegionalOfferPhasePrice(
				ctx,
				packageName,
				productID,
				offer.BasePlanID,
				offer.OfferID,
				*phaseIndex,
				regionCode,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to get offer phase price: %w", err)
			}
			currentPrice = *phasePrice
		} else {
			currentPrice = *basePlanPrice
		}

	} else {
		// For base plans without offers, use base price
		currentPrice = *basePlanPrice
	}

	// Prepare updates
	updates := map[string]interface{}{
		"base_plan_id":              offer.BasePlanID,
		"offer_id":                  offerID,
		"offer_tags":                offerTags,
		"base_plan_price":           *basePlanPrice,
		"current_offer_phase_index": currentPhase,
		"current_phase_price":       currentPrice,
	}

	// Create history entry before updating
	history := models.OfferDetailsHistory{
		OfferDetailsID: existing.ID,
		SubscriptionID: subscriptionID,
		LineItemID:     existingLineItemModel.ID,
		// PackageName:             packageName,
		// ProductID:               productID,
		BasePlanID:              offer.BasePlanID,
		OfferID:                 offerID,
		OfferTags:               offerTags,
		PreviousOfferPhaseIndex: existing.CurrentOfferPhaseIndex,
		CurrentOfferPhaseIndex:  currentPhase,
		CurrentBasePlanPrice:    *basePlanPrice,
		CurrentPhasePrice:       currentPrice,
		PreviousBasePlanID:      &existing.BasePlanID,
		PreviousOfferID:         existing.OfferID,
		PreviousOfferTags:       existing.OfferTags,
		PreviousBasePlanPrice:   &existing.BasePlanPrice,
		PreviousPhasePrice:      &existing.CurrentPhasePrice,
		ChangeEventID:           changeEventID,
		ChangeType:              "UPDATED",
	}

	// Update the record
	if err := tx.Model(&existing).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update offer details: %w", err)
	}

	// Create history entry
	if err := tx.Create(&history).Error; err != nil {
		return nil, fmt.Errorf("failed to create offer details history: %w", err)
	}

	return &existing.ID, nil
}

// OfferDetails expiration
func (s *playstoreSubscriptionService) expireOfferDetails(
	ctx context.Context,
	tx *gorm.DB,
	expiredLineItem models.SubscriptionLineItem,
	changeEventID uuid.UUID,
) error {
	// Verify we have offer details to expire
	if expiredLineItem.OfferDetails == nil {
		return nil // Nothing to expire
	}

	offer := expiredLineItem.OfferDetails

	// Create comprehensive history before soft deleting
	history := models.OfferDetailsHistory{
		OfferDetailsID: offer.ID,
		SubscriptionID: offer.SubscriptionID,
		// PackageName:             expiredLineItem.OfferDetails.PackageName,
		// ProductID:               expiredLineItem.ProductID,
		LineItemID:              offer.LineItemID,
		BasePlanID:              offer.BasePlanID,
		CurrentBasePlanPrice:    offer.BasePlanPrice,
		PreviousBasePlanPrice:   &offer.BasePlanPrice,
		PreviousPhasePrice:      &offer.CurrentPhasePrice,
		PreviousOfferPhaseIndex: offer.CurrentOfferPhaseIndex,
		PreviousOfferTags:       offer.OfferTags,
		PreviousOfferID:         offer.OfferID,
		ChangeEventID:           changeEventID,
		ChangeType:              "REMOVED",
	}

	if err := tx.Create(&history).Error; err != nil {
		return fmt.Errorf("failed to create offer details history: %w", err)
	}

	// Soft delete the offer details
	if err := tx.Delete(offer).Error; err != nil {
		return fmt.Errorf("failed to soft delete offer details: %w", err)
	}

	return nil
}

// GetBasePlanPrice retrieves the base plan price for a specific region
func (s *playstoreSubscriptionService) GetRegionalBasePlanPrice(
	ctx context.Context,
	packageName, productID, basePlanID, regionCode string,
) (*models.Money, error) {
	// Get price from catalog service
	catalogPrice, err := s.playstoreCatalogService.GetRegionalBasePlanPrice(packageName, productID, basePlanID, regionCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get base plan price: %w", err)
	}

	// Convert catalog money to our model
	return &models.Money{
		CurrencyCode: catalogPrice.CurrencyCode,
		Units:        catalogPrice.Units,
		Nanos:        catalogPrice.Nanos,
	}, nil
}

// GetBasePlanPrice retrieves the base plan price for a specific region
func (s *playstoreSubscriptionService) GetCurrentOfferPhaseIndex(
	ctx context.Context,
	packageName, productID, basePlanID, offerID string, startTime, checkTime time.Time,
) (*int, error) {
	// Get price from catalog service
	currentPhaseIndex, err := s.playstoreCatalogService.GetCurrentOfferPhaseIndex(packageName, productID, basePlanID, offerID, startTime, checkTime)
	if err != nil {
		return nil, fmt.Errorf("failed to determine current phase for %s/%s/%s/%s: %w",
			packageName, productID, basePlanID, offerID, err)
	}

	// Convert catalog money to our model
	return currentPhaseIndex, nil
}

// GetRegionalOfferPhasePrice calls the CatalogService to fetch the current offer phase price for a subscription
func (s *playstoreSubscriptionService) GetRegionalOfferPhasePrice(
	ctx context.Context,
	packageName string,
	productID string,
	basePlanID string,
	offerID string,
	phaseIndex int,
	regionCode string,
) (*models.Money, error) {
	// Input validation (optional, can be adjusted as needed)
	if packageName == "" || productID == "" || regionCode == "" {
		return nil, errors.New("packageName, productID, and regionCode are required")
	}

	// Call the CatalogService's GetRegionalOfferPhasePrice function
	price, err := s.playstoreCatalogService.GetRegionalOfferPhasePrice(
		packageName,
		productID,
		basePlanID,
		offerID,
		phaseIndex,
		regionCode,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get offer phase price: %w", err)
	}

	// Convert catalog money to our model
	return &models.Money{
		CurrencyCode: price.CurrencyCode,
		Units:        price.Units,
		Nanos:        price.Nanos,
	}, nil
}
