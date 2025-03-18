package repository

import (
	"context"
	"errors"
	"fmt"

	"subsnotifpro-go/internal/google_playstore/subscription_catalog/models"
	"subsnotifpro-go/internal/google_playstore/subscription_catalog/utils"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Custom errors for better error handling
var (
	ErrNotFound      = errors.New("record not found")
	ErrInactiveOffer = errors.New("offer is inactive")
)

// SubscriptionCatalogRepository defines methods for managing subscription catalog data.
type SubscriptionCatalogRepository interface {
	UpsertSubscriptionProducts(ctx context.Context, subscription *models.ProductSubscription) error
	DeleteAllSubscriptionProducts(ctx context.Context) error
	UpsertSubscriptionBasePlans(ctx context.Context, basePlan *models.ProductBasePlan) error
	UpsertSubscriptionOffers(ctx context.Context, offers []models.SubscriptionOffer) error
	DeleteAllSubscriptionOffers(ctx context.Context, packageName, productID, basePlanID string) error

	ListSubscriptionProducts(ctx context.Context) ([]models.ProductSubscription, error)
	GetSubscriptionProduct(ctx context.Context, packageName, productID string) (*models.ProductSubscription, error)
	CheckSubscriptionProductExists(ctx context.Context, packageName, productID string) (bool, error)

	GetBasePlanDetails(ctx context.Context, packageName, productID, basePlanID string) (*models.ProductBasePlan, error)
	IsBasePlanActive(ctx context.Context, packageName, productID, basePlanID string) (bool, error)
	IsBasePlanAvailableInRegion(ctx context.Context, packageName, productID, basePlanID, regionCode string) (bool, error)
	ListBasePlanNames(ctx context.Context, packageName, productID string) ([]string, error)
	GetRegionalBasePlanPrice(ctx context.Context, packageName, productID, basePlanID, regionCode string) (*models.Money, error)
	GetOtherRegionsBasePlanPrice(ctx context.Context, packageName, productID, basePlanID, currency string) (*models.Money, error)

	GetSubscriptionOffer(ctx context.Context, packageName, productID, basePlanID, offerID string) (*models.SubscriptionOffer, error)
	ListOfferNamesForBasePlan(ctx context.Context, packageName, productID, basePlanID string) ([]string, error)
	IsSubscriptionOfferActive(ctx context.Context, packageName, productID, basePlanID, offerID string) (bool, error)

	GetSubscriptionOfferPhases(ctx context.Context, packageName, productID, basePlanID, offerID string) ([]models.SubscriptionOfferPhase, error)
	IsSubscriptionOfferPhaseExists(ctx context.Context, packageName, productID, basePlanID, offerID string) (bool, error)
	GetRegionalOfferPhaseConfig(ctx context.Context, packageName, productID, basePlanID, offerID string, phaseIndex int, regionCode string) (*models.RegionalSubscriptionOfferPhaseConfig, error)
	GetOtherRegionsOfferPhaseConfig(ctx context.Context, packageName, productID, basePlanID, offerID string, phaseIndex int) (*models.OtherRegionsSubscriptionOfferPhaseConfig, error)
}

// Concrete implementation
type subscriptionCatalogRepository struct {
	db        *gorm.DB
	batchSize int
}

// ✅ Ensure struct implements the interface at compile-time
var _ SubscriptionCatalogRepository = (*subscriptionCatalogRepository)(nil)

// NewSubscriptionCatalogRepository creates a new instance.
func NewSubscriptionCatalogRepository(db *gorm.DB, batchSize int) SubscriptionCatalogRepository {
	return &subscriptionCatalogRepository{
		db:        db,
		batchSize: batchSize,
	}
}

// ✅ Upsert Subscription Product (Insert or Update inside Transaction)
func (r *subscriptionCatalogRepository) UpsertSubscriptionProducts(ctx context.Context, subscription *models.ProductSubscription) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "package_name"}, {Name: "product_id"}}, // ✅ Unique constraints
			DoUpdates: clause.AssignmentColumns([]string{"package_name", "product_id"}),
		}).Create(subscription).Error

		if err != nil {
			return fmt.Errorf("failed to upsert subscription product: %w", err)
		}

		return nil
	})
}

func (r *subscriptionCatalogRepository) DeleteAllSubscriptionProducts(ctx context.Context) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("1 = 1").Delete(&models.ProductSubscription{}).Error; err != nil {
			return fmt.Errorf("failed to delete all subscription products: %w", err)
		}
		return nil
	})
}

// ✅ Upsert Base Plan (Insert or Update inside Transaction)
func (r *subscriptionCatalogRepository) UpsertSubscriptionBasePlans(ctx context.Context, basePlan *models.ProductBasePlan) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "package_name"}, {Name: "product_id"}, {Name: "base_plan_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"package_name", "product_id", "base_plan_id"}),
		}).Create(basePlan).Error

		if err != nil {
			return fmt.Errorf("failed to upsert base plan: %w", err)
		}

		return nil
	})
}

// ✅ Upsert Subscription Offers (Bulk Insert)
func (r *subscriptionCatalogRepository) UpsertSubscriptionOffers(ctx context.Context, offers []models.SubscriptionOffer) error {
	if len(offers) == 0 {
		return nil // No offers to insert
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		batchSize := r.batchSize // Use configured batch size
		for i := 0; i < len(offers); i += batchSize {
			end := i + batchSize
			if end > len(offers) {
				end = len(offers)
			}

			err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "package_name"}, {Name: "product_id"}, {Name: "base_plan_id"}, {Name: "offer_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"state"}), // ✅ Only update the state
			}).CreateInBatches(offers[i:end], batchSize).Error

			if err != nil {
				return fmt.Errorf("failed to upsert offers in batch: %w", err)
			}
		}
		return nil
	})
}

// ✅ Delete All Subscription Offers (Safe and Efficient)
func (r *subscriptionCatalogRepository) DeleteAllSubscriptionOffers(ctx context.Context, packageName string, productID string, basePlanID string) error {
	// ✅ Unscoped delete ensures soft deletes are considered (if enabled)
	return r.db.WithContext(ctx).Where(
		"package_name = ? AND product_id = ? AND base_plan_id = ?", packageName, productID, basePlanID).
		Unscoped().Delete(&models.SubscriptionOffer{}).Error
}

//// GET functions

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
