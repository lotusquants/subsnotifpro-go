// internal/google_playstore/rtdn/service/service.go
package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/metrics"
	apiService "subsnotifpro-go/internal/playstore/api/service"
	"subsnotifpro-go/internal/playstore/rtdn/models"
	"subsnotifpro-go/internal/playstore/rtdn/repository"
	playstoreSubscriptionService "subsnotifpro-go/internal/playstore/subscription/service"

	"gorm.io/gorm"
)

// RTDNService defines an interface for RTDN service methods
type RTDNService interface {
	SaveWebhookEvent(event *models.GooglePlayWebhookEvent) error
	ProcessWebhookEvent(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionEvent(event models.GooglePlayWebhookEvent) error
	ProcessOneTimeProductEvent(event models.GooglePlayWebhookEvent) error
	ProcessVoidedPurchaseEvent(event models.GooglePlayWebhookEvent) error

	// Individual event handlers for different notification types
	ProcessSubscriptionPurchased(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionRenewed(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionCanceled(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionRecovered(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionOnHold(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionInGracePeriod(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionRestarted(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionPriceChangeConfirmed(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionDeferred(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionPaused(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionPauseScheduleChanged(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionRevoked(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionExpired(event models.GooglePlayWebhookEvent) error
	ProcessSubscriptionPendingPurchaseCanceled(event models.GooglePlayWebhookEvent) error
	ProcessOneTimeProductPurchased(event models.GooglePlayWebhookEvent) error
	ProcessOneTimeProductCanceled(event models.GooglePlayWebhookEvent) error
	ProcessVoidedSubscription(event models.GooglePlayWebhookEvent) error
	ProcessVoidedOneTimePurchase(event models.GooglePlayWebhookEvent) error
}

// rtdnService implements the RTDNService interface
type rtdnService struct {
	repo                         repository.RTDNRepository
	ctx                          context.Context
	apiService                   apiService.PlaystoreApiService
	playstoreSubscriptionService playstoreSubscriptionService.PlaystoreSubscriptionService
	db                           *gorm.DB

	// Handler maps
	subscriptionHandlers   map[models.SubscriptionNotificationType]func(models.GooglePlayWebhookEvent) error
	oneTimeProductHandlers map[models.OneTimeProductNotificationType]func(models.GooglePlayWebhookEvent) error
	voidedPurchaseHandlers map[models.VoidedPurchaseProductType]func(models.GooglePlayWebhookEvent) error
}

// NewRTDNService creates a new instance of RTDNService
func NewRTDNService(ctx context.Context,
	repo repository.RTDNRepository,
	apiService apiService.PlaystoreApiService,
	playstoreSubscriptionService playstoreSubscriptionService.PlaystoreSubscriptionService,
	db *gorm.DB,
) RTDNService {
	service := &rtdnService{repo: repo,
		ctx:                          ctx,
		apiService:                   apiService,
		playstoreSubscriptionService: playstoreSubscriptionService,
		db:                           db,
	}

	// Initialize handler maps with instance methods
	service.subscriptionHandlers = map[models.SubscriptionNotificationType]func(models.GooglePlayWebhookEvent) error{
		models.SubscriptionPurchased:               service.ProcessSubscriptionPurchased,
		models.SubscriptionRenewed:                 service.ProcessSubscriptionRenewed,
		models.SubscriptionCanceled:                service.ProcessSubscriptionCanceled,
		models.SubscriptionRecovered:               service.ProcessSubscriptionRecovered,
		models.SubscriptionOnHold:                  service.ProcessSubscriptionOnHold,
		models.SubscriptionInGracePeriod:           service.ProcessSubscriptionInGracePeriod,
		models.SubscriptionRestarted:               service.ProcessSubscriptionRestarted,
		models.SubscriptionPriceChangeConfirmed:    service.ProcessSubscriptionPriceChangeConfirmed,
		models.SubscriptionDeferred:                service.ProcessSubscriptionDeferred,
		models.SubscriptionPaused:                  service.ProcessSubscriptionPaused,
		models.SubscriptionPauseScheduleChanged:    service.ProcessSubscriptionPauseScheduleChanged,
		models.SubscriptionRevoked:                 service.ProcessSubscriptionRevoked,
		models.SubscriptionExpired:                 service.ProcessSubscriptionExpired,
		models.SubscriptionPendingPurchaseCanceled: service.ProcessSubscriptionPendingPurchaseCanceled,
	}

	service.oneTimeProductHandlers = map[models.OneTimeProductNotificationType]func(models.GooglePlayWebhookEvent) error{
		models.OneTimeProductPurchased: service.ProcessOneTimeProductPurchased,
		models.OneTimeProductCanceled:  service.ProcessOneTimeProductCanceled,
	}

	service.voidedPurchaseHandlers = map[models.VoidedPurchaseProductType]func(models.GooglePlayWebhookEvent) error{
		models.ProductTypeSubscription: service.ProcessVoidedSubscription,
		models.ProductTypeOneTime:      service.ProcessVoidedOneTimePurchase,
	}

	return service
}

// SaveWebhookEvent stores the Google Play RTDN webhook event
func (s *rtdnService) SaveWebhookEvent(event *models.GooglePlayWebhookEvent) error {
	log.Println("📩 Storing Google Play webhook event:", event.ID)
	return s.repo.SaveWebhookEvent(s.ctx, s.db, event)
}

// ProcessWebhookEvent routes Google Play webhook events based on event type
func (s *rtdnService) ProcessWebhookEvent(event models.GooglePlayWebhookEvent) error {
	startTime := time.Now()
	var err error

	switch {
	case event.SubscriptionNotification != nil:
		err = s.ProcessSubscriptionEvent(event)
	case event.OneTimeProductNotification != nil:
		err = s.ProcessOneTimeProductEvent(event)
	case event.VoidedPurchaseNotification != nil:
		err = s.ProcessVoidedPurchaseEvent(event)
	default:
		log.Println("⚠️ Unrecognized RTDN event type:", event.ID)
		return nil
	}

	metrics.EventProcessingTime.WithLabelValues(event.PackageName).Observe(time.Since(startTime).Seconds())

	if err != nil {
		metrics.FailedEvents.WithLabelValues(event.PackageName).Inc()
		return err
	}

	metrics.ProcessedEvents.WithLabelValues(event.PackageName).Inc()
	return nil
}

// Process subscription events
func (s *rtdnService) ProcessSubscriptionEvent(event models.GooglePlayWebhookEvent) error {
	if handler, found := s.subscriptionHandlers[event.SubscriptionNotification.NotificationType]; found {
		return handler(event)
	}

	log.Println("⚠️ No handler for subscription event type:", event.SubscriptionNotification.NotificationType)
	return nil
}

// Process one-time product events
func (s *rtdnService) ProcessOneTimeProductEvent(event models.GooglePlayWebhookEvent) error {
	if handler, found := s.oneTimeProductHandlers[event.OneTimeProductNotification.NotificationType]; found {
		return handler(event)
	}
	log.Println("⚠️ No handler for one-time product event type:", event.OneTimeProductNotification.NotificationType)
	return nil
}

// Process voided purchase events
func (s *rtdnService) ProcessVoidedPurchaseEvent(event models.GooglePlayWebhookEvent) error {
	if handler, found := s.voidedPurchaseHandlers[event.VoidedPurchaseNotification.ProductType]; found {
		return handler(event)
	}
	log.Println("⚠️ No handler for voided purchase event type:", event.VoidedPurchaseNotification.ProductType)
	return nil
}

func (s *rtdnService) ProcessSubscriptionPurchased(event models.GooglePlayWebhookEvent) error {
	logger.Log.Info("🔹 Processing Subscription Purchased Event")

	subscriptionNotification := event.SubscriptionNotification
	if subscriptionNotification == nil {
		return fmt.Errorf("missing subscription notification data")
	}

	// 🔹 Extract Subscription Details
	purchaseToken := subscriptionNotification.PurchaseToken
	subscriptionId := subscriptionNotification.SubscriptionID
	packageName := event.PackageName

	if purchaseToken == "" || subscriptionId == "" {
		return fmt.Errorf("invalid subscription event: missing purchaseToken or subscriptionId")
	}

	// 🔹 Fetch Subscription Data from Google Play API
	subscriptionData, err := s.apiService.GetUserSubscriptionPurchase(s.ctx, purchaseToken, packageName)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription purchase data: %w", err)
	}

	// delegate the call subscription service...
	err = s.playstoreSubscriptionService.UpsertSubscription(s.ctx, subscriptionData, event)
	if err != nil {
		return fmt.Errorf("failed to process subscription: %w", err)
	}

	logger.Log.Infof("✅ Subscription processed successfully for purchase token %s", subscriptionData.LinkedPurchaseToken)
	return nil

}

func (s *rtdnService) ProcessSubscriptionRenewed(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Subscription Renewed Event")
	return nil
}

func (s *rtdnService) ProcessSubscriptionCanceled(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Subscription Canceled Event")
	return nil
}

func (s *rtdnService) ProcessSubscriptionRecovered(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Subscription Recovered Event")
	return nil
}

func (s *rtdnService) ProcessSubscriptionOnHold(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Subscription On Hold Event")
	return nil
}

func (s *rtdnService) ProcessSubscriptionInGracePeriod(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Subscription In Grace Period Event")
	return nil
}

func (s *rtdnService) ProcessSubscriptionRestarted(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Subscription Restarted Event")
	return nil
}

func (s *rtdnService) ProcessSubscriptionPriceChangeConfirmed(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Subscription Price Change Confirmed Event")
	return nil
}

func (s *rtdnService) ProcessSubscriptionDeferred(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Subscription Deferred Event")
	return nil
}

func (s *rtdnService) ProcessSubscriptionPaused(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Subscription Paused Event")
	return nil
}

func (s *rtdnService) ProcessSubscriptionPauseScheduleChanged(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Subscription Pause Schedule Changed Event")
	return nil
}

func (s *rtdnService) ProcessSubscriptionRevoked(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Subscription Revoked Event")
	return nil
}

func (s *rtdnService) ProcessSubscriptionExpired(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Subscription Expired Event")
	return nil
}

func (s *rtdnService) ProcessSubscriptionPendingPurchaseCanceled(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Subscription Pending Purchase Canceled Event")
	return nil
}

func (s *rtdnService) ProcessOneTimeProductPurchased(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing One-Time Product Purchased Event")
	return nil
}

func (s *rtdnService) ProcessOneTimeProductCanceled(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing One-Time Product Canceled Event")
	return nil
}

func (s *rtdnService) ProcessVoidedSubscription(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Voided Subscription Event")
	return nil
}

func (s *rtdnService) ProcessVoidedOneTimePurchase(event models.GooglePlayWebhookEvent) error {
	log.Println("🔹 Processing Voided One-Time Purchase Event")
	return nil
}
