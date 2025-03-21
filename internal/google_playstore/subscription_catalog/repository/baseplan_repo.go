package repository

import (
	"context"
	"errors"
	"fmt"
	"subsnotifpro-go/internal/google_playstore/subscription_catalog/models"
	"subsnotifpro-go/internal/google_playstore/subscription_catalog/utils"

	"gorm.io/gorm"
)

// ✅ Get Base Plan Details (With All Nested Data)
func (r *subscriptionCatalogRepository) GetBasePlanDetails(ctx context.Context, packageName, productID, basePlanID string) (*models.ProductBasePlan, error) {
	var basePlan models.ProductBasePlan

	err := r.db.WithContext(ctx).
		Preload("RegionalConfigs").
		Preload("OfferTags").
		Preload("OtherRegionsConfig").
		Preload("AutoRenewingBasePlanType").
		Preload("PrepaidBasePlanType").
		Preload("InstallmentsBasePlanType").
		Where("package_name = ? AND product_id = ? AND base_plan_id = ?", packageName, productID, basePlanID).
		First(&basePlan).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("base plan not found: %s", basePlanID)
		}
		return nil, fmt.Errorf("error fetching base plan: %w", err)
	}

	return &basePlan, nil
}

// ✅ Check if Base Plan is Active
func (r *subscriptionCatalogRepository) IsBasePlanActive(ctx context.Context, packageName, productID, basePlanID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ProductBasePlan{}).
		Where("package_name = ? AND product_id = ? AND base_plan_id = ? AND state = ?",
			packageName, productID, basePlanID, "ACTIVE").
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("error checking base plan state: %w", err)
	}

	return count > 0, nil
}

// ✅ Check If Base Plan is Available in a Specific Region (Optimized)
func (r *subscriptionCatalogRepository) IsBasePlanAvailableInRegion(
	ctx context.Context, packageName, productID, basePlanID, regionCode string,
) (bool, error) {
	var exists bool

	// ✅ Use Model instead of Table Name
	err := r.db.WithContext(ctx).
		Model(&models.BasePlanRegionalConfig{}). // ✅ Use GORM model
		Select("1").                             // ✅ Efficient EXISTS-like check
		Where("package_name = ? AND product_id = ? AND base_plan_id = ? AND region_code = ?",
					packageName, productID, basePlanID, regionCode).
		First(&exists).Error // ✅ First() is optimized for existence checks

	// ✅ Handle Errors
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // ✅ No record found means Base Plan is NOT available
		}
		return false, fmt.Errorf("error checking base plan availability in region: %w", err)
	}

	return true, nil // ✅ If a record is found, Base Plan is available
}

// ✅ List All Base Plan Names of a Subscription
func (r *subscriptionCatalogRepository) ListBasePlanNames(ctx context.Context, packageName, productID string) ([]string, error) {
	var basePlans []string

	err := r.db.WithContext(ctx).
		Model(&models.ProductBasePlan{}).
		Distinct("base_plan_id").
		Where("package_name = ? AND product_id = ?", packageName, productID).
		Order("base_plan_id ASC").
		Pluck("base_plan_id", &basePlans).Error

	if err != nil {
		return nil, fmt.Errorf("error fetching base plans: %w", err)
	}

	return basePlans, nil
}

// ✅ Get Regional Base Plan Price
func (r *subscriptionCatalogRepository) GetRegionalBasePlanPrice(
	ctx context.Context, packageName, productID, basePlanID, regionCode string,
) (*models.Money, error) {
	var regionalConfig models.BasePlanRegionalConfig

	err := r.db.WithContext(ctx).
		Where("package_name = ? AND product_id = ? AND base_plan_id = ? AND region_code = ?",
			packageName, productID, basePlanID, regionCode).
		First(&regionalConfig).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no base plan price found for region: %s", regionCode)
		}
		return nil, fmt.Errorf("error fetching regional base plan price: %w", err)
	}

	return &regionalConfig.Price, nil
}

// ✅ Get Other Regions Base Plan Price (With Currency Selection)
func (r *subscriptionCatalogRepository) GetOtherRegionsBasePlanPrice(
	ctx context.Context, packageName, productID, basePlanID, currency string,
) (*models.Money, error) {
	var otherRegionsConfig models.OtherRegionsBasePlanConfig

	err := r.db.WithContext(ctx).
		Where("package_name = ? AND product_id = ? AND base_plan_id = ?",
			packageName, productID, basePlanID).
		First(&otherRegionsConfig).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no base plan price found in other regions")
		}
		return nil, fmt.Errorf("error fetching other regions base plan price: %w", err)
	}

	// ✅ Check if the price is valid using a helper function
	switch currency {
	case "USD":
		if !utils.IsZeroMoney(otherRegionsConfig.USDPrice) {
			return &otherRegionsConfig.USDPrice, nil
		}
	case "EUR":
		if !utils.IsZeroMoney(otherRegionsConfig.EURPrice) {
			return &otherRegionsConfig.EURPrice, nil
		}
	default:
		return nil, fmt.Errorf("unsupported currency (valid [USD,EUR]): %s", currency)
	}

	return nil, fmt.Errorf("no valid price found for currency: %s", currency)
}
