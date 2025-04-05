package service

import (
	"context"
	"fmt"
	"log"
	"time"

	apiService "subsnotifpro-go/internal/playstore/api/service"
	"subsnotifpro-go/internal/playstore/products/mapper"
	"subsnotifpro-go/internal/playstore/products/models"
	"subsnotifpro-go/internal/playstore/products/repository"
	"subsnotifpro-go/internal/playstore/products/utils"
)

// SubscriptionCatalogService defines the interface for subscription syncing
type SubscriptionCatalogService interface {
	SyncSubscriptionCatalog(packageName string) error
	SyncSubscriptionOffers(packageName, productID, basePlanID string) error

	ListSubscriptionProducts() ([]map[string]string, error)
	GetSubscriptionProduct(packageName, productID string) (*models.ProductSubscription, error)
	CheckSubscriptionProductExists(packageName, productID string) (bool, error)

	GetBasePlanDetails(packageName, productID, basePlanID string) (*models.ProductBasePlan, error)
	IsBasePlanActive(packageName, productID, basePlanID string) (bool, error)
	IsBasePlanAvailableInRegion(packageName, productID, basePlanID, regionCode string) (bool, error)
	ListBasePlanNames(packageName, productID string) ([]string, error)
	GetRegionalBasePlanPrice(packageName, productID, basePlanID, regionCode string) (*models.Money, error)
	GetOtherRegionsBasePlanPrice(packageName, productID, basePlanID, currency string) (*models.Money, error)
	GetSubscriptionOffer(packageName, productID, basePlanID, offerID string) (*models.SubscriptionOffer, error)
	ListOfferNamesForBasePlan(packageName, productID, basePlanID string) ([]string, error)
	IsSubscriptionOfferActive(packageName, productID, basePlanID, offerID string) (bool, error)

	GetOfferPhases(packageName, productID, basePlanID, offerID string) ([]models.SubscriptionOfferPhase, error)
	IsOfferPhaseExists(packageName, productID, basePlanID, offerID string) (bool, error)
	GetCurrentOfferPhaseIndex(packageName, productID, basePlanID, offerID string, startTime, checkTime time.Time) (*int, error)
	GetRegionalOfferPhasePrice(packageName, productID, basePlanID, offerID string, phaseIndex int, regionCode string) (*models.Money, error)
}

// subscriptionCatalogService implements SubscriptionCatalogService
type subscriptionCatalogService struct {
	repo       repository.SubscriptionCatalogRepository
	apiService apiService.PlaystoreApiService
	ctx        context.Context
}

// ✅ Constructor for SubscriptionCatalogService
func NewSubscriptionCatalogService(
	ctx context.Context,
	repo repository.SubscriptionCatalogRepository,
	apiService apiService.PlaystoreApiService,
) SubscriptionCatalogService {
	return &subscriptionCatalogService{
		repo:       repo,
		apiService: apiService,
		ctx:        ctx,
	}
}

