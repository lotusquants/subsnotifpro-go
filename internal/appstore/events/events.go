package events

import (
	"context"

	"github.com/google/uuid"
)

type AppStorePublishPayload struct {
	ID uuid.UUID `json:"id"`
}

type EventPublisher interface {
	PublishAppStoreEvent(ctx context.Context, event *AppStorePublishPayload) error
}

type EventProcessor interface {
	ProcessAppStoreEvent(ctx context.Context, id string) error
}
