package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"subsnotifpro-go/internal/google_playstore/settings/repository"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

var (
	publisherService    *androidpublisher.Service
	publisherServiceErr error
	lastLoadedTime      time.Time
	reloadMutex         sync.Mutex
)

const cacheDuration = 10 * time.Minute

type PlaystoreClientService interface {
	// Fetch subscription purchase details using a purchase token.
	GetUserSubscriptionPurchase(purchaseToken string, packageName string) (*androidpublisher.SubscriptionPurchaseV2, error)

	// Fetch all available subscriptions from Play Store.
	ListSubscriptionProducts(packageName string) (*androidpublisher.ListSubscriptionsResponse, error)

	// Fetch details of a specific subscription product from Play Store.
	GetSubscriptionProductDetails(productID string, packageName string) (*androidpublisher.Subscription, error)

	// Fetch offers for a specific base plan in a subscription product.
	GetSubscriptionOffers(packageName, productID, basePlanID string) ([]*androidpublisher.SubscriptionOffer, error)

	// GetPublisherService returns the singleton Android Publisher Service client.
	GetPublisherService() (*androidpublisher.Service, error)

	// LoadPublisherService loads AndroidPublisher service using the latest valid service account.
	LoadPublisherService() (*androidpublisher.Service, error)
}

// **playstoreClientService implements the PlaystoreSettingsService interface
type playstoreClientService struct {
	settingsRepo repository.PlaystoreSettingsRepository
	ctx          context.Context
	db           *gorm.DB
}

// **NewGooglePlayClientService creates a new instance of GooglePlayClientService.**
func NewGooglePlayClientService(ctx context.Context, settingsRepo repository.PlaystoreSettingsRepository, db *gorm.DB) PlaystoreClientService {
	return &playstoreClientService{settingsRepo: settingsRepo, ctx: ctx, db: db}
}

// **GetPublisherService returns the singleton Android Publisher Service client.**
func (s *playstoreClientService) GetPublisherService() (*androidpublisher.Service, error) {
	if time.Since(lastLoadedTime) < cacheDuration {
		return publisherService, publisherServiceErr
	}

	reloadMutex.Lock()
	defer reloadMutex.Unlock()

	// **Double-check cache after acquiring lock**
	if time.Since(lastLoadedTime) < cacheDuration {
		return publisherService, publisherServiceErr
	}

	log.Println("🔄 Reloading Google Play Publisher Service due to cache expiry.")
	publisherService, publisherServiceErr = s.LoadPublisherService()
	lastLoadedTime = time.Now()

	if publisherServiceErr != nil {
		log.Printf("❌ Reload failed: %v\n", publisherServiceErr)
	} else {
		log.Println("✅ Service successfully reloaded.")
	}

	return publisherService, publisherServiceErr
}

// **loadPublisherService loads AndroidPublisher service using the latest valid service account.**
func (s *playstoreClientService) LoadPublisherService() (*androidpublisher.Service, error) {
	serviceAccount, err := s.settingsRepo.GetLatestServiceAccount(s.db)
	if err != nil {
		log.Printf("❌ Failed to fetch latest service account: %v\n", err)
		return nil, fmt.Errorf("failed to load service account: %w", err)
	}

	if serviceAccount.FilePath == "" {
		log.Println("❌ No valid service account file path found.")
		return nil, fmt.Errorf("no valid service account file found")
	}

	log.Printf("📂 Using service account file: %s\n", serviceAccount.FilePath)

	service, err := androidpublisher.NewService(s.ctx, option.WithCredentialsFile(serviceAccount.FilePath))
	if err != nil {
		log.Printf("❌ Failed to create Google Play Publisher Service with file [%s]: %v\n", serviceAccount.FilePath, err)
		// ✅ Reset service on failure (Prevents caching a broken service)
		publisherService = nil
		return nil, fmt.Errorf("failed to create Android Publisher service: %w", err)
	}

	log.Printf("✅ Google Play Publisher Service successfully created using file: %s\n", serviceAccount.FilePath)
	return service, nil
}

// **GetUserSubscriptionPurchase fetches subscription purchase details using a purchase token.**
func (s *playstoreClientService) GetUserSubscriptionPurchase(purchaseToken string, packageName string) (*androidpublisher.SubscriptionPurchaseV2, error) {
	publisherService, err := s.GetPublisherService()
	if err != nil {
		return nil, fmt.Errorf("failed to get publisher service: %w", err)
	}

	log.Printf("📦 Fetching subscription details for package: %s, token: %s\n", packageName, purchaseToken)

	resp, err := publisherService.Purchases.Subscriptionsv2.Get(packageName, purchaseToken).Context(s.ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription purchase details: %w", err)
	}

	log.Printf("✅ Subscription details fetched successfully for token: %s\n", purchaseToken)
	return resp, nil
}

// ListSubscriptionProducts fetches all available subscription products from Play Store (raw response).
func (s *playstoreClientService) ListSubscriptionProducts(packageName string) (*androidpublisher.ListSubscriptionsResponse, error) {
	publisherService, err := s.GetPublisherService()
	if err != nil {
		return nil, fmt.Errorf("failed to get publisher service: %w", err)
	}

	// Fetch raw subscription product data
	resp, err := publisherService.Monetization.Subscriptions.List(packageName).Context(s.ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	// ✅ Return the entire raw response for debugging
	return resp, nil
}

// GetSubscriptionProductDetails fetches details of a specific subscription product.
func (s *playstoreClientService) GetSubscriptionProductDetails(productID string, packageName string) (*androidpublisher.Subscription, error) {
	publisherService, err := s.GetPublisherService()
	if err != nil {
		return nil, fmt.Errorf("failed to get publisher service: %w", err)
	}

	subscription, err := publisherService.Monetization.Subscriptions.Get(packageName, productID).Context(s.ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription %s: %w", productID, err)
	}

	// ✅ Return raw API response, avoid mapping here
	return subscription, nil
}

// GetSubscriptionOffers fetches offers for a specific base plan in a subscription product.
func (s *playstoreClientService) GetSubscriptionOffers(packageName string, productID string, basePlanID string) ([]*androidpublisher.SubscriptionOffer, error) {
	publisherService, err := s.GetPublisherService()
	if err != nil {
		return nil, fmt.Errorf("failed to get publisher service: %w", err)
	}

	// Fetch subscription offers from Google Play API
	resp, err := publisherService.Monetization.Subscriptions.BasePlans.Offers.List(packageName, productID, basePlanID).Context(s.ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription offers for %s: %w", productID, err)
	}

	// ✅ Return the raw API response to maintain separation of concerns
	return resp.SubscriptionOffers, nil
}