// -------------------------
// ✅ Sync Subscription Products from Google Play API
// -------------------------
func (s *subscriptionCatalogService) SyncSubscriptionCatalog(packageName string) error {
	log.Println("🔄 Starting Subscription Sync for package:", packageName)

	// ✅ Fetch subscription products
	log.Printf("🔍 Fetching subscription products for package: %s", packageName)
	response, err := s.apiService.ListSubscriptionProducts(s.ctx, packageName)
	if err != nil {
		log.Printf("❌ API Error: Failed to fetch subscription products for package %s - %v", packageName, err)
		return fmt.Errorf("failed to fetch subscription products: %w", err)
	}

	if response == nil || len(response.Subscriptions) == 0 {
		log.Printf("⚠️ No subscription products found for package: %s", packageName)
		return nil
	}

	log.Printf("📦 Received %d subscriptions from API for package: %s", len(response.Subscriptions), packageName)

	// ✅ Step 1: Delete existing subscriptions
	log.Printf("🗑️ Deleting existing subscription products for package: %s", packageName)
	if err := s.repo.DeleteAllSubscriptionProducts(s.ctx); err != nil {
		log.Printf("❌ Error deleting subscriptions: %v", err)
		return fmt.Errorf("failed to delete subscriptions: %w", err)
	}

	// ✅ Step 2: Insert new subscriptions
	for _, sub := range response.Subscriptions {
		subscription := mapper.ConvertSubscriptionModel(sub, packageName)

		if err := s.repo.UpsertSubscriptionProducts(s.ctx, subscription); err != nil {
			log.Printf("⚠️ Skipping subscription %s due to error: %v", sub.ProductId, err)
			continue // ✅ Continue processing other subscriptions
		}

		// ✅ Process Base Plans
		for _, basePlan := range sub.BasePlans {
			basePlanModel := mapper.ConvertBasePlanModel(basePlan, packageName, sub.ProductId)

			if err := s.repo.UpsertSubscriptionBasePlans(s.ctx, basePlanModel); err != nil {
				log.Printf("⚠️ Skipping base plan %s due to error: %v", basePlan.BasePlanId, err)
				continue
			}

			// ✅ Sync Offers for each Base Plan
			if err := s.SyncSubscriptionOffers(packageName, sub.ProductId, basePlan.BasePlanId); err != nil {
				log.Printf("⚠️ Offer sync failed for product %s, base plan %s: %v", sub.ProductId, basePlan.BasePlanId, err)
			}
		}
	}

	log.Println("✅ Subscription sync completed successfully for package:", packageName)
	return nil
}

// -------------------------
// ✅ Sync Subscription Offers from Google Play API
// -------------------------
func (s *subscriptionCatalogService) SyncSubscriptionOffers(packageName, productID, basePlanID string) error {
	log.Printf("🔄 Full Sync: Deleting and re-fetching Offers for Package: %s, Product: %s, BasePlan: %s", packageName, productID, basePlanID)

	// ✅ Step 1: Delete all existing offers for this base plan
	log.Printf("🗑️ Deleting existing offers for product: %s, base plan: %s", productID, basePlanID)
	if err := s.repo.DeleteAllSubscriptionOffers(s.ctx, packageName, productID, basePlanID); err != nil {
		return fmt.Errorf("❌ Failed to delete existing offers: %w", err)
	}

	// ✅ Step 2: Fetch subscription offers from Google Play API
	log.Printf("🔍 Fetching subscription offers for product: %s, base plan: %s", productID, basePlanID)
	offers, err := s.apiService.GetSubscriptionOffers(s.ctx, packageName, productID, basePlanID)
	if err != nil {
		log.Printf("❌ API Error: Failed to fetch subscription offers for product: %s, base plan: %s - %v", productID, basePlanID, err)
		return fmt.Errorf("failed to fetch subscription offers: %w", err)
	}

	// ✅ Ensure API Response is Valid
	if len(offers) == 0 {
		log.Printf("⚠️ No offers found for package: %s, product: %s, basePlan: %s", packageName, productID, basePlanID)
		return nil
	}

	log.Printf("📦 Received %d offers for Product %s, BasePlan %s", len(offers), productID, basePlanID)

	// ✅ Step 3: Convert API response to internal models
	var offerModels []models.SubscriptionOffer
	for _, offer := range offers {
		offerModel := mapper.ConvertSubscriptionOfferModel(offer, packageName, productID, basePlanID)
		offerModels = append(offerModels, offerModel)
	}

	// ✅ Step 4: Insert new offers with better error handling
	if err := s.repo.UpsertSubscriptionOffers(s.ctx, offerModels); err != nil {
		log.Printf("❌ Database Error: Failed to upsert offers for product %s, base plan %s: %v", productID, basePlanID, err)
		return fmt.Errorf("failed to upsert offers: %w", err)
	}

	log.Printf("✅ Successfully synced offers for Product %s, BasePlan %s", productID, basePlanID)
	return nil
}

