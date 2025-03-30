package service

import (
	"context"
	"fmt"
	"log"

	clientService "subsnotifpro-go/internal/playstore/client/service"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

// PlaystoreApiService defines the interface for interacting with the Google Play API.
type PlaystoreApiService interface {

	// VerifyServiceAccountAccess validates that the service account has access to the specified package.
	VerifyServiceAccountAccess(ctx context.Context, serviceAccountPath, packageName string) (bool, error)

	// GetUserSubscriptionPurchase fetches subscription purchase details using a purchase token.
	GetUserSubscriptionPurchase(ctx context.Context, purchaseToken, packageName string) (*androidpublisher.SubscriptionPurchaseV2, error)

	// ListSubscriptionProducts lists all available subscription products for a package.
	ListSubscriptionProducts(ctx context.Context, packageName string) (*androidpublisher.ListSubscriptionsResponse, error)

	// GetSubscriptionProductDetails fetches details of a specific subscription product.
	GetSubscriptionProductDetails(ctx context.Context, productID, packageName string) (*androidpublisher.Subscription, error)

	// GetSubscriptionOffers fetches offers for a specific base plan in a subscription product.
	GetSubscriptionOffers(ctx context.Context, packageName, productID, basePlanID string) ([]*androidpublisher.SubscriptionOffer, error)

	// AcknowledgeSubscription sends an acknowledgement request for a given subscription
	AcknowledgeSubscription(ctx context.Context, packageName, subscriptionID, purchaseToken string) error
}

// playstoreApiService implements the PlaystoreClientService interface.
type playstoreApiService struct {
	clientService clientService.PlaystoreClientService
}

// NewPlaystoreClientService creates a new instance of PlaystoreClientService.
func NewPlaystoreApiService(clientService clientService.PlaystoreClientService) PlaystoreApiService {
	return &playstoreApiService{clientService: clientService}
}

// -------------------------
// 🚀 VerifyServiceAccountAccess
// -------------------------
func (s *playstoreApiService) VerifyServiceAccountAccess(ctx context.Context, serviceAccountPath, packageName string) (bool, error) {
	// Initialize the API client with the provided service account
	service, err := androidpublisher.NewService(ctx, option.WithCredentialsFile(serviceAccountPath))
	if err != nil {
		return false, fmt.Errorf("failed to create Play Developer API client: %w", err)
	}

	// Test API access by listing subscriptions
	_, err = service.Monetization.Subscriptions.List(packageName).Do()
	if err != nil {
		// If permission denied, return false without error
		return false, nil
	}

	return true, nil
}

// -------------------------
// 🚀 GetUserSubscriptionPurchase
// -------------------------
func (s *playstoreApiService) GetUserSubscriptionPurchase(ctx context.Context, purchaseToken, packageName string) (*androidpublisher.SubscriptionPurchaseV2, error) {
	// Get the singleton API client
	service, err := s.clientService.GetPublisherService(ctx, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to get publisher service: %w", err)
	}

	// Fetch subscription details
	resp, err := service.Purchases.Subscriptionsv2.Get(packageName, purchaseToken).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription purchase details: %w", err)
	}

	log.Printf("✅ Subscription details fetched successfully for token: %s\n", purchaseToken)
	return resp, nil
}

// -------------------------
// 🚀 ListSubscriptionProducts
// -------------------------
func (s *playstoreApiService) ListSubscriptionProducts(ctx context.Context, packageName string) (*androidpublisher.ListSubscriptionsResponse, error) {
	// Get the singleton API client
	service, err := s.clientService.GetPublisherService(ctx, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to get publisher service: %w", err)
	}

	// Fetch subscription products
	resp, err := service.Monetization.Subscriptions.List(packageName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list subscription products: %w", err)
	}

	log.Printf("✅ Subscription products listed successfully for package: %s\n", packageName)
	return resp, nil
}

// -------------------------
// 🚀 GetSubscriptionProductDetails
// -------------------------
func (s *playstoreApiService) GetSubscriptionProductDetails(ctx context.Context, productID, packageName string) (*androidpublisher.Subscription, error) {
	// Get the singleton API client
	service, err := s.clientService.GetPublisherService(ctx, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to get publisher service: %w", err)
	}

	// Fetch subscription product details
	subscription, err := service.Monetization.Subscriptions.Get(packageName, productID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription product details: %w", err)
	}

	log.Printf("✅ Subscription product details fetched successfully for product: %s\n", productID)
	return subscription, nil
}

// -------------------------
// 🚀 GetSubscriptionOffers
// -------------------------
func (s *playstoreApiService) GetSubscriptionOffers(ctx context.Context, packageName, productID, basePlanID string) ([]*androidpublisher.SubscriptionOffer, error) {
	// Get the singleton API client
	service, err := s.clientService.GetPublisherService(ctx, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to get publisher service: %w", err)
	}

	// Fetch subscription offers
	resp, err := service.Monetization.Subscriptions.BasePlans.Offers.List(packageName, productID, basePlanID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscription offers: %w", err)
	}

	log.Printf("✅ Subscription offers fetched successfully for product: %s, base plan: %s\n", productID, basePlanID)
	return resp.SubscriptionOffers, nil
}

func (s *playstoreApiService) AcknowledgeSubscription(
	ctx context.Context,
	packageName string,
	subscriptionID string,
	purchaseToken string,
) error {
	service, err := s.clientService.GetPublisherService(ctx, packageName)
	if err != nil {
		return fmt.Errorf("failed to get publisher service: %w", err)
	}

	req := &androidpublisher.SubscriptionPurchasesAcknowledgeRequest{
		// Optional: add DeveloperPayload if needed
	}

	err = service.Purchases.Subscriptions.Acknowledge(packageName, subscriptionID, purchaseToken, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to acknowledge subscription: %w", err)
	}

	log.Printf("✅ Acknowledged subscription: %s for token: %s\n", subscriptionID, purchaseToken)
	return nil
}
