// internal/appstore/dispatch/consumer.go
package dispatch

import (
	"context"
	"encoding/json"
	"fmt"
	"subsnotifpro-go/config"
	"subsnotifpro-go/internal/appstore/events"
	"subsnotifpro-go/internal/appstore/webhooks/models"
	"subsnotifpro-go/internal/appstore/webhooks/repository"
	"subsnotifpro-go/internal/appstore/webhooks/service"
	"subsnotifpro-go/internal/pkg/contextutil"
	"subsnotifpro-go/internal/pkg/logger"
	messaging "subsnotifpro-go/internal/pkg/messaging"

	"github.com/streadway/amqp"
)

type AppStoreConsumer struct {
	Consumer *messaging.Consumer
	repo     repository.AppstoreNotificationsRepository
	service  service.AppStoreNotificationsService
}

func NewAppStoreConsumer(
	ch *amqp.Channel,
	repo repository.AppstoreNotificationsRepository,
	svc service.AppStoreNotificationsService,
	publisher messaging.MessagePublisher,
	cfg *config.Config,
) *AppStoreConsumer {
	return &AppStoreConsumer{
		Consumer: messaging.NewConsumer(
			ch,
			cfg.RabbitMQ.AppStore.Exchange,
			cfg.RabbitMQ.AppStore.RoutingKey,
			cfg.RabbitMQ.AppStore.Queue,
			cfg.RabbitMQ.AppStore.DLQ,
			cfg.RabbitMQ.MaxRetries,
			cfg.RabbitMQ.WorkerCount,
			func(ctx context.Context, payload []byte) error {
				// Extract delivery from context
				msg, ok := contextutil.DeliveryFromContext(ctx)
				if !ok {
					return fmt.Errorf("missing message delivery in context")
				}
				return processAppStoreMessage(ctx, payload, msg, repo, svc)
			},
			publisher,
		),
		repo:    repo,
		service: svc,
	}
}

func processAppStoreMessage(
	ctx context.Context,
	payload []byte,
	msg *amqp.Delivery,
	repo repository.AppstoreNotificationsRepository,
	svc service.AppStoreNotificationsService,
) error {
	var publishPayload events.AppStorePublishPayload
	if err := json.Unmarshal(payload, &publishPayload); err != nil {
		logger.Log.Warnf("❌ Failed to decode App Store event: %v", err)
		_ = msg.Nack(false, false) // Immediate DLQ on unmarshal failure
		return err
	}

	return repo.WithTransaction(ctx, func(txCtx context.Context) error {
		// Get current retry count from headers
		retryCount := messaging.GetRetryCount(msg.Headers)

		// 1. Verify notification exists
		exists, err := repo.Exists(txCtx, publishPayload.ID)
		if err != nil {
			_ = msg.Nack(false, true)
			return fmt.Errorf("failed to check notification existence: %w", err)
		}
		if !exists {
			_ = msg.Nack(false, false) // Move to DLQ if notification doesn't exist
			return fmt.Errorf("notification not found: %s", publishPayload.ID)
		}

		// 2. Update status (different status for retries)
		status := models.StatusProcessing
		if retryCount > 0 {
			status = models.StatusRetrying
			if err := repo.IncrementRetryCount(txCtx, publishPayload.ID); err != nil {
				_ = msg.Nack(false, true)
				return fmt.Errorf("failed to increment retry count: %w", err)
			}
		}

		if err := repo.UpdateStatus(txCtx, publishPayload.ID, status, ""); err != nil {
			_ = msg.Nack(false, true)
			return err
		}

		// 3. Process event
		if err := svc.ProcessAppStoreEvent(txCtx, publishPayload.ID); err != nil {
			_ = msg.Nack(false, true) // Requeue on processing failure
			return err
		}

		// 4. Mark as processed
		if err := repo.UpdateStatus(txCtx, publishPayload.ID, models.StatusProcessed, ""); err != nil {
			_ = msg.Nack(false, true) // Requeue on status update failure
			return err
		}

		return nil
	})
}
