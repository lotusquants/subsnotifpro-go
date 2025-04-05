// internal/playstore/dispatch/publisher.go
package dispatch

import (
	"context"
	"subsnotifpro-go/internal/constants"
	messaging "subsnotifpro-go/internal/pkg/messaging"
	"subsnotifpro-go/internal/playstore/events"
	"subsnotifpro-go/internal/playstore/rtdn/models"
)

const GooglePlayQueue = constants.RTDNQueue

type GooglePlayPublisher struct {
	publisher messaging.MessagePublisher
}

// NewGooglePlayPublisher creates a new publisher that implements EventPublisher
func NewGooglePlayPublisher(p messaging.MessagePublisher) *GooglePlayPublisher {
	return &GooglePlayPublisher{publisher: p}
}

// PublishRTDNEvent implements events.EventPublisher interface
func (p *GooglePlayPublisher) PublishRTDNEvent(ctx context.Context, event *models.GooglePublishPayload) error {
	return p.publisher.Publish(ctx, GooglePlayQueue, event)
}

// Compile-time interface implementation check
var _ events.EventPublisher = (*GooglePlayPublisher)(nil)
