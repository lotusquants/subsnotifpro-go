// internal/google_playstore/rtdn/service/service.go
package service

import (
	"log"
	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/google_playstore/models"
	"subsnotifpro-go/internal/google_playstore/rtdn/repository"
	"subsnotifpro-go/internal/metrics"
	"time"
)

// EventHandlerFunc defines a function signature for handling subscription events
type EventHandlerFunc func(event models.GooglePlayWebhookEvent) error

// eventHandlers maps subscription notification types to handler functions
var eventHandlers = map[int]EventHandlerFunc{
	constants.SUBSCRIPTION_PURCHASED:                 repository.SaveSubscription,
	constants.SUBSCRIPTION_RENEWED:                   repository.UpdateSubscriptionRenewal,
	constants.SUBSCRIPTION_CANCELED:                  repository.CancelSubscription,
	constants.SUBSCRIPTION_RECOVERED:                 repository.RecoverSubscription,
	constants.SUBSCRIPTION_ON_HOLD:                   repository.HandleSubscriptionOnHold,
	constants.SUBSCRIPTION_IN_GRACE_PERIOD:           repository.HandleSubscriptionInGracePeriod,
	constants.SUBSCRIPTION_RESTARTED:                 repository.HandleSubscriptionRestart,
	constants.SUBSCRIPTION_PRICE_CHANGE_CONFIRMED:    repository.HandlePriceChangeConfirmation,
	constants.SUBSCRIPTION_DEFERRED:                  repository.HandleSubscriptionDeferred,
	constants.SUBSCRIPTION_PAUSED:                    repository.HandleSubscriptionPaused,
	constants.SUBSCRIPTION_PAUSE_SCHEDULE_CHANGED:    repository.HandlePauseScheduleChanged,
	constants.SUBSCRIPTION_REVOKED:                   repository.HandleSubscriptionRevoked,
	constants.SUBSCRIPTION_EXPIRED:                   repository.HandleSubscriptionExpired,
	constants.SUBSCRIPTION_PENDING_PURCHASE_CANCELED: repository.HandlePendingPurchaseCanceled,
}

// One-Time Product Event Handlers
var oneTimeProductHandlers = map[int]EventHandlerFunc{
	constants.ONE_TIME_PRODUCT_PURCHASED: repository.SaveOneTimePurchase,
	constants.ONE_TIME_PRODUCT_CANCELED:  repository.HandleOneTimePurchaseCanceled,
}

// Voided Purchase Event Handlers
var voidedPurchaseHandlers = map[int]EventHandlerFunc{
	constants.PRODUCT_TYPE_SUBSCRIPTION: repository.HandleVoidedSubscription,
	constants.PRODUCT_TYPE_ONE_TIME:     repository.HandleVoidedOneTimePurchase,
}

// SaveWebhookEvent processes and stores Google Play RTDN webhook events
func SaveWebhookEvent(event *models.GooglePlayWebhookEvent) error {
	log.Println("📩 Storing Google Play webhook event:", event.ID)

	// Pass event to repository layer
	return repository.SaveWebhookEvent(event)
}

// ProcessWebhookEvent routes Google Play webhook events based on event type
func ProcessWebhookEvent(event models.GooglePlayWebhookEvent) error {
	startTime := time.Now() // Track processing time

	var err error
	if event.SubscriptionNotification != nil {
		err = handleSubscriptionEvent(event)
	} else if event.OneTimeProductNotification != nil {
		err = handleOneTimePurchaseEvent(event)
	} else if event.VoidedPurchaseNotification != nil {
		err = handleVoidedPurchaseEvent(event)
	} else if event.TestNotification != nil {
		log.Println("🟢 Test notification received:", event.ID)
		return nil
	} else {
		log.Println("⚠️ Unrecognized RTDN event type:", event.ID)
		return nil
	}

	// ✅ Record processing time
	metrics.EventProcessingTime.WithLabelValues(event.PackageName).Observe(time.Since(startTime).Seconds())

	if err != nil {
		// ✅ Increment Failed Events Counter
		metrics.FailedEvents.WithLabelValues(event.PackageName).Inc()
		return err
	}

	// ✅ Increment Processed Events Counter
	metrics.ProcessedEvents.WithLabelValues(event.PackageName).Inc()
	return nil
}

// handleSubscriptionEvent dynamically processes subscription events
func handleSubscriptionEvent(event models.GooglePlayWebhookEvent) error {
	notification := event.SubscriptionNotification
	eventType, exists := constants.SubscriptionNotificationTypes[notification.NotificationType]

	if !exists {
		log.Printf("⚠️ Unknown subscription event type: %d", notification.NotificationType)
		return nil
	}

	log.Printf("📢 Processing Subscription Event: %s for Subscription ID: %s", eventType, notification.SubscriptionID)

	// Get the handler function from the map
	if handler, found := eventHandlers[notification.NotificationType]; found {
		return handler(event)
	}

	log.Printf("⚠️ No handler defined for event type: %d", notification.NotificationType)
	return nil
}

// handleOneTimePurchaseEvent routes one-time purchase events
func handleOneTimePurchaseEvent(event models.GooglePlayWebhookEvent) error {
	notification := event.OneTimeProductNotification
	eventType, exists := constants.OneTimeProductNotificationTypes[notification.NotificationType]

	if !exists {
		log.Printf("⚠️ Unknown one-time purchase event type: %d", notification.NotificationType)
		return nil
	}

	log.Printf("📢 Processing One-Time Purchase Event: %s for SKU: %s", eventType, notification.Sku)

	// Route event to appropriate handler
	if handler, found := oneTimeProductHandlers[notification.NotificationType]; found {
		return handler(event)
	}

	log.Printf("⚠️ No handler defined for one-time product event type: %d", notification.NotificationType)
	return nil
}

// handleVoidedPurchaseEvent routes voided purchase events
func handleVoidedPurchaseEvent(event models.GooglePlayWebhookEvent) error {
	notification := event.VoidedPurchaseNotification
	eventType, exists := constants.VoidedPurchaseNotificationTypes[notification.ProductType]

	if !exists {
		log.Printf("⚠️ Unknown voided purchase event type: %d", notification.ProductType)
		return nil
	}

	log.Printf("🚫 Processing Voided Purchase Event: %s for Order ID: %s", eventType, notification.OrderID)

	// Route event to appropriate handler
	if handler, found := voidedPurchaseHandlers[notification.ProductType]; found {
		return handler(event)
	}

	log.Printf("⚠️ No handler defined for voided purchase event type: %d", notification.ProductType)
	return nil
}
