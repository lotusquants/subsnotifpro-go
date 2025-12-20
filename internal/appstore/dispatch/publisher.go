package dispatch

import (
	"context"
	"subsnotifpro-go/config"
	"subsnotifpro-go/internal/appstore/events"
	"subsnotifpro-go/internal/pkg/messaging"
)

type AppStorePublisher struct {
	publisher    messaging.MessagePublisher
	exchangeName string
	routingKey   string
}

// NewAppStorePublisher creates a new publisher that implements EventPublisher
func NewAppStorePublisher(
	p messaging.MessagePublisher,
	cfg *config.Config,

) *AppStorePublisher {
	return &AppStorePublisher{publisher: p,
		exchangeName: cfg.RabbitMQ.AppStore.Exchange,
		routingKey:   cfg.RabbitMQ.AppStore.RoutingKey}
}

// PublishAppStoreEvent implements events.EventPublisher interface
func (p *AppStorePublisher) PublishAppStoreEvent(ctx context.Context, event *events.AppStorePublishPayload) error {
	return p.publisher.PublishToExchange(ctx, p.exchangeName, p.routingKey, event)
}

// Compile-time interface implementation check
var _ events.EventPublisher = (*AppStorePublisher)(nil)
