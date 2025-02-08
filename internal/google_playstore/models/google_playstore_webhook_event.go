// internal/google_playstore/models/google_playstore_webhook_event.go
package models

import (
	"time"

	"gorm.io/gorm"
)

// GooglePlayWebhookEvent represents the top-level RTDN webhook payload
type GooglePlayWebhookEvent struct {
	ID                         string                      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Version                    string                      `gorm:"type:varchar(10);not null"`
	PackageName                string                      `gorm:"type:varchar(255);not null"`
	EventTimeMillis            int64                       `gorm:"not null"`
	SubscriptionNotification   *SubscriptionNotification   `gorm:"embedded;embeddedPrefix:sub_"`
	OneTimeProductNotification *OneTimeProductNotification `gorm:"embedded;embeddedPrefix:one_"`
	VoidedPurchaseNotification *VoidedPurchaseNotification `gorm:"embedded;embeddedPrefix:void_"`
	TestNotification           *TestNotification           `gorm:"embedded;embeddedPrefix:test_"`
	RawPayload                 string                      `gorm:"type:jsonb;not null"`
	Status                     string                      `gorm:"type:varchar(50);default:'pending';index"`
	RetryCount                 int                         `gorm:"default:0"`
	CreatedAt                  time.Time                   `gorm:"autoCreateTime;index"`
	UpdatedAt                  time.Time                   `gorm:"autoUpdateTime"`
	DeletedAt                  gorm.DeletedAt              `gorm:"index"`
}

// SubscriptionNotification contains subscription-specific details
type SubscriptionNotification struct {
	Version          string `gorm:"type:varchar(10)"`
	NotificationType int    `gorm:"not null"`
	PurchaseToken    string `gorm:"type:varchar(255);not null"`
	SubscriptionID   string `gorm:"type:varchar(255);not null"`
}

// OneTimeProductNotification contains one-time purchase details
type OneTimeProductNotification struct {
	Version          string `gorm:"type:varchar(10)"`
	NotificationType int    `gorm:"not null"`
	PurchaseToken    string `gorm:"type:varchar(255);not null"`
	Sku              string `gorm:"type:varchar(255);not null"`
}

// VoidedPurchaseNotification contains voided purchase details
type VoidedPurchaseNotification struct {
	PurchaseToken string `gorm:"type:varchar(255);not null"`
	OrderID       string `gorm:"type:varchar(255);not null"`
	ProductType   int    `gorm:"not null"` // 1 = Subscription, 2 = One-time purchase
	RefundType    int    `gorm:"not null"` // 1 = Full refund, 2 = Partial refund
}

// TestNotification represents a test notification from Google Play Console
type TestNotification struct {
	Version string `gorm:"type:varchar(10)"`
}
