package models

import (
	"encoding/json"
	"fmt"
	"time"

	"subsnotifpro-go/internal/playstore/rtdn/dto"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WebhookEventStatus string

const (
	StatusReceived   WebhookEventStatus = "RECEIVED"
	StatusPublished  WebhookEventStatus = "PUBLISHED"
	StatusProcessing WebhookEventStatus = "PROCESSING"
	StatusProcessed  WebhookEventStatus = "PROCESSED"
	StatusFailed     WebhookEventStatus = "FAILED"
	StatusRetrying   WebhookEventStatus = "RETRYING"
	StatusDeadLetter WebhookEventStatus = "DEAD_LETTER"
)

type GooglePlayWebhookEvent struct {
	// System fields
	ID          uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ReceivedAt  time.Time          `gorm:"not null;index"`
	ProcessedAt *time.Time         `gorm:"index"`
	Status      WebhookEventStatus `gorm:"type:varchar(20);default:'RECEIVED';index"`
	RetryCount  int                `gorm:"default:0"`
	Error       *string            `gorm:"type:text"`
	CreatedAt   time.Time          `gorm:"autoCreateTime;index"`
	UpdatedAt   time.Time          `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt     `gorm:"index"`

	// Payload fields
	Version          string `gorm:"not null"`
	PackageName      string `gorm:"not null;index"`
	EventTimeMillis  int64  `gorm:"not null"`
	RawPayload       string `gorm:"type:jsonb;not null"`
	NotificationType string `gorm:"type:varchar(50);index"`

	// Notification types (embedded)
	Subscription   *SubscriptionNotification   `gorm:"embedded;embeddedPrefix:sub_"`
	OneTimeProduct *OneTimeProductNotification `gorm:"embedded;embeddedPrefix:otp_"`
	VoidedPurchase *VoidedPurchaseNotification `gorm:"embedded;embeddedPrefix:void_"`
	Test           *TestNotification           `gorm:"embedded;embeddedPrefix:test_"`
}

// FromDTO converts DTO to model
func (e *GooglePlayWebhookEvent) FromDTO(dto *dto.GooglePlayWebhookEvent) error {
	raw, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("failed to marshal DTO: %w", err)
	}

	e.ID = dto.ID
	e.ReceivedAt = dto.ReceivedAt
	e.Version = dto.Version
	e.PackageName = dto.PackageName
	e.EventTimeMillis = dto.EventTimeMillis
	e.RawPayload = string(raw)
	e.NotificationType = dto.GetNotificationType()

	// Map notification types
	if dto.Subscription != nil {
		e.Subscription = &SubscriptionNotification{
			Version:          dto.Subscription.Version,
			NotificationType: SubscriptionNotificationType(dto.Subscription.NotificationType),
			PurchaseToken:    dto.Subscription.PurchaseToken,
			SubscriptionID:   dto.Subscription.SubscriptionID,
		}
	}

	if dto.OneTimeProduct != nil {
		e.OneTimeProduct = &OneTimeProductNotification{
			Version:          dto.OneTimeProduct.Version,
			NotificationType: OneTimeProductNotificationType(dto.OneTimeProduct.NotificationType),
			PurchaseToken:    dto.OneTimeProduct.PurchaseToken,
			Sku:              dto.OneTimeProduct.Sku,
		}
	}

	if dto.VoidedPurchase != nil {
		e.VoidedPurchase = &VoidedPurchaseNotification{
			PurchaseToken: dto.VoidedPurchase.PurchaseToken,
			OrderID:       dto.VoidedPurchase.OrderID,
			ProductType:   VoidedPurchaseProductType(dto.VoidedPurchase.ProductType),
			RefundType:    VoidedPurchaseRefundType(dto.VoidedPurchase.RefundType),
		}
	}

	if dto.Test != nil {
		e.Test = &TestNotification{
			Version: dto.Test.Version,
		}
	}

	return nil
}

// ToDTO converts model to DTO
func (e *GooglePlayWebhookEvent) ToDTO() (*dto.GooglePlayWebhookEvent, error) {
	var dto dto.GooglePlayWebhookEvent
	if err := json.Unmarshal([]byte(e.RawPayload), &dto); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw payload: %w", err)
	}
	return &dto, nil
}

// Status Helpers

func (e *GooglePlayWebhookEvent) MarkAsProcessing() {
	e.Status = StatusProcessing
	now := time.Now()
	e.ProcessedAt = &now
	e.Error = nil
}

func (e *GooglePlayWebhookEvent) MarkAsProcessed() {
	e.Status = StatusProcessed
	now := time.Now()
	e.ProcessedAt = &now
	e.Error = nil
}

func (e *GooglePlayWebhookEvent) MarkAsFailed(err error) {
	e.Status = StatusFailed
	now := time.Now()
	e.ProcessedAt = &now
	errMsg := err.Error()
	e.Error = &errMsg
	e.RetryCount++
}

func (e *GooglePlayWebhookEvent) MarkForRetry() {
	e.Status = StatusRetrying
	e.ProcessedAt = nil
	e.RetryCount++
}

func (e *GooglePlayWebhookEvent) MarkAsDeadLetter() {
	e.Status = StatusDeadLetter
	now := time.Now()
	e.ProcessedAt = &now
}

// Validation Helpers

func (e *GooglePlayWebhookEvent) IsTerminalState() bool {
	return e.Status == StatusProcessed || e.Status == StatusDeadLetter
}

func (e *GooglePlayWebhookEvent) CanRetry() bool {
	return e.Status == StatusFailed || e.Status == StatusRetrying
}

func (e *GooglePlayWebhookEvent) ShouldDeadLetter(maxRetries int) bool {
	return e.Status == StatusFailed && e.RetryCount >= maxRetries
}

// Database Hooks

func (e *GooglePlayWebhookEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.ReceivedAt.IsZero() {
		e.ReceivedAt = time.Now()
	}
	if e.Status == "" {
		e.Status = StatusReceived
	}
	return nil
}

func (e *GooglePlayWebhookEvent) BeforeUpdate(tx *gorm.DB) error {
	if e.Status == StatusProcessed || e.Status == StatusFailed || e.Status == StatusDeadLetter {
		now := time.Now()
		e.ProcessedAt = &now
	}
	return nil
}
