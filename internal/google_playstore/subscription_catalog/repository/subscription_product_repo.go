package repository

import (
	"context"
	"errors"
	"fmt"
	"subsnotifpro-go/internal/google_playstore/subscription_catalog/models"

	"gorm.io/gorm"
)

// ✅ List Subscription Products (Only Package Name & Product ID)
func (r *subscriptionCatalogRepository) ListSubscriptionProducts(ctx context.Context) ([]models.ProductSubscription, error) {
	var products []models.ProductSubscription

	err := r.db.WithContext(ctx).
		Select("package_name, product_id").
		Find(&products).Error

	if err != nil {
		return nil, fmt.Errorf("error fetching subscription products: %w", err)
	}

	return products, nil
}

// ✅ Get Full Subscription Product Details (With Base Plans)
func (r *subscriptionCatalogRepository) GetSubscriptionProduct(ctx context.Context, packageName, productID string) (*models.ProductSubscription, error) {
	var product models.ProductSubscription

	err := r.db.WithContext(ctx).
		Preload("BasePlans").
		Preload("BasePlans.RegionalConfigs").          // ✅ Load regional pricing
		Preload("BasePlans.OfferTags").                // ✅ Load offer tags
		Preload("BasePlans.OtherRegionsConfig").       // ✅ Load Other Regions pricing
		Preload("BasePlans.AutoRenewingBasePlanType"). // ✅ Load Auto-Renewing plans
		Preload("BasePlans.PrepaidBasePlanType").      // ✅ Load Prepaid plans
		Preload("BasePlans.InstallmentsBasePlanType"). // ✅ Load Installments plans
		Preload("Listings").
		Preload("TaxAndComplianceSettings").
		Preload("RestrictedPaymentCountries").
		Where("package_name = ? AND product_id = ?", packageName, productID).
		First(&product).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("subscription product not found: %s", productID)
		}
		return nil, fmt.Errorf("error fetching subscription product: %w", err)
	}

	return &product, nil
}

// ✅ Check if Subscription Product Exists (Optimized)
func (r *subscriptionCatalogRepository) CheckSubscriptionProductExists(
	ctx context.Context, packageName, productID string,
) (bool, error) {
	var exists bool

	// ✅ Use `Select("1")` + `First()` for Efficient Existence Check
	err := r.db.WithContext(ctx).
		Model(&models.ProductSubscription{}). // ✅ Use GORM model
		Select("1").
		Where("package_name = ? AND product_id = ?", packageName, productID).
		First(&exists).Error // ✅ Stops at first match (efficient)

	// ✅ Handle Errors
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // ✅ No record found means Subscription Product does not exist
		}
		return false, fmt.Errorf("error checking if subscription product exists: %w", err)
	}

	return true, nil // ✅ If found, Subscription Product exists
}
