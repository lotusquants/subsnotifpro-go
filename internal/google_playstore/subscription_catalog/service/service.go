package service

import (
	"context"
	"fmt"
	"log"

	clientService "subsnotifpro-go/internal/google_playstore/client/service"
	"subsnotifpro-go/internal/google_playstore/subscription_catalog/mapper"
	"subsnotifpro-go/internal/google_playstore/subscription_catalog/models"
	"subsnotifpro-go/internal/google_playstore/subscription_catalog/repository"
)

// SubscriptionCatalogService defines the interface for subscription syncing
type SubscriptionCatalogService interface {
	SyncSubscriptionCatalog(packageName string) error
	SyncSubscriptionOffers(packageName, productID, basePlanID string) error
}

// subscriptionCatalogService implements SubscriptionCatalogService
type subscriptionCatalogService struct {
	repo          repository.SubscriptionCatalogRepository
	clientService clientService.PlaystoreClientService
	ctx           context.Context
}

// ✅ Constructor for SubscriptionCatalogService
func NewSubscriptionCatalogService(
	ctx context.Context,
	repo repository.SubscriptionCatalogRepository,
	clientService clientService.PlaystoreClientService,
) SubscriptionCatalogService {
	return &subscriptionCatalogService{
		repo:          repo,
		clientService: clientService,
		ctx:           ctx,
	}
}

// -------------------------
// ✅ Sync Subscription Products from Google Play API
// -------------------------
func (s *subscriptionCatalogService) SyncSubscriptionCatalog(packageName string) error {
	log.Println("🔄 Starting Subscription Sync for package:", packageName)

	// ✅ Fetch subscription products
	log.Printf("🔍 Fetching subscription products for package: %s", packageName)
	response, err := s.clientService.ListSubscriptionProducts(packageName)
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
	offers, err := s.clientService.GetSubscriptionOffers(packageName, productID, basePlanID)
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
