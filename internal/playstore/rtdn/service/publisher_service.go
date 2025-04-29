package service

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/pkg/logger"
	apiDto "subsnotifpro-go/internal/playstore/api/dto"
	"subsnotifpro-go/internal/playstore/rtdn/dto"
	"subsnotifpro-go/internal/playstore/rtdn/models"
	"time"

	"github.com/google/uuid"
)

func (s *rtdnService) ProcessWebhookEventForPublish(ctx context.Context, dtoEvent *dto.GooglePlayWebhookEvent) error {
	// Convert DTO to domain model
	domainEvent := &models.GooglePlayWebhookEvent{}
	if err := domainEvent.FromDTO(dtoEvent); err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	var payload *models.GooglePublishPayload

	// 1. Do all database work in transaction
	err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Save event
		if err := s.repo.Create(txCtx, domainEvent); err != nil {
			return fmt.Errorf("save event failed: %w", err)
		}

		// Fetch additional data (if needed for DB operations)
		var err error
		payload, err = s.fetchAPIData(txCtx, domainEvent)
		if err != nil {
			return fmt.Errorf("api fetch failed: %w", err)
		}

		// Update status
		return s.repo.UpdateStatus(txCtx, domainEvent.ID, models.StatusProcessed, "")
	})

	if err != nil {
		logger.Log.Errorf("Transaction failed: %v", err)
		return err
	}

	// 2. Only after successful commit, publish the message
	if err := s.publishEvent(ctx, payload); err != nil {
		// If publish fails, update status to indicate failure
		updateErr := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
			return s.repo.UpdateStatus(txCtx, domainEvent.ID, models.StatusFailed, err.Error())
		})
		if updateErr != nil {
			logger.Log.Errorf("Failed to update status after publish failure: %v", updateErr)
		}
		return fmt.Errorf("publish failed: %w", err)
	}

	// 3. Update status to published (optional)
	if err := s.repo.WithTransaction(ctx, func(txCtx context.Context) error {
		return s.repo.UpdateStatus(txCtx, domainEvent.ID, models.StatusPublished, "")
	}); err != nil {
		logger.Log.Errorf("Failed to update status to published: %v", err)
	}

	return nil
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
