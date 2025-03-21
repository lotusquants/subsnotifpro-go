// internal/google_playstore/rtdn/service/service.go
package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/google_playstore/rtdn/models"
	"subsnotifpro-go/internal/google_playstore/rtdn/repository"
	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/metrics"

	clientService "subsnotifpro-go/internal/google_playstore/client/service"
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
	repo          repository.RTDNRepository
	ctx           context.Context
	clientService clientService.PlaystoreClientService

	// Handler maps
	subscriptionHandlers   map[int]func(models.GooglePlayWebhookEvent) error
	oneTimeProductHandlers map[int]func(models.GooglePlayWebhookEvent) error
	voidedPurchaseHandlers map[int]func(models.GooglePlayWebhookEvent) error
}

// NewRTDNService creates a new instance of RTDNService
func NewRTDNService(ctx context.Context, repo repository.RTDNRepository, clientService clientService.PlaystoreClientService) RTDNService {
	service := &rtdnService{repo: repo, ctx: ctx, clientService: clientService}

	// Initialize handler maps with instance methods
	service.subscriptionHandlers = map[int]func(models.GooglePlayWebhookEvent) error{
		constants.SUBSCRIPTION_PURCHASED:                 service.ProcessSubscriptionPurchased,
		constants.SUBSCRIPTION_RENEWED:                   service.ProcessSubscriptionRenewed,
		constants.SUBSCRIPTION_CANCELED:                  service.ProcessSubscriptionCanceled,
		constants.SUBSCRIPTION_RECOVERED:                 service.ProcessSubscriptionRecovered,
		constants.SUBSCRIPTION_ON_HOLD:                   service.ProcessSubscriptionOnHold,
		constants.SUBSCRIPTION_IN_GRACE_PERIOD:           service.ProcessSubscriptionInGracePeriod,
		constants.SUBSCRIPTION_RESTARTED:                 service.ProcessSubscriptionRestarted,
		constants.SUBSCRIPTION_PRICE_CHANGE_CONFIRMED:    service.ProcessSubscriptionPriceChangeConfirmed,
		constants.SUBSCRIPTION_DEFERRED:                  service.ProcessSubscriptionDeferred,
		constants.SUBSCRIPTION_PAUSED:                    service.ProcessSubscriptionPaused,
		constants.SUBSCRIPTION_PAUSE_SCHEDULE_CHANGED:    service.ProcessSubscriptionPauseScheduleChanged,
		constants.SUBSCRIPTION_REVOKED:                   service.ProcessSubscriptionRevoked,
		constants.SUBSCRIPTION_EXPIRED:                   service.ProcessSubscriptionExpired,
		constants.SUBSCRIPTION_PENDING_PURCHASE_CANCELED: service.ProcessSubscriptionPendingPurchaseCanceled,
	}

	service.oneTimeProductHandlers = map[int]func(models.GooglePlayWebhookEvent) error{
		constants.ONE_TIME_PRODUCT_PURCHASED: service.ProcessOneTimeProductPurchased,
		constants.ONE_TIME_PRODUCT_CANCELED:  service.ProcessOneTimeProductCanceled,
	}

	service.voidedPurchaseHandlers = map[int]func(models.GooglePlayWebhookEvent) error{
		constants.PRODUCT_TYPE_SUBSCRIPTION: service.ProcessVoidedSubscription,
		constants.PRODUCT_TYPE_ONE_TIME:     service.ProcessVoidedOneTimePurchase,
	}

	return service
}

// SaveWebhookEvent stores the Google Play RTDN webhook event
func (s *rtdnService) SaveWebhookEvent(event *models.GooglePlayWebhookEvent) error {
	log.Println("📩 Storing Google Play webhook event:", event.ID)
	return s.repo.SaveWebhookEvent(s.ctx, event)
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
	log.Println("🔹 Processing Subscription Purchased Event")
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
	subscriptionData, err := s.clientService.GetUserSubscriptionPurchase(purchaseToken, packageName)
	if err != nil {
		return fmt.Errorf("failed to fetch subscription purchase data: %w", err)
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