// ✅ Get Full Subscription Product (Delegates to Repository)
func (s *subscriptionCatalogService) GetSubscriptionProduct(packageName, productID string) (*models.ProductSubscription, error) {
	product, err := s.repo.GetSubscriptionProduct(s.ctx, packageName, productID)
	if err != nil {
		return nil, err
	}
	return product, nil
}

// ✅ Check if Subscription Product Exists (Delegates to Repository)
func (s *subscriptionCatalogService) CheckSubscriptionProductExists(packageName, productID string) (bool, error) {
	isActive, err := s.repo.CheckSubscriptionProductExists(s.ctx, packageName, productID)
	if err != nil {
		return false, err
	}
	return isActive, nil
}

// ✅ List Subscription Products Service (Ensures Business Logic Separation)
func (s *subscriptionCatalogService) ListSubscriptionProducts() ([]map[string]string, error) {
	products, err := s.repo.ListSubscriptionProducts(s.ctx)
	if err != nil {
		return nil, err
	}

	// ✅ Transform to Simple JSON Response Format
	response := make([]map[string]string, len(products))
	for i, product := range products {
		response[i] = map[string]string{
			"packageName": product.PackageName,
			"productId":   product.ProductID,
		}
	}

	return response, nil
}

// ✅ Service: Get Full Base Plan Details
func (s *subscriptionCatalogService) GetBasePlanDetails(packageName, productID, basePlanID string) (*models.ProductBasePlan, error) {
	return s.repo.GetBasePlanDetails(s.ctx, packageName, productID, basePlanID)
}

// ✅ Service: Check if Base Plan is Active
func (s *subscriptionCatalogService) IsBasePlanActive(packageName, productID, basePlanID string) (bool, error) {
	return s.repo.IsBasePlanActive(s.ctx, packageName, productID, basePlanID)
}

// ✅ Check If Base Plan is Available in a Region
func (s *subscriptionCatalogService) IsBasePlanAvailableInRegion(
	packageName, productID, basePlanID, regionCode string,
) (bool, error) {
	return s.repo.IsBasePlanAvailableInRegion(s.ctx, packageName, productID, basePlanID, regionCode)
}

// ✅ Service Function to Get Base Plan Names
func (s *subscriptionCatalogService) ListBasePlanNames(packageName, productID string) ([]string, error) {
	basePlans, err := s.repo.ListBasePlanNames(s.ctx, packageName, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to list base plan names: %w", err)
	}
	return basePlans, nil
}

// ✅ Get Base Plan Price (Strict Regional Check - No Fallback)
func (s *subscriptionCatalogService) GetRegionalBasePlanPrice(
	packageName, productID, basePlanID, regionCode string,
) (*models.Money, error) {
	// ✅ Fetch price from regional config only
	price, err := s.repo.GetRegionalBasePlanPrice(s.ctx, packageName, productID, basePlanID, regionCode)
	if err != nil {
		return nil, err // Return error if price is not found
	}

	return price, nil
}

// ✅ Get Other Regions Base Plan Price with Currency Selection
func (s *subscriptionCatalogService) GetOtherRegionsBasePlanPrice(
	packageName, productID, basePlanID, currency string,
) (*models.Money, error) {
	// ✅ Fetch Price from Repository
	price, err := s.repo.GetOtherRegionsBasePlanPrice(s.ctx, packageName, productID, basePlanID, currency)
	if err != nil {
		return nil, err // Pass the error up if there is an issue
	}

	// ✅ If no valid price is found, return a descriptive error
	if price == nil {
		return nil, fmt.Errorf("no valid price found for base plan %s in other regions for currency: %s", basePlanID, currency)
	}

	return price, nil
}

