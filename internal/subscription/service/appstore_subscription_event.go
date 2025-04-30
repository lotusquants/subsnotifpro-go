package service

import (
	"context"
	"fmt"

	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/subscription/models"

	"github.com/google/uuid"
)

func (s *unifiedSubscriptionService) GetAppStoreSubscriptionEvents(
	ctx context.Context,
	subscriptionID string,
	page, pageSize int,
) ([]models.UnifiedSubscriptionEvent, int, error) {
	// Log entry with context
	logger.Log.WithFields(map[string]interface{}{
		"subscription_id": subscriptionID,
		"page":            page,
		"page_size":       pageSize,
	}).Info("Fetching App Store subscription events")

	// Convert subscriptionID to UUID
	subUUID, err := uuid.Parse(subscriptionID)
	if err != nil {
		logger.Log.WithError(err).WithFields(map[string]interface{}{
			"subscription_id": subscriptionID,
		}).Error("Invalid subscription ID format")
		return nil, 0, fmt.Errorf("invalid subscription ID format: %w", err)
	}

	// 1. Get base events with pagination
	baseEvents, total, err := s.appstoreEventRepo.GetBaseEvents(ctx, subUUID, page, pageSize)
	if err != nil {
		logger.Log.WithError(err).WithFields(map[string]interface{}{
			"subscription_id": subscriptionID,
		}).Error("Failed to get base events")
		return nil, 0, fmt.Errorf("failed to get base events: %w", err)
	}

	logger.Log.WithFields(map[string]interface{}{
		"subscription_id": subscriptionID,
		"event_count":     len(baseEvents),
		"total_events":    total,
	}).Debug("Retrieved base events")

	// 2. Compose unified events
	unifiedEvents := make([]models.UnifiedSubscriptionEvent, 0, len(baseEvents))
	for _, baseEvent := range baseEvents {
		event := models.UnifiedSubscriptionEvent{
			ID:             baseEvent.ID.String(),
			SubscriptionID: subscriptionID,
			EventType:      string(baseEvent.Type),
			Timestamp:      baseEvent.EventDate,

			Platform:  models.PlatformApple,
			CreatedAt: baseEvent.CreatedAt,
		}

		// Enrich with pricing information if available
		if baseEvent.Price != nil && baseEvent.Currency != nil {
			event.Amount = float64(*baseEvent.Price) / 1000 // Convert milliunits to standard currency
			event.Currency = *baseEvent.Currency
		}

		// Enrich with product information
		if baseEvent.ProductID != nil {
			event.ProductID = *baseEvent.ProductID
		}

		// Enrich with offer details if available
		if baseEvent.OfferIdentifier != nil {
			event.ActiveOfferID = baseEvent.OfferIdentifier
		}

		unifiedEvents = append(unifiedEvents, event)
	}

	logger.Log.WithFields(map[string]interface{}{
		"subscription_id": subscriptionID,
		"event_count":     len(unifiedEvents),
	}).Info("Successfully processed App Store subscription events")

	return unifiedEvents, int(total), nil
}
