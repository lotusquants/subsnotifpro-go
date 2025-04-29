package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"subsnotifpro-go/config"
	"subsnotifpro-go/internal/pkg/contextutil"
	"subsnotifpro-go/internal/pkg/logger"
	"subsnotifpro-go/internal/pkg/messaging"
	"subsnotifpro-go/internal/subscription/events"
	"subsnotifpro-go/internal/subscription/repository"
	service "subsnotifpro-go/internal/subscription/service"

	"github.com/streadway/amqp"
)

type UnifiedSubscriptionConsumer struct {
	Consumer *messaging.Consumer
	repo     repository.SubscriptionRepository
	service  service.UnifiedSubscriptionService
}

type UnifiedConsumerOpts struct {
	Channel   *amqp.Channel
	Repo      repository.SubscriptionRepository
	Service   service.UnifiedSubscriptionService
	Publisher messaging.MessagePublisher
	Cfg       *config.RabbitMQConfig
}

func NewUnifiedConsumer(opts UnifiedConsumerOpts) *UnifiedSubscriptionConsumer {
	consumer := &UnifiedSubscriptionConsumer{
		repo:    opts.Repo,
		service: opts.Service,
	}

	consumer.Consumer = messaging.NewConsumer(
		opts.Channel,
		opts.Cfg.UnifiedSubs.Exchange,
		opts.Cfg.UnifiedSubs.RoutingKey,
		opts.Cfg.UnifiedSubs.Queue,
		opts.Cfg.UnifiedSubs.DLQ,
		opts.Cfg.MaxRetries,
		opts.Cfg.WorkerCount,
		func(ctx context.Context, payload []byte) error {
			msg, ok := contextutil.DeliveryFromContext(ctx)
			if !ok {
				return fmt.Errorf("missing message delivery in context")
			}
			return consumer.processUnifiedSubscriptionEvent(ctx, payload, msg)
		},
		opts.Publisher,
	)

	return consumer
}

func (uc *UnifiedSubscriptionConsumer) processUnifiedSubscriptionEvent(
	ctx context.Context,
	payload []byte,
	msg *amqp.Delivery,
) error {
	var event events.UnifiedEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		logger.Log.Warnf("❌ Failed to decode unified subscription event: %v", err)
		_ = msg.Nack(false, false) // Immediate DLQ on unmarshal failure
		return err
	}

	if err := uc.service.ProcessUnifiedSubscriptionEvent(ctx, event); err != nil {
		logger.Log.WithError(err).Error("Failed to process subscription event")
		return err
	}

	logger.Log.WithFields(map[string]interface{}{
		"subscription_id": event.SubscriptionID,
		"user_id":         event.UserID,
		"status":          event.Status,
	}).Info("Processed unified subscription event")

	return nil

}

func (uc *UnifiedSubscriptionConsumer) Start(ctx context.Context) {
	go uc.Consumer.Start(ctx)
}
