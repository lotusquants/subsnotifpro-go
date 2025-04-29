package service

import (
	"context"
	"fmt"
	"log"
	"subsnotifpro-go/internal/metrics"
	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/playstore/rtdn/models"
	"time"
)

// ProcessWebhookEvent routes Google Play webhook events based on event type
func (s *rtdnService) ProcessWebhookEvent(ctx context.Context, payload models.GooglePublishPayload) error {

	startTime := time.Now()

	var err error

	switch {
	case payload.Event.Subscription != nil:
		err = s.ProcessSubscriptionEvent(ctx, payload)
	case payload.Event.OneTimeProduct != nil:
		err = s.ProcessOneTimeProductEvent(ctx, payload)
	case payload.Event.VoidedPurchase != nil:
		err = s.ProcessVoidedPurchaseEvent(ctx, payload)
	case payload.Event.Test != nil:
		err = s.ProcessTestPurchaseEvent(ctx, payload)
	default:
		log.Println("⚠️ Unrecognized RTDN event type:", payload.ID)
		return nil
	}

	metrics.EventProcessingTime.WithLabelValues(payload.Event.PackageName).Observe(time.Since(startTime).Seconds())

	if err != nil {
		metrics.FailedEvents.WithLabelValues(payload.Event.PackageName).Inc()
		return err
	}

	metrics.ProcessedEvents.WithLabelValues(payload.Event.PackageName).Inc()
	return nil
}

// Process subscription events
func (s *rtdnService) ProcessSubscriptionEvent(ctx context.Context, payload models.GooglePublishPayload) error {

	logger.Log.Info("🔹 Processing Subscription Purchased Event")

	subscriptionNotification := payload.Event.Subscription
	if subscriptionNotification == nil {
		return fmt.Errorf("missing subscription notification data")
	}

	// 🔹 Extract Subscription Details
	purchaseToken := subscriptionNotification.PurchaseToken
	subscriptionId := subscriptionNotification.SubscriptionID
	packageName := payload.Event.PackageName

	if purchaseToken == "" || subscriptionId == "" || packageName == "" {
		return fmt.Errorf("invalid subscription event: missing purchaseToken or subscriptionId or packageName")
	}

	subscriptionData := payload.SubscriptionPurchase

	eventDto, err := payload.Event.ToDTO()
	if err != nil {
		return fmt.Errorf("failed to convert event model to dto")
	}

	err = s.playstoreSubscriptionService.UpsertSubscription(ctx, subscriptionData, eventDto)
	if err != nil {
		return fmt.Errorf("failed to process subscription: %w", err)
	}

	// Publish to websockets and notification service..

	// other things to do..

	logger.Log.Infof("✅ Subscription processed successfully for purchase token %s", purchaseToken)

	return nil

}

// Process one-time product events
func (s *rtdnService) ProcessOneTimeProductEvent(ctx context.Context, payload models.GooglePublishPayload) error {

	return nil
}

// Process voided purchase events
func (s *rtdnService) ProcessVoidedPurchaseEvent(ctx context.Context, payload models.GooglePublishPayload) error {

	return nil
}

// Process voided purchase events
func (s *rtdnService) ProcessTestPurchaseEvent(ctx context.Context, payload models.GooglePublishPayload) error {

	logger.Log.Infof("✅ Test Purchase processed successfully ")

	return nil
}

const (
	publishTimeout         = 5 * time.Second
	maxStatusUpdateRetries = 3
)
