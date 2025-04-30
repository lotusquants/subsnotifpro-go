package service

import (
	"context"
	"fmt"
	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/subscription/models"

	playstoreSubsModels "subsnotifpro-go/internal/playstore/subscription/models"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

// Event type conversion helpers
func convertPlaystoreEventType(eventType string) string {
	switch eventType {
	case "SUBSCRIPTION_PURCHASED":
		return "PURCHASE"
	case "SUBSCRIPTION_RENEWED":
		return "RENEWAL"
	case "SUBSCRIPTION_CANCELED":
		return "CANCELLATION"
	case "SUBSCRIPTION_RESTARTED":
		return "RESTART"
	default:
		return eventType
	}
}

func (s *unifiedSubscriptionService) GetPlaystoreSubscriptionEvents(
	ctx context.Context,
	subscriptionID string,
	page, pageSize int,
) ([]models.UnifiedSubscriptionEvent, int, error) {
	// Log entry with context
	logger.Log.WithFields(map[string]interface{}{
		"subscription_id": subscriptionID,
		"page":            page,
		"page_size":       pageSize,
	}).Info("Fetching Playstore subscription events")

	// Convert subscriptionID to UUID
	subUUID, err := uuid.Parse(subscriptionID)
	if err != nil {
		logger.Log.WithError(err).WithFields(map[string]interface{}{
			"subscription_id": subscriptionID,
		}).Error("Invalid subscription ID format")
		return nil, 0, fmt.Errorf("invalid subscription ID format: %w", err)
	}

	// 1. Get base events with pagination
	baseEvents, total, err := s.playstoreEventRepo.GetBaseEvents(ctx, subUUID, page, pageSize)
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

	// 2. Collect all event IDs for batch queries
	eventIDs := make([]uuid.UUID, len(baseEvents))
	for i, event := range baseEvents {
		eventIDs[i] = event.ID
	}

	// 3. Parallel fetch of all related data
	var (
		lineItemHistories  []playstoreSubsModels.SubscriptionLineItemHistory
		autoRenewHistories []playstoreSubsModels.AutoRenewingPlanHistory
		offerHistories     []playstoreSubsModels.OfferDetailsHistory
	)

	errGroup, ctx := errgroup.WithContext(ctx)

	errGroup.Go(func() error {
		var err error
		lineItemHistories, err = s.playstoreEventRepo.GetLineItemHistories(ctx, eventIDs)
		if err != nil {
			logger.Log.WithError(err).WithFields(map[string]interface{}{
				"subscription_id": subscriptionID,
				"event_ids":       eventIDs,
			}).Error("Failed to get line item histories")
		}
		return err
	})

	errGroup.Go(func() error {
		var err error
		autoRenewHistories, err = s.playstoreEventRepo.GetAutoRenewHistories(ctx, eventIDs)
		if err != nil {
			logger.Log.WithError(err).WithFields(map[string]interface{}{
				"subscription_id": subscriptionID,
				"event_ids":       eventIDs,
			}).Error("Failed to get auto renew histories")
		}
		return err
	})

	errGroup.Go(func() error {
		var err error
		offerHistories, err = s.playstoreEventRepo.GetOfferHistories(ctx, eventIDs)
		if err != nil {
			logger.Log.WithError(err).WithFields(map[string]interface{}{
				"subscription_id": subscriptionID,
				"event_ids":       eventIDs,
			}).Error("Failed to get offer histories")
		}
		return err
	})

	if err := errGroup.Wait(); err != nil {
		return nil, 0, fmt.Errorf("failed to fetch event details: %w", err)
	}

	logger.Log.WithFields(map[string]interface{}{
		"subscription_id":       subscriptionID,
		"line_item_histories":   len(lineItemHistories),
		"auto_renew_histories":  len(autoRenewHistories),
		"offer_details_history": len(offerHistories),
	}).Debug("Retrieved related event histories")

	// 4. Build lookup maps for efficient data access
	lineItemMap := buildLineItemHistoryMap(lineItemHistories)
	autoRenewMap := buildAutoRenewHistoryMap(autoRenewHistories)
	offerMap := buildOfferHistoryMap(offerHistories)

	// 5. Compose unified events
	unifiedEvents := make([]models.UnifiedSubscriptionEvent, 0, len(baseEvents))
	for _, baseEvent := range baseEvents {
		event := models.UnifiedSubscriptionEvent{
			ID:             baseEvent.ID.String(),
			SubscriptionID: subscriptionID,
			EventType:      convertPlaystoreEventType(baseEvent.EventType),
			Timestamp:      baseEvent.CreatedAt,

			Platform:  models.PlatformGoogle,
			CreatedAt: baseEvent.CreatedAt,
		}

		// Enrich with line item history if available
		if li, ok := lineItemMap[baseEvent.ID]; ok {

			event.ProductID = li.ProductID

		}

		// Enrich with auto-renew history if available
		if ar, ok := autoRenewMap[baseEvent.ID]; ok {
			if ar.CurrentPrice != nil {
				event.Amount = float64(ar.CurrentPrice.Units) / 1000000
				event.Currency = ar.CurrentPrice.CurrencyCode
			}

		}

		// Enrich with offer details if available
		if offer, ok := offerMap[baseEvent.ID]; ok {
			if offer.CurrentPhasePrice.Units != 0 || offer.CurrentPhasePrice.CurrencyCode != "" {
				event.Amount = float64(offer.CurrentPhasePrice.Units) / 1000000
				event.Currency = offer.CurrentPhasePrice.CurrencyCode
			}

			event.BasePlanID = &offer.BasePlanID
			if offer.OfferID != nil {
				event.ActiveOfferID = offer.OfferID
			}
		}

		unifiedEvents = append(unifiedEvents, event)
	}

	logger.Log.WithFields(map[string]interface{}{
		"subscription_id": subscriptionID,
		"event_count":     len(unifiedEvents),
	}).Info("Successfully processed Playstore subscription events")

	return unifiedEvents, int(total), nil
}

// Helper functions to build lookup maps
func buildLineItemHistoryMap(histories []playstoreSubsModels.SubscriptionLineItemHistory) map[uuid.UUID]playstoreSubsModels.SubscriptionLineItemHistory {
	m := make(map[uuid.UUID]playstoreSubsModels.SubscriptionLineItemHistory)
	for _, h := range histories {
		m[h.ChangeEventID] = h
	}
	return m
}

func buildAutoRenewHistoryMap(histories []playstoreSubsModels.AutoRenewingPlanHistory) map[uuid.UUID]playstoreSubsModels.AutoRenewingPlanHistory {
	m := make(map[uuid.UUID]playstoreSubsModels.AutoRenewingPlanHistory)
	for _, h := range histories {
		m[h.ChangeEventID] = h
	}
	return m
}

func buildOfferHistoryMap(histories []playstoreSubsModels.OfferDetailsHistory) map[uuid.UUID]playstoreSubsModels.OfferDetailsHistory {
	m := make(map[uuid.UUID]playstoreSubsModels.OfferDetailsHistory)
	for _, h := range histories {
		m[h.ChangeEventID] = h
	}
	return m
}
