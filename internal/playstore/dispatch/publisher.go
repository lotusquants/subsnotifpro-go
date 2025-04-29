// internal/playstore/dispatch/publisher.go
package dispatch

import (
	"context"
	"subsnotifpro-go/config"
	messaging "subsnotifpro-go/internal/pkg/messaging"
	"subsnotifpro-go/internal/playstore/events"
	"subsnotifpro-go/internal/playstore/rtdn/models"
)

type GooglePlayPublisher struct {
	publisher    messaging.MessagePublisher
	exchangeName string
	routingKey   string
}

// NewGooglePlayPublisher creates a new publisher that implements EventPublisher
func NewGooglePlayPublisher(
	p messaging.MessagePublisher,
	cfg *config.Config,

) *GooglePlayPublisher {
	return &GooglePlayPublisher{publisher: p,
		exchangeName: cfg.RabbitMQ.RTDN.Exchange,
		routingKey:   cfg.RabbitMQ.RTDN.RoutingKey}
}

// PublishRTDNEvent implements events.EventPublisher interface
func (p *GooglePlayPublisher) PublishRTDNEvent(ctx context.Context, event *models.GooglePublishPayload) error {
	return p.publisher.PublishToExchange(ctx, p.exchangeName, p.routingKey, event)
}

// Compile-time interface implementation check
var _ events.EventPublisher = (*GooglePlayPublisher)(nil)
