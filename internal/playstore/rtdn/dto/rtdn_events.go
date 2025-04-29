package dto

import (
	"log"
	"time"

	"github.com/google/uuid"
)

// type GooglePlayWebhookEventStatus string

// const (
// 	GooglePlayWebhookStatusReceived   GooglePlayWebhookEventStatus = "RECEIVED"
// 	GooglePlayWebhookStatusProcessing GooglePlayWebhookEventStatus = "PROCESSING"
// 	GooglePlayWebhookStatusProcessed  GooglePlayWebhookEventStatus = "PROCESSED"
// 	GooglePlayWebhookStatusFailed     GooglePlayWebhookEventStatus = "FAILED"
// 	GooglePlayWebhookStatusRetrying   GooglePlayWebhookEventStatus = "RETRYING"
// 	GooglePlayWebhookStatusDeadLetter GooglePlayWebhookEventStatus = "DEAD_LETTER"
// )

// dto
type GooglePlayWebhookEvent struct {
	// System-generated fields
	ID         uuid.UUID `json:"Id" validate:"required"` // Primary identifier
	ReceivedAt time.Time `json:"receivedAt" validate:"required"`

	// Payload fields
	Version         string `json:"version" validate:"required"`
	PackageName     string `json:"packageName" validate:"required"`
	EventTimeMillis int64  `json:"eventTimeMillis" validate:"required"`
	RawPayload      string `json:"rawPayload" validate:"required"`

	// Notification types
	Subscription   *SubscriptionNotification   `json:"subscriptionNotification,omitempty"`
	OneTimeProduct *OneTimeProductNotification `json:"oneTimeProductNotification,omitempty"`
	VoidedPurchase *VoidedPurchaseNotification `json:"voidedPurchaseNotification,omitempty"`
	Test           *TestNotification           `json:"testNotification,omitempty"`
}

// Init ensures required fields are set
func (e *GooglePlayWebhookEvent) Init() {
	if e.ID == uuid.Nil {

		// Generate UUIDv7
		uuidV7, err := uuid.NewV7()
		if err != nil {
			log.Println("failed to generate UUIDv7: %w", err)
			return
		}
		e.ID = uuidV7
	}
	if e.ReceivedAt.IsZero() {
		e.ReceivedAt = time.Now().UTC()
	}
}

// GetNotificationType returns the event type as string
func (e *GooglePlayWebhookEvent) GetNotificationType() string {
	switch {
	case e.Subscription != nil:
		return "subscription"
	case e.OneTimeProduct != nil:
		return "one_time_product"
	case e.VoidedPurchase != nil:
		return "voided_purchase"
	case e.Test != nil:
		return "test"
	default:
		return "unknown"
	}
}