// ✅ Get Subscription Offer with all Nested Fields
func (s *subscriptionCatalogService) GetSubscriptionOffer(packageName, productID, basePlanID, offerID string) (*models.SubscriptionOffer, error) {
	// Fetch offer from the repository
	offer, err := s.repo.GetSubscriptionOffer(s.ctx, packageName, productID, basePlanID, offerID)
	if err != nil {
		return nil, err
	}

	// ✅ Ensure the offer has at least one phase
	if len(offer.Phases) == 0 {
		return nil, fmt.Errorf("offer has no valid phases")
	}

	return offer, nil
}

// ✅ Check if Offer is Active via Service
func (s *subscriptionCatalogService) IsSubscriptionOfferActive(packageName, productID, basePlanID, offerID string) (bool, error) {
	isActive, err := s.repo.IsSubscriptionOfferActive(s.ctx, packageName, productID, basePlanID, offerID)
	if err != nil {
		return false, fmt.Errorf("failed to check if offer is active: %w", err)
	}
	return isActive, nil
}

// ✅ Service Function to Get Offer Names for a Base Plan
func (s *subscriptionCatalogService) ListOfferNamesForBasePlan(packageName, productID, basePlanID string) ([]string, error) {
	offerNames, err := s.repo.ListOfferNamesForBasePlan(s.ctx, packageName, productID, basePlanID)
	if err != nil {
		return nil, fmt.Errorf("failed to list offer names: %w", err)
	}
	return offerNames, nil
}

// ✅ Get Offer Phases
func (s *subscriptionCatalogService) GetOfferPhases(packageName, productID, basePlanID, offerID string) ([]models.SubscriptionOfferPhase, error) {
	phases, err := s.repo.GetSubscriptionOfferPhases(s.ctx, packageName, productID, basePlanID, offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get offer phases: %w", err)
	}
	return phases, nil
}

// ✅ Check if an Offer Phase Exists
func (s *subscriptionCatalogService) IsOfferPhaseExists(packageName, productID, basePlanID, offerID string) (bool, error) {
	exists, err := s.repo.IsSubscriptionOfferPhaseExists(s.ctx, packageName, productID, basePlanID, offerID)
	if err != nil {
		return false, fmt.Errorf("failed to check offer phase existence: %w", err)
	}
	return exists, nil
}

// ✅ Get Current Offer Phase Index
func (s *subscriptionCatalogService) GetCurrentOfferPhaseIndex(
	packageName, productID, basePlanID, offerID string, startTime, checkTime time.Time,
) (*int, error) {
	// 🔹 Fetch Offer Phases from Repository
	phases, err := s.repo.GetSubscriptionOfferPhases(s.ctx, packageName, productID, basePlanID, offerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription offer phases: %w", err)
	}

	// 🔹 Validate if Phases Exist
	if len(phases) == 0 {
		return nil, fmt.Errorf("no subscription offer phases found for offer: %s", offerID)
	}

	// 🔹 Determine the Current Phase
	phaseStartTime := startTime

	for _, phase := range phases {
		// ✅ Validate RecurrenceCount
		if phase.RecurrenceCount < 0 {
			return nil, fmt.Errorf("invalid recurrence count (%d) for offer %s's phase %d", phase.RecurrenceCount, offerID, phase.PhaseIndex)
		}

		// ✅ Skip if RecurrenceCount is 0 (invalid phase)
		if phase.RecurrenceCount == 0 {
			continue
		}

		// ✅ Parse Phase Duration
		phaseDuration, parseErr := utils.ParseISO8601Duration(phase.Duration, phaseStartTime)
		if parseErr != nil {
			return nil, fmt.Errorf("error parsing phase duration for offer %s: %w", offerID, parseErr)
		}

		// ✅ Compute Total Duration for the Phase
		totalPhaseDuration := phaseDuration * time.Duration(phase.RecurrenceCount)

		// ✅ Compute Phase End Time
		phaseEndTime := phaseStartTime.Add(totalPhaseDuration)

		// ✅ Check if the Given Time Falls Within This Phase
		if checkTime.After(phaseStartTime) && checkTime.Before(phaseEndTime) {
			return &phase.PhaseIndex, nil // ✅ Return Phase Index Instead of Full Phase
		}

		// ✅ Move to the Next Phase
		phaseStartTime = phaseEndTime
	}

	// ❌ If No Valid Phase Exists, Return nil...no valid phase exists, use base plan price
	return nil, nil
}

