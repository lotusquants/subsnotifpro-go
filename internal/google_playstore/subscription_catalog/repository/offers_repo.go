package repository

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/google_playstore/subscription_catalog/models"

	"gorm.io/gorm"
)

// ✅ Get Subscription Offer with ALL Nested Fields
func (r *subscriptionCatalogRepository) GetSubscriptionOffer(ctx context.Context, packageName, productID, basePlanID, offerID string) (*models.SubscriptionOffer, error) {
	var offer models.SubscriptionOffer

	err := r.db.WithContext(ctx).
		Preload("Phases").
		Preload("Phases.RegionalConfigs").
		Preload("Phases.OtherRegionsConfig.OtherRegionsPrices").
		Preload("Targeting.AcquisitionRule").
		Preload("Targeting.UpgradeRule").
		Preload("RegionalConfigs").
		Preload("OtherRegionsConfig").
		Where("package_name = ? AND product_id = ? AND base_plan_id = ? AND offer_id = ?",
			packageName, productID, basePlanID, offerID).
		First(&offer).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("subscription offer not found")
		}
		return nil, fmt.Errorf("error fetching subscription offer: %w", err)
	}

	return &offer, nil
}

// ✅ List All Offer Names for a Base Plan
func (r *subscriptionCatalogRepository) ListOfferNamesForBasePlan(
	ctx context.Context, packageName, productID, basePlanID string,
) ([]string, error) {
	var offerNames []string

	err := r.db.WithContext(ctx).
		Model(&models.SubscriptionOffer{}). // ✅ Use model instead of table name
		Select("DISTINCT offer_id").
		Where("package_name = ? AND product_id = ? AND base_plan_id = ?", packageName, productID, basePlanID).
		Order("offer_id ASC").
		Pluck("offer_id", &offerNames).Error

	if err != nil {
		return nil, fmt.Errorf("error fetching offer names for base plan: %w", err)
	}

	return offerNames, nil
}

// ✅ Check if Subscription Offer is Active (Optimized)
func (r *subscriptionCatalogRepository) IsSubscriptionOfferActive(
	ctx context.Context, packageName, productID, basePlanID, offerID string,
) (bool, error) {
	var exists bool

	// ✅ Use Model instead of Table Name
	err := r.db.WithContext(ctx).
		Model(&models.SubscriptionOffer{}). // ✅ Use Model
		Select("1").                        // ✅ Select 1 (Optimized for EXISTS check)
		Where("package_name = ? AND product_id = ? AND base_plan_id = ? AND offer_id = ? AND state = ?",
					packageName, productID, basePlanID, offerID, "ACTIVE").
		First(&exists).Error // ✅ Use First instead of Count

	// ✅ Handle Errors
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // ✅ Return false if no record found
		}
		return false, fmt.Errorf("error checking subscription offer status: %w", err)
	}

	return true, nil // ✅ Return true if found
}

// ✅ IsOfferAvailableInRegion - Checks if an offer is available in a specific region
func (r *subscriptionCatalogRepository) IsSubscriptionOfferAvailableInRegion(
	ctx context.Context, packageName, productID, basePlanID, offerID, regionCode string,
) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).
		Model(&models.RegionalSubscriptionOfferConfig{}).
		Select("COUNT(*) > 0").
		Where("package_name = ? AND product_id = ? AND base_plan_id = ? AND offer_id = ? AND region_code = ? AND new_subscriber_availability = TRUE",
			packageName, productID, basePlanID, offerID, regionCode).
		Find(&exists).Error

	if err != nil {
		return false, fmt.Errorf("error checking offer availability in region: %w", err)
	}

	return exists, nil
}

// ✅ Get Subscription Offer Phases (With Nested Data)
func (r *subscriptionCatalogRepository) GetSubscriptionOfferPhases(ctx context.Context, packageName, productID, basePlanID, offerID string) ([]models.SubscriptionOfferPhase, error) {
	var phases []models.SubscriptionOfferPhase

	err := r.db.WithContext(ctx).
		Preload("RegionalConfigs").
		Preload("OtherRegionsConfig").
		Where("package_name = ? AND product_id = ? AND base_plan_id = ? AND offer_id = ?",
			packageName, productID, basePlanID, offerID).
		Find(&phases).Error

	if err != nil {
		return nil, fmt.Errorf("error fetching offer phases: %w", err)
	}
	return phases, nil
}

// ✅ Check if a Subscription Offer Phase Exists (Optimized)
func (r *subscriptionCatalogRepository) IsSubscriptionOfferPhaseExists(
	ctx context.Context, packageName, productID, basePlanID, offerID string,
) (bool, error) {
	var exists bool

	// ✅ Use Model instead of Raw SQL
	err := r.db.WithContext(ctx).
		Model(&models.SubscriptionOfferPhase{}). // ✅ Use GORM Model
		Select("1").                             // ✅ Select 1 (Mimics EXISTS)
		Where("package_name = ? AND product_id = ? AND base_plan_id = ? AND offer_id = ?",
					packageName, productID, basePlanID, offerID).
		First(&exists).Error // ✅ Use First() for efficient EXISTS check

	// ✅ Handle Errors
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // ✅ Return false if no record found
		}
		return false, fmt.Errorf("error checking subscription offer phase existence: %w", err)
	}

	return true, nil // ✅ Return true if a record exists
}

// ✅ GetRegionalOfferPhasePrice - Fetches price for a specific offer phase in a given region
func (r *subscriptionCatalogRepository) GetRegionalOfferPhaseConfig(
	ctx context.Context, packageName, productID, basePlanID, offerID string, phaseIndex int, regionCode string,
) (*models.RegionalSubscriptionOfferPhaseConfig, error) {
	var regionalPhaseConfig models.RegionalSubscriptionOfferPhaseConfig

	err := r.db.WithContext(ctx).
		Where("package_name = ? AND product_id = ? AND base_plan_id = ? AND offer_id = ? AND phase_index = ? AND region_code = ?",
			packageName, productID, basePlanID, offerID, phaseIndex, regionCode).
		First(&regionalPhaseConfig).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no offer phase config found for region: %s", regionCode)
		}
		return nil, fmt.Errorf("error fetching regional offer phase config: %w", err)
	}

	return &regionalPhaseConfig, nil
}

// ✅ GetOtherRegionsOfferPhaseConfig - Fetches full offer phase config for other regions
func (r *subscriptionCatalogRepository) GetOtherRegionsOfferPhaseConfig(
	ctx context.Context, packageName, productID, basePlanID, offerID string, phaseIndex int,
) (*models.OtherRegionsSubscriptionOfferPhaseConfig, error) {
	var otherRegionsPhaseConfig models.OtherRegionsSubscriptionOfferPhaseConfig

	err := r.db.WithContext(ctx).
		Preload("OtherRegionsPrices"). // ✅ Preload nested struct for prices
		Where("package_name = ? AND product_id = ? AND base_plan_id = ? AND offer_id = ? AND phase_index = ?",
			packageName, productID, basePlanID, offerID, phaseIndex).
		First(&otherRegionsPhaseConfig).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no offer phase config found in other regions")
		}
		return nil, fmt.Errorf("error fetching other regions offer phase config: %w", err)
	}

	return &otherRegionsPhaseConfig, nil
}
