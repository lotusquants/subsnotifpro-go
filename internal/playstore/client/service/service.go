package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"
)

var (
	publisherService    *androidpublisher.Service // Singleton API client
	publisherServiceErr error                     // Error during API client initialization
	lastLoadedTime      time.Time                 // Last time the API client was loaded
	reloadMutex         sync.Mutex                // Mutex for thread-safe reloading
)

const (
	cacheDuration = 24 * time.Hour // Cache duration for the API client
)

type ServiceAccountProvider interface {
	GetServiceAccountPath(ctx context.Context, packageName string) (string, error)
}

// PlaystoreClientService defines the interface for interacting with the Google Play API.
type PlaystoreClientService interface {
	// GetPublisherService returns the singleton Android Publisher Service client.
	GetPublisherService(ctx context.Context, packageName string) (*androidpublisher.Service, error)
	SetAccountProvider(provider ServiceAccountProvider)
}

// playstoreCleintService implements the PlaystoreClientService interface.
type playstoreClientService struct {
	accountProvider ServiceAccountProvider
}

func NewPlaystoreClientService() *playstoreClientService {
	return &playstoreClientService{}
}

func (s *playstoreClientService) SetAccountProvider(provider ServiceAccountProvider) {
	s.accountProvider = provider
} // -------------------------
// 🚀 GetPublisherService
// -------------------------

func (s *playstoreClientService) GetPublisherService(ctx context.Context, packageName string) (*androidpublisher.Service, error) {
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
	publisherService, publisherServiceErr = s.loadPublisherService(ctx, packageName)
	lastLoadedTime = time.Now()

	if publisherServiceErr != nil {
		log.Printf("❌ Reload failed: %v\n", publisherServiceErr)
	} else {
		log.Println("✅ Service successfully reloaded.")
	}

	return publisherService, publisherServiceErr
}

// **loadPublisherService loads AndroidPublisher service using the latest valid service account.**
func (s *playstoreClientService) loadPublisherService(ctx context.Context, packageName string) (*androidpublisher.Service, error) {
	// Step 1: Fetch the service account details for the given package name
	filePath, err := s.accountProvider.GetServiceAccountPath(ctx, packageName)
	if err != nil {
		log.Printf("❌ Failed to fetch latest service account for package %s: %v\n", packageName, err)
		return nil, fmt.Errorf("failed to load service account for package %s: %w", packageName, err)
	}

	log.Printf("📂 Using service account file for package %s: %s\n", packageName, filePath)

	service, err := androidpublisher.NewService(ctx, option.WithCredentialsFile(filePath))
	if err != nil {
		log.Printf("❌ Failed to create Google Play Publisher Service with file [%s]: %v\n", filePath, err)
		// ✅ Reset service on failure (Prevents caching a broken service)
		publisherService = nil
		return nil, fmt.Errorf("failed to create Android Publisher service: %w", err)
	}

	log.Printf("✅ Google Play Publisher Service successfully created using file: %s\n", filePath)
	return service, nil
}