// ✅ GetRegionalOfferPhasePrice - Computes Offer Phase Price for a Specific Region
func (s *subscriptionCatalogService) GetRegionalOfferPhasePrice(
	packageName, productID, basePlanID, offerID string, phaseIndex int, regionCode string,
) (*models.Money, error) {

	// 🔹 Fetch Regional Offer Phase Config
	offerPhaseConfig, err := s.repo.GetRegionalOfferPhaseConfig(s.ctx, packageName, productID, basePlanID, offerID, phaseIndex, regionCode)
	if err != nil {
		return nil, err // 🔴 Return error if no data found
	}

	// ✅ If Offer is Free, Return 0.0 in Money Format (Google Play Standard)
	if offerPhaseConfig.Free {
		return &models.Money{CurrencyCode: "XXX", Units: 0, Nanos: 0}, nil
	}

	// ✅ If Direct Price Exists, Return It Immediately
	if offerPhaseConfig.Price != nil {
		return offerPhaseConfig.Price, nil
	}

	// ✅ If No Discount Exists, Return Error (No Valid Price Found)
	if offerPhaseConfig.RelativeDiscount == nil && offerPhaseConfig.AbsoluteDiscount == nil {
		return nil, fmt.Errorf("no valid price or discount configured for region: %s", regionCode)
	}

	// 🔹 Fetch Base Plan Price Only If Discount Is Applied
	basePlanPrice, err := s.repo.GetRegionalBasePlanPrice(s.ctx, packageName, productID, basePlanID, regionCode)
	if err != nil {
		return nil, fmt.Errorf("error fetching base plan price: %w", err)
	}
	if basePlanPrice == nil {
		return nil, fmt.Errorf("no base plan price found for region: %s", regionCode)
	}

	// 🔹 Initialize Final Price with Base Plan Price
	finalPrice := *basePlanPrice

	// 🔹 Apply Relative Discount (Percentage-Based Reduction)
	if offerPhaseConfig.RelativeDiscount != nil {
		discountFraction := *offerPhaseConfig.RelativeDiscount

		// ✅ Validate Discount Fraction (Strictly Between 0 and 1)
		if discountFraction <= 0 || discountFraction >= 1 {
			return nil, fmt.Errorf("invalid relative discount: must be > 0 and < 1, got %f", discountFraction)
		}

		// ✅ Compute Discounted Price (Avoids Floating-Point Precision Issues)
		totalPriceFloat := float64(finalPrice.Units) + float64(finalPrice.Nanos)/1e9
		discountAmountFloat := totalPriceFloat * discountFraction

		// ✅ Convert Discount Amount to Money Format
		discountAmount := models.Money{
			CurrencyCode: finalPrice.CurrencyCode,
			Units:        int64(discountAmountFloat),
			Nanos:        int64((discountAmountFloat - float64(int64(discountAmountFloat))) * 1e9),
		}

		// ✅ Apply Discount Using Utility Function
		finalPrice = utils.SubtractMoney(finalPrice, discountAmount)
	}

	// 🔹 Apply Absolute Discount (Fixed Amount)
	if offerPhaseConfig.AbsoluteDiscount != nil {
		finalPrice = utils.SubtractMoney(finalPrice, *offerPhaseConfig.AbsoluteDiscount)
	}

	// ✅ Ensure Price Is Not Negative (Google Play Minimum Price Check)
	if finalPrice.Units < 0 {
		finalPrice.Units = 0
		finalPrice.Nanos = 0
	}

	// ✅ Return Computed Price
	return &finalPrice, nil
}
