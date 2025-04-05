package events

import (
	"context"
	"subsnotifpro-go/internal/playstore/rtdn/models"
)

type EventPublisher interface {
	PublishRTDNEvent(ctx context.Context, event *models.GooglePublishPayload) error
}

type EventProcessor interface {
	ProcessWebhookEvent(ctx context.Context, payload models.GooglePublishPayload) error
}
