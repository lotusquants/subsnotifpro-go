// internal/google_playstore/rtdn/service/service.go
package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/metrics"
	apiDto "subsnotifpro-go/internal/playstore/api/dto"
	apiService "subsnotifpro-go/internal/playstore/api/service"
	"subsnotifpro-go/internal/playstore/events"
	"subsnotifpro-go/internal/playstore/rtdn/dto"
	"subsnotifpro-go/internal/playstore/rtdn/models"
	"subsnotifpro-go/internal/playstore/rtdn/repository"
	playstoreSubscriptionService "subsnotifpro-go/internal/playstore/subscription/service"

	"github.com/google/uuid"
	"github.com/sony/gobreaker"
	"gorm.io/gorm"
)

// RTDNService defines an interface for RTDN service methods
type RTDNService interface {
	ProcessWebhookEventForPublish(ctx context.Context, event *dto.GooglePlayWebhookEvent) error
	ProcessWebhookEvent(ctx context.Context, payload models.GooglePublishPayload) error

	ProcessSubscriptionEvent(ctx context.Context, payload models.GooglePublishPayload) error
	ProcessOneTimeProductEvent(ctx context.Context, payload models.GooglePublishPayload) error
	ProcessVoidedPurchaseEvent(ctx context.Context, payload models.GooglePublishPayload) error
}

// rtdnService implements the RTDNService interface

// Compile-time interface implementation check for event processor
var _ events.EventProcessor = (*rtdnService)(nil)

type rtdnService struct {
	repo                         repository.RTDNRepository
	ctx                          context.Context
	apiService                   apiService.PlaystoreApiService
	playstoreSubscriptionService playstoreSubscriptionService.PlaystoreSubscriptionService
	db                           *gorm.DB
	cb                           *gobreaker.CircuitBreaker
	publisher                    events.EventPublisher
}

// NewRTDNService creates a new instance of RTDNService
func NewRTDNService(ctx context.Context,
	repo repository.RTDNRepository,
	apiService apiService.PlaystoreApiService,
	playstoreSubscriptionService playstoreSubscriptionService.PlaystoreSubscriptionService,
	db *gorm.DB,
	publisher events.EventPublisher,
) RTDNService {
	service := &rtdnService{repo: repo,
		ctx:                          ctx,
		apiService:                   apiService,
		playstoreSubscriptionService: playstoreSubscriptionService,
		db:                           db,

		cb: gobreaker.NewCircuitBreaker(gobreaker.Settings{
			Name:        "PlayStoreAPI",
			MaxRequests: 5,
			Interval:    1 * time.Minute,
			Timeout:     15 * time.Second,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > 3
			},
		}),
		publisher: publisher,
	}

	return service
}

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
	event := payload.Event

	log.Println("Reached Here")
	log.Println("data = ", subscriptionData)
	log.Println("event = ", event)

	err := s.playstoreSubscriptionService.UpsertSubscription(ctx, subscriptionData, event)
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

const (
	publishTimeout         = 5 * time.Second
	maxStatusUpdateRetries = 3
)

func (s *rtdnService) ProcessWebhookEventForPublish(ctx context.Context, dtoEvent *dto.GooglePlayWebhookEvent) error {
	// Convert DTO to domain model
	domainEvent := &models.GooglePlayWebhookEvent{}
	if err := domainEvent.FromDTO(dtoEvent); err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Process within transaction
	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// 1. Save event
		if err := s.repo.Create(txCtx, domainEvent); err != nil {
			return fmt.Errorf("save event failed: %w", err)
		}

		// 2. Fetch additional data
		payload, err := s.fetchAPIData(txCtx, domainEvent)
		if err != nil {
			return fmt.Errorf("api fetch failed: %w", err)
		}

		// 3. Publish to queue
		if err := s.publishEvent(txCtx, payload); err != nil {
			return fmt.Errorf("publish failed: %w", err)
		}

		// 4. Update status
		return s.repo.UpdateStatus(txCtx, domainEvent.ID, models.StatusPublished, "")
	})

	if err != nil {
		logger.Log.Errorf("Failed to process webhook event: %v", err)
	}
	return err
}

func (s *rtdnService) fetchAPIData(ctx context.Context, event *models.GooglePlayWebhookEvent) (*models.GooglePublishPayload, error) {
	payload := &models.GooglePublishPayload{
		ID:    event.ID,
		Event: event,
	}

	logger.Log.Debugf("Processing notification type: %s", event.NotificationType)

	switch event.NotificationType {
	case "subscription":
		return s.handleSubscriptionEvent(ctx, event, payload)
	case "one_time_product":
		return s.handleOneTimeProductEvent(ctx, event, payload)
	case "voided_purchase":
		return s.handleVoidedPurchaseEvent(ctx, event, payload)
	case "test":
		return s.handleTestNotification(ctx, event, payload)
	default:
		logger.Log.Warnf("Unrecognized notification type: %s", event.NotificationType)
		return payload, nil
	}
}

func (s *rtdnService) handleSubscriptionEvent(ctx context.Context, event *models.GooglePlayWebhookEvent, payload *models.GooglePublishPayload) (*models.GooglePublishPayload, error) {
	if event.Subscription == nil {
		return nil, fmt.Errorf("missing subscription data")
	}

	data, err := s.cb.Execute(func() (interface{}, error) {
		return s.apiService.GetUserSubscriptionPurchase(
			ctx,
			event.Subscription.PurchaseToken,
			event.PackageName,
		)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get subscription data: %w", err)
	}

	subPurchase, ok := data.(*apiDto.SubscriptionPurchaseV2)
	if !ok {
		return nil, fmt.Errorf("invalid subscription data type")
	}

	payload.SubscriptionPurchase = subPurchase
	return payload, nil
}

func (s *rtdnService) handleOneTimeProductEvent(ctx context.Context, event *models.GooglePlayWebhookEvent, payload *models.GooglePublishPayload) (*models.GooglePublishPayload, error) {
	// Implement one-time product logic
	return payload, nil
}

func (s *rtdnService) handleVoidedPurchaseEvent(ctx context.Context, event *models.GooglePlayWebhookEvent, payload *models.GooglePublishPayload) (*models.GooglePublishPayload, error) {
	// Implement voided purchase logic
	return payload, nil
}

func (s *rtdnService) handleTestNotification(ctx context.Context, event *models.GooglePlayWebhookEvent, payload *models.GooglePublishPayload) (*models.GooglePublishPayload, error) {
	// Implement test notification logic
	return payload, nil
}

func (s *rtdnService) publishEvent(ctx context.Context, payload *models.GooglePublishPayload) error {
	publishCtx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()

	if err := s.publisher.PublishRTDNEvent(publishCtx, payload); err != nil {
		// Async status update with retries
		go s.safeUpdateStatusWithRetry(context.Background(), payload.ID, models.StatusFailed, err.Error())
		return fmt.Errorf("queue publish failed: %w", err)
	}
	return nil
}

func (s *rtdnService) safeUpdateStatusWithRetry(ctx context.Context, id uuid.UUID, status models.WebhookEventStatus, errorMsg string) {
	var lastErr error
	for i := 0; i < maxStatusUpdateRetries; i++ {
		err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
			return s.repo.UpdateStatus(txCtx, id, status, errorMsg)
		})

		if err == nil {
			return
		}

		lastErr = err
		time.Sleep(time.Duration(i+1) * time.Second) // Exponential backoff
	}

	logger.Log.Errorf("Failed to update status after %d retries for event %s: %v",
		maxStatusUpdateRetries, id, lastErr)
}
