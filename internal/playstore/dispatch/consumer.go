package dispatch

import (
	"context"
	"encoding/json"
	"fmt"
	"subsnotifpro-go/internal/constants"
	"subsnotifpro-go/internal/logger"
	"subsnotifpro-go/internal/pkg/contextutil"
	messaging "subsnotifpro-go/internal/pkg/messaging"
	"subsnotifpro-go/internal/playstore/rtdn/models"
	"subsnotifpro-go/internal/playstore/rtdn/repository"
	"subsnotifpro-go/internal/playstore/rtdn/service"

	// "subsnotifpro-go/internal/playstore/rtdn/service"

	"github.com/streadway/amqp"
)

type GooglePlayConsumer struct {
	Consumer *messaging.Consumer
	repo     repository.RTDNRepository
	service  service.RTDNService
}

func NewGooglePlayConsumer(
	ch *amqp.Channel,
	repo repository.RTDNRepository,
	svc service.RTDNService,
	publisher messaging.MessagePublisher,
) *GooglePlayConsumer {
	return &GooglePlayConsumer{
		Consumer: messaging.NewConsumer(
			ch,
			constants.RTDNQueue,
			constants.RTDNDLQ,
			constants.MaxRetries,
			constants.WorkerCount,
			func(ctx context.Context, payload []byte) error {
				// Extract delivery from context
				msg, ok := contextutil.DeliveryFromContext(ctx)
				if !ok {
					return fmt.Errorf("missing message delivery in context")
				}
				return processMessage(ctx, payload, msg, repo, svc)
			},
			publisher,
		),
		repo:    repo,
		service: svc,
	}
}

// processMessage handles domain logic WITH transaction-safe acknowledgment
func processMessage(
	ctx context.Context,
	payload []byte,
	msg *amqp.Delivery,
	repo repository.RTDNRepository,
	svc service.RTDNService,
) error {
	var publishPayLoad models.GooglePublishPayload
	if err := json.Unmarshal(payload, &publishPayLoad); err != nil {
		logger.Log.Warnf("❌ Failed to decode event: %v", err)
		_ = msg.Nack(false, false) // Immediate DLQ on unmarshal failure
		return err
	}

	return repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get current retry count from headers
		retryCount := messaging.GetRetryCount(msg.Headers)

		// 1. Verify event exists first
		exists, err := repo.Exists(txCtx, publishPayLoad.ID)
		if err != nil {
			_ = msg.Nack(false, true)
			return fmt.Errorf("failed to check event existence: %w", err)
		}
		if !exists {
			_ = msg.Nack(false, false) // Move to DLQ if event doesn't exist
			return fmt.Errorf("event not found: %s", publishPayLoad.ID)
		}

		// 1. Update status (different status for retries)
		status := models.StatusProcessing
		if retryCount > 0 {
			status = models.StatusRetrying
			if err := repo.IncrementRetryCount(txCtx, publishPayLoad.ID); err != nil {
				_ = msg.Nack(false, true)
				return fmt.Errorf("failed to increment retry count: %w", err)
			}
		}

		if err := repo.UpdateStatus(txCtx, publishPayLoad.ID, status, ""); err != nil {
			_ = msg.Nack(false, true)
			return err
		}

		// 2. Process event
		if err := svc.ProcessWebhookEvent(txCtx, publishPayLoad); err != nil {
			_ = msg.Nack(false, true) // Requeue on processing failure
			return err
		}

		// 3. Mark as processed
		if err := repo.UpdateStatus(txCtx, publishPayLoad.ID, models.StatusProcessed, ""); err != nil {
			_ = msg.Nack(false, true) // Requeue on status update failure
			return err
		}

		// metrics.ProcessedEvents.Inc()
		return nil
	})
}
